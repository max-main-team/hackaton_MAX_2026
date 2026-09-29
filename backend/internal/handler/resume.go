package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
	"max-miniapp/backend/internal/scoring"
)

const (
	parseTextMaxLen = 30_000
	parseTextMinLen = 50
)

type ResumeHandler struct {
	resumes *repository.ResumeRepo
	ai      *scoring.AIClient
	log     *slog.Logger
}

func NewResumeHandler(resumes *repository.ResumeRepo, ai *scoring.AIClient, log *slog.Logger) *ResumeHandler {
	return &ResumeHandler{resumes: resumes, ai: ai, log: log}
}

// Get возвращает резюме текущего кандидата.
//
//	@Summary     Моё резюме
//	@Tags        resume
//	@Produce     json
//	@Success     200 {object} dto.Resume
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     404 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/resume [get]
func (h *ResumeHandler) Get(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	resume, err := h.resumes.GetByUserID(c.Request().Context(), userID)
	if errors.Is(err, repository.ErrResumeNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "resume not found")
	}
	if err != nil {
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromResume(resume))
}

// Put создаёт или обновляет резюме (upsert + версия в историю).
//
//	@Summary     Создать/обновить резюме
//	@Description При каждом сохранении пишется снапшот в resume_versions.
//	@Description Ручное редактирование не затирает распарсенный source_text.
//	@Tags        resume
//	@Accept      json
//	@Produce     json
//	@Param       request body dto.ResumeInput true "резюме"
//	@Success     200 {object} dto.Resume
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/resume [put]
func (h *ResumeHandler) Put(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var in dto.ResumeInput
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := dto.ValidateResumeInput(in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if len(in.Links) == 0 {
		in.Links = []byte("[]")
	}

	source, sourceText := "manual", ""
	existing, err := h.resumes.GetByUserID(c.Request().Context(), userID)
	switch {
	case errors.Is(err, repository.ErrResumeNotFound):
		// первое сохранение
	case err != nil:
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	default:
		// ручная правка не должна терять распарсенный из файла текст
		if existing.Source == "file_parse" {
			source = existing.Source
			sourceText = existing.SourceText
		}
	}

	saved, err := h.resumes.UpsertResume(c.Request().Context(), repository.Resume{
		UserID:           userID,
		Title:            in.Title,
		Skills:           in.Skills,
		ExperienceMonths: in.ExperienceMonths,
		About:            in.About,
		Education:        in.Education,
		Links:            string(in.Links),
		City:             in.City,
		WorkFormat:       in.WorkFormat,
		EmploymentType:   in.EmploymentType,
		SalaryMin:        in.SalaryMin,
		SalaryMax:        in.SalaryMax,
		Source:           source,
		SourceText:       sourceText,
		IsActive:         true,
	})
	if err != nil {
		h.log.Error("upsert resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.FromResume(saved))
}

// ConfirmActivity — ответ на еженедельный опрос «подбор актуален?».
//
//	@Summary     Подтвердить актуальность подбора
//	@Tags        resume
//	@Accept      json
//	@Produce     json
//	@Param       request body dto.ConfirmActivityRequest true "active"
//	@Success     200 {object} dto.Resume
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     404 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/resume/confirm-activity [post]
func (h *ResumeHandler) ConfirmActivity(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.ConfirmActivityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.resumes.SetConfirmActivity(c.Request().Context(), userID, req.Active); err != nil {
		h.log.Error("confirm activity failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	resume, err := h.resumes.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusNotFound, "resume not found")
	}

	return c.JSON(http.StatusOK, dto.FromResume(resume))
}

// Parse структурирует текст резюме (извлечённый на фронте из PDF) в черновик через AI.
//
//	@Summary     Распарсить текст резюме
//	@Description Принимает текст резюме, AI извлекает поля по нашей схеме.
//	@Description Создаёт черновик (source=file_parse, is_active=false, source_text сохранён).
//	@Description Если AI недоступен — черновик только с текстом, поля пустые.
//	@Tags        resume
//	@Accept      json
//	@Produce     json
//	@Param       request body dto.ParseResumeRequest true "текст резюме"
//	@Success     200 {object} dto.ParseResumeResponse
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/resume/parse [post]
func (h *ResumeHandler) Parse(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var in dto.ParseResumeRequest
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "text is required")
	}
	if len(text) > parseTextMaxLen {
		text = text[:parseTextMaxLen]
	}

	ctx := c.Request().Context()
	draft, extractErr := h.ai.ExtractResume(ctx, text)
	notes := draft.Notes
	if extractErr != nil {
		h.log.Error("ai extract resume failed", slog.Any("err", err))
		notes = "Не удалось автоматически распознать текст — заполните поля вручную."
	}

	existing, err := h.resumes.GetByUserID(ctx, userID)
	switch {
	case errors.Is(err, repository.ErrResumeNotFound):
		existing = repository.Resume{IsActive: false}
	case err != nil:
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	if extractErr != nil || (strings.TrimSpace(draft.Title) == "" && strings.TrimSpace(draft.Skills) == "") {
		// Не затираем резюме пользователя пустым черновиком.
		if existing.ID != 0 {
			return c.JSON(http.StatusOK, dto.ParseResumeResponse{
				Resume:    dto.FromResume(existing),
				AIComment: notes,
			})
		}
		return c.JSON(http.StatusOK, dto.ParseResumeResponse{
			Resume:    dto.FromResume(repository.Resume{IsActive: false}),
			AIComment: notes,
		})
	}

	resumeInput := repository.Resume{
		UserID:           userID,
		Title:            draft.Title,
		Skills:           draft.Skills,
		ExperienceMonths: draft.ExperienceMonths,
		About:            draft.About,
		Education:        draft.Education,
		Links:            "[]",
		City:             draft.City,
		WorkFormat:       draft.WorkFormat,
		EmploymentType:   draft.EmploymentType,
		SalaryMin:        draft.SalaryMin,
		SalaryMax:        draft.SalaryMax,
		Source:           "file_parse",
		SourceText:       text,
		IsActive:         existing.IsActive,
	}
	if existing.ID == 0 {
		resumeInput.IsActive = false
	}

	saved, err := h.resumes.UpsertResume(ctx, resumeInput)
	if err != nil {
		h.log.Error("upsert parsed resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.ParseResumeResponse{
		Resume:    dto.FromResume(saved),
		AIComment: notes,
	})
}
