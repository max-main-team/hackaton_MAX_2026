package handler

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
	"max-miniapp/backend/internal/scoring"
)

type MatchingHandler struct {
	companies *repository.CompanyRepo
	vacancies *repository.VacancyRepo
	users     *repository.UserRepo
	resumes   *repository.ResumeRepo
	matching  *repository.MatchingRepo
	scores    *repository.ScoreRepo
	ai        *scoring.AIClient
	aiModel   string
	log       *slog.Logger
}

func NewMatchingHandler(
	companies *repository.CompanyRepo,
	vacancies *repository.VacancyRepo,
	users *repository.UserRepo,
	resumes *repository.ResumeRepo,
	matching *repository.MatchingRepo,
	scores *repository.ScoreRepo,
	ai *scoring.AIClient,
	aiModel string,
	log *slog.Logger,
) *MatchingHandler {
	return &MatchingHandler{
		companies: companies, vacancies: vacancies, users: users,
		resumes: resumes, matching: matching, scores: scores,
		ai: ai, aiModel: aiModel, log: log,
	}
}

// @Summary     Кандидаты под вакансию
// @Description mode=list — постранично (limit/offset), mode=feed — лента без пагинации.
// @Description Только кандидаты с активным резюме и без действия по этой вакансии.
// @Tags        matching
// @Produce     json
// @Param       id     path integer true "ID вакансии"
// @Param       mode   query  string  false "list" Enums(list, feed)
// @Param       limit  query  integer false "20"
// @Param       offset query  integer false "0"
// @Success     200 {object} dto.CandidatesResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/vacancies/{id}/candidates [get]
func (h *MatchingHandler) Candidates(c echo.Context) error {
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
		hasAction, err := h.matching.HasActionForVacancy(ctx, vacancyID, r.UserID)
		if err != nil {
			h.log.Error("check action failed", slog.Any("err", err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
		}
		if hasAction {
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
			Score:      score,
			FinalScore: score,
			Breakdown:  breakdown,
			User:       dto.FromUser(user),
			Resume:     dto.FromResume(r),
		}})
	}

	aiMissing := false
	if h.ai.Enabled() {
		for idx := range entries {
			r := entries[idx].item.Resume
			cached, err := h.scores.GetScore(ctx, r.ID, vacancyID)
			if err == nil && cached != nil && cached.AIScore != nil {
				entries[idx].item.AIScore = cached.AIScore
				entries[idx].item.AIComment = cached.AIComment
				entries[idx].item.FinalScore = int(float64(entries[idx].item.Score)*0.6 + float64(*entries[idx].item.AIScore)*0.4)
				continue
			}
			if err != nil {
				h.log.Error("get cached score failed", slog.Any("err", err))
			}
			aiMissing = true
		}
	}
	if aiMissing {
		sem := make(chan struct{}, 4)
		var wg sync.WaitGroup
		for idx := range entries {
			if entries[idx].item.AIScore != nil {
				continue
			}
			wg.Go(func() {
				sem <- struct{}{}
				defer func() {
					<-sem
					if r := recover(); r != nil {
						h.log.Error("ai enrichment panic", slog.Any("err", r))
					}
				}()

				r := entries[idx].item.Resume
				bgCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				result, err := h.ai.ScoreAI(bgCtx, vac, scoring.Resume{
					Skills:           r.Skills,
					ExperienceMonths: r.ExperienceMonths,
					City:             r.City,
					WorkFormat:       r.WorkFormat,
				})
				if err != nil && strings.Contains(err.Error(), "429") {
					time.Sleep(3 * time.Second)
					result, err = h.ai.ScoreAI(bgCtx, vac, scoring.Resume{
						Skills:           r.Skills,
						ExperienceMonths: r.ExperienceMonths,
						City:             r.City,
						WorkFormat:       r.WorkFormat,
					})
				}
				if err != nil {
					h.log.Error("ai scoring failed", slog.Any("err", err))
					return
				}
				ai := result.Score
				_ = h.scores.UpsertScore(ctx, repository.ResumeScore{
					ResumeID: r.ID, VacancyID: vacancyID,
					AlgoScore: entries[idx].item.Score, AIScore: &ai,
					AIComment: result.Comment, AIModel: h.aiModel,
				})
			})
		}
	}

	slices.SortStableFunc(entries, func(a, b entry) int {
		return cmp.Compare(b.item.FinalScore, a.item.FinalScore)
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

// @Summary     Действие рекрутера на кандидата
// @Tags        matching
// @Accept      json
// @Produce     json
// @Param       id              path integer true "ID вакансии"
// @Param       candidateUserId path integer true "ID кандидата"
// @Param       request body dto.RecruiterActionResponse true "action"
// @Success     200 {object} dto.RecruiterActionResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/vacancies/{id}/candidates/{candidateUserId}/action [post]
func (h *MatchingHandler) Action(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	vacancyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid vacancy id")
	}
	candidateID, err := strconv.ParseInt(c.Param("candidateUserId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid candidate id")
	}

	var in dto.RecruiterActionResponse
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if in.Action != "invite" && in.Action != "skip" {
		return echo.NewHTTPError(http.StatusBadRequest, "action must be invite or skip")
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

	if in.Action == "invite" {
		ok, err := h.companies.ConsumeInvite(ctx, vacancy.CompanyID)
		if err != nil {
			h.log.Error("consume invite failed", slog.Any("err", err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
		}
		if !ok {
			return echo.NewHTTPError(http.StatusConflict, "invite quota exceeded")
		}
	}

	action, err := h.matching.GetOrCreateAction(ctx, vacancyID, userID, candidateID, in.Action)
	if err != nil {
		h.log.Error("create action failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.RecruiterActionResponse{
		ID: action.ID, Action: action.Action, CreatedAt: action.CreatedAt,
	})
}

// @Summary     Мои приглашения
// @Tags        invitations
// @Produce     json
// @Success     200 {array} dto.Invitation
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/my/invitations [get]
func (h *MatchingHandler) Invitations(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	details, err := h.matching.ListInvitationsByCandidate(c.Request().Context(), userID)
	if err != nil {
		h.log.Error("list invitations failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	now := time.Now()
	rank := map[string]int{"pending": 0, "overdue": 1, "responded": 2}
	out := make([]dto.Invitation, 0, len(details))
	for _, d := range details {
		deadline := d.CreatedAt.Add(time.Duration(d.TTLHours) * time.Hour)
		status, hoursLeft := "responded", 0.0
		switch {
		case d.Response != nil:
		case now.After(deadline):
			status = "overdue"
		default:
			status = "pending"
			hoursLeft = deadline.Sub(now).Hours()
		}
		out = append(out, dto.Invitation{
			ID:         d.ActionID,
			Status:     status,
			DeadlineAt: deadline,
			HoursLeft:  hoursLeft,
			Company:    dto.InvitationCompany{ID: d.CompanyID, Name: d.CompanyName, Verified: d.CompanyVerified},
			Vacancy: dto.InvitationVacancy{
				ID: d.VacancyID, Title: d.VacancyTitle, City: d.VacancyCity,
				WorkFormat: d.VacancyWorkFormat, SalaryMin: d.VacancySalaryMin, SalaryMax: d.VacancySalaryMax,
			},
			Response: d.Response,
		})
	}

	slices.SortStableFunc(out, func(a, b dto.Invitation) int {
		if ra, rb := rank[a.Status], rank[b.Status]; ra != rb {
			return cmp.Compare(ra, rb)
		}
		return a.DeadlineAt.Compare(b.DeadlineAt)
	})

	return c.JSON(http.StatusOK, out)
}

// @Summary     Ответить на приглашение
// @Tags        invitations
// @Accept      json
// @Produce     json
// @Param       id     path integer true "ID приглашения"
// @Param       request body dto.RespondRequest true "ответ"
// @Success     200 {object} dto.MatchResult
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/invitations/{id}/respond [post]
func (h *MatchingHandler) Respond(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	actionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid invitation id")
	}

	var in dto.RespondRequest
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if in.Response != "accept" && in.Response != "decline" {
		return echo.NewHTTPError(http.StatusBadRequest, "response must be accept or decline")
	}

	ctx := c.Request().Context()
	detail, err := h.matching.GetInvitationDetail(ctx, actionID)
	if err != nil {
		h.log.Error("get invitation failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusNotFound, "invitation not found")
	}
	if detail.CandidateUserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not your invitation")
	}
	if detail.Response != nil {
		return echo.NewHTTPError(http.StatusConflict, "already responded")
	}
	deadline := detail.CreatedAt.Add(time.Duration(detail.TTLHours) * time.Hour)
	if time.Now().After(deadline) {
		return echo.NewHTTPError(http.StatusConflict, "invitation expired")
	}

	if err := h.matching.CreateResponse(ctx, actionID, in.Response); err != nil {
		h.log.Error("create response failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	result := dto.MatchResult{
		ID:               actionID,
		Response:         in.Response,
		Company:          dto.InvitationCompany{ID: detail.CompanyID, Name: detail.CompanyName, Verified: detail.CompanyVerified},
		Vacancy:          dto.InvitationVacancy{ID: detail.VacancyID, Title: detail.VacancyTitle, City: detail.VacancyCity},
		RecruiterContact: detail.RecruiterLogin,
	}
	return c.JSON(http.StatusOK, result)
}
