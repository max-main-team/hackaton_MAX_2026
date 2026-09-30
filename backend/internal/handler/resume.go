package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	ledongthucpdf "github.com/ledongthuc/pdf"

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
	text := scoring.NormalizeResumeText(in.Text)
	if text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "text is required")
	}
	text = scoring.TruncateText(text, parseTextMaxLen)

	ctx := c.Request().Context()
	existing, err := h.resumes.GetByUserID(ctx, userID)
	switch {
	case errors.Is(err, repository.ErrResumeNotFound):
		existing = repository.Resume{IsActive: false, Source: "file_parse"}
	case err != nil:
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	draft := existing
	draft.UserID = userID
	draft.Source = "file_parse"
	draft.SourceText = text
	draft.ParseStatus = "processing"
	if draft.Links == "" {
		draft.Links = "[]"
	}
	if existing.ID == 0 {
		draft.IsActive = false
	}

	saved, err := h.resumes.UpsertResume(ctx, draft)
	if err != nil {
		h.log.Error("upsert parsed resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	// AI-распознавание уходит в фон: вебвью MAX обрывает запросы длиннее
	// ~60 секунд (URLSession), ждать внутри HTTP-запроса нельзя.
	go h.processParsedResume(userID, text, saved, existing)

	return c.JSON(http.StatusAccepted, dto.ParseStartResponse{Status: "processing"})
}

// processParsedResume распознаёт текст резюме через AI и обновляет
// черновик. Статус: processing → done | failed (виден фронту).
func (h *ResumeHandler) processParsedResume(userID int64, text string, saved repository.Resume, existing repository.Resume) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("parse worker panic", slog.Any("panic", r))
			_ = h.resumes.SetParseStatus(context.Background(), userID, "failed")
		}
	}()

	bgCtx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	aiDraft, extractErr := h.ai.ExtractResume(bgCtx, text)
	if extractErr != nil {
		h.log.Error("ai extract resume failed", slog.Any("err", extractErr))
		_ = h.resumes.SetParseStatus(bgCtx, userID, "failed")
		return
	}
	if strings.TrimSpace(aiDraft.Title) == "" && strings.TrimSpace(aiDraft.Skills) == "" {
		h.log.Warn("ai extract returned empty draft")
		_ = h.resumes.SetParseStatus(bgCtx, userID, "failed")
		return
	}

	if _, err := h.resumes.UpsertResume(bgCtx, repository.Resume{
		UserID:           userID,
		Title:            aiDraft.Title,
		Skills:           aiDraft.Skills,
		ExperienceMonths: aiDraft.ExperienceMonths,
		About:            aiDraft.About,
		Education:        aiDraft.Education,
		Links:            "[]",
		City:             aiDraft.City,
		WorkFormat:       aiDraft.WorkFormat,
		EmploymentType:   aiDraft.EmploymentType,
		SalaryMin:        aiDraft.SalaryMin,
		SalaryMax:        aiDraft.SalaryMax,
		Source:           "file_parse",
		SourceText:       text,
		ParseStatus:      "done",
		IsActive:         true, // успешный парсинг = юзер хочет быть виден компаниям
	}); err != nil {
		h.log.Error("upsert parsed resume failed", slog.Any("err", err))
		_ = h.resumes.SetParseStatus(bgCtx, userID, "failed")
	}
}

// extractPDFText сервер-сайд извлечение текста из PDF (Go, без браузера).
func extractPDFText(path string) (string, error) {
	_, r, err := ledongthucpdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	var b strings.Builder
	total := 0
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text = scoring.NormalizeResumeText(text)
		b.WriteString(text)
		b.WriteString("\n\n")
		total += len(text)
		if total > parseTextMaxLen {
			break
		}
	}
	return b.String(), nil
}

// ParseFile — загрузка PDF-файла и фоновый AI-парсинг (текст извлекается
// на сервере, вебвью не участвует).
//
//	@Summary     Загрузить PDF-резюме
//	@Tags        resume
//	@Accept      multipart/form-data
//	@Produce     json
//	@Param       file formData file true "PDF-файл резюме (до 10 МБ)"
//	@Success     202 {object} dto.ParseStartResponse
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     422 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/my/resume/parse-file [post]
func (h *ResumeHandler) ParseFile(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}
	const maxUploadBytes = 25 * 1024 * 1024
	if fileHeader.Size > maxUploadBytes {
		return echo.NewHTTPError(http.StatusBadRequest, "file is larger than 25 MB — compress or export a smaller PDF")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot read file")
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return echo.NewHTTPError(http.StatusBadRequest, "cannot read file")
	}
	tmp.Close()

	head := make([]byte, 5)
	if f, err := os.Open(tmpPath); err == nil {
		_, _ = f.Read(head)
		f.Close()
		if string(head) != "%PDF-" {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "this is not a PDF file — export the resume as PDF")
		}
	}

	text, extractErr := extractPDFText(tmpPath)
	text = scoring.NormalizeResumeText(text)
	if extractErr != nil || len(text) < parseTextMinLen {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "this PDF has no text layer (likely a scan) — fill the resume manually")
	}
	text = scoring.TruncateText(text, parseTextMaxLen)

	ctx := c.Request().Context()
	existing, err := h.resumes.GetByUserID(ctx, userID)
	switch {
	case errors.Is(err, repository.ErrResumeNotFound):
		existing = repository.Resume{IsActive: false, Source: "file_parse"}
	case err != nil:
		h.log.Error("get resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	draft := existing
	draft.UserID = userID
	draft.Source = "file_parse"
	draft.SourceText = text
	draft.ParseStatus = "processing"
	if draft.Links == "" {
		draft.Links = "[]"
	}
	if existing.ID == 0 {
		draft.IsActive = false
	}

	saved, err := h.resumes.UpsertResume(ctx, draft)
	if err != nil {
		h.log.Error("upsert parsed resume failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	go h.processParsedResume(userID, text, saved, existing)

	return c.JSON(http.StatusAccepted, dto.ParseStartResponse{Status: "processing"})
}
