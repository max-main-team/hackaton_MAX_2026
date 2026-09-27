package handler

import (
	"log/slog"
	"net/http"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
	"max-miniapp/backend/internal/scoring"
)

type CandidatesHandler struct {
	companies *repository.CompanyRepo
	vacancies *repository.VacancyRepo
	users     *repository.UserRepo
	resumes   *repository.ResumeRepo
	log       *slog.Logger
}

func NewCandidatesHandler(
	companies *repository.CompanyRepo,
	vacancies *repository.VacancyRepo,
	users *repository.UserRepo,
	resumes *repository.ResumeRepo,
	log *slog.Logger,
) *CandidatesHandler {
	return &CandidatesHandler{
		companies: companies, vacancies: vacancies, users: users, resumes: resumes, log: log,
	}
}

// Candidates возвращает приоритизированный список кандидатов под вакансию.
//
//	@Summary     Кандидаты под вакансию
//	@Description mode=list — постранично (limit/offset), mode=feed — лента без пагинации.
//	@Description Только кандидаты с активным резюме. Сортировка по score.
//	@Tags        matching
//	@Produce     json
//	@Param       id     path integer true "ID вакансии"
//	@Param       mode   query  string  false "list" Enums(list, feed)
//	@Param       limit  query  integer false "20"
//	@Param       offset query  integer false "0"
//	@Success     200 {object} dto.CandidatesResponse
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     403 {object} dto.ErrorResponse
//	@Failure     404 {object} dto.ErrorResponse
//	@Failure     409 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/vacancies/{id}/candidates [get]
func (h *CandidatesHandler) Candidates(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	vacancyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid vacancy id")
	}

	ctx := c.Request().Context()
	vacancy, err := h.vacancies.GetByID(ctx, vacancyID)
	if err != nil {
		h.log.Error("get vacancy failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusNotFound, "vacancy not found")
	}

	isMember, err := h.companies.IsMember(ctx, vacancy.CompanyID, userID)
	if err != nil {
		h.log.Error("check membership failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a member of this company")
	}
	if !vacancy.IsActive {
		return echo.NewHTTPError(http.StatusConflict, "vacancy is not active")
	}

	resumes, err := h.resumes.ListActive(ctx)
	if err != nil {
		h.log.Error("list resumes failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	vac := scoring.Vacancy{
		RequiredSkills:      vacancy.RequiredSkills,
		MinExperienceMonths: vacancy.MinExperienceMonths,
		City:                vacancy.City,
		WorkFormat:          vacancy.WorkFormat,
	}

	type entry struct {
		item dto.CandidateItem
	}
	var entries []entry
	for _, r := range resumes {
		if r.UserID == userID {
			continue
		}
		user, err := h.users.GetByID(ctx, r.UserID)
		if err != nil || user.Role != "candidate" {
			continue
		}
		score, breakdown := scoring.Score(vac, scoring.Resume{
			Skills:           r.Skills,
			ExperienceMonths: r.ExperienceMonths,
			City:             r.City,
			WorkFormat:       r.WorkFormat,
		})
		entries = append(entries, entry{item: dto.CandidateItem{
			Score:     score,
			Breakdown: breakdown,
			User:      dto.FromUser(user),
			Resume:    dto.FromResume(r),
		}})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].item.Score > entries[j].item.Score
	})

	mode := c.QueryParam("mode")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	resp := dto.CandidatesResponse{Total: len(entries)}
	if mode == "feed" {
		for _, e := range entries {
			resp.Items = append(resp.Items, e.item)
		}
	} else {
		for i := offset; i < len(entries) && len(resp.Items) < limit; i++ {
			resp.Items = append(resp.Items, entries[i].item)
		}
	}
	if resp.Items == nil {
		resp.Items = []dto.CandidateItem{}
	}

	return c.JSON(http.StatusOK, resp)
}
