package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
)

type CandidatesHandler struct {
	resumes *repository.ResumeRepo
	users   *repository.UserRepo
	log     *slog.Logger
}

func NewCandidatesHandler(resumes *repository.ResumeRepo, users *repository.UserRepo, log *slog.Logger) *CandidatesHandler {
	return &CandidatesHandler{resumes: resumes, users: users, log: log}
}

// @Summary     Все кандидаты (для компаний)
// @Description Активные резюме всех кандидатов, поиск по title/skills/city,
// @Description пагинация limit/offset. Роль рекрутера обязательна.
// @Tags        matching
// @Produce     json
// @Param       q      query string  false "подстрока в title/skills/city"
// @Param       limit  query integer false "20"
// @Param       offset query integer false "0"
// @Success     200 {object} dto.AllCandidatesResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/candidates [get]
func (h *CandidatesHandler) AllCandidates(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := h.users.GetByID(c.Request().Context(), userID)
	if err != nil {
		h.log.Error("get user failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if user.Role != "recruiter" {
		return echo.NewHTTPError(http.StatusForbidden, "recruiter role required")
	}

	query := strings.TrimSpace(c.QueryParam("q"))
	if len(query) > 100 {
		query = query[:100]
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	cards, total, err := h.resumes.ListCandidateCards(c.Request().Context(), query, limit, offset)
	if err != nil {
		h.log.Error("list candidate cards failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	items := make([]dto.CandidateCard, 0, len(cards))
	for _, card := range cards {
		items = append(items, dto.CandidateCard{
			User:   dto.FromUser(card.User),
			Resume: dto.FromResume(card.Resume),
		})
	}
	return c.JSON(http.StatusOK, dto.AllCandidatesResponse{Total: total, Items: items})
}
