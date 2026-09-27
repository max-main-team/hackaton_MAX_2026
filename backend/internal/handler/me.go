package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/middleware"
	"max-miniapp/backend/internal/repository"
)

type MeHandler struct {
	users *repository.UserRepo
	log   *slog.Logger
}

func NewMeHandler(users *repository.UserRepo, log *slog.Logger) *MeHandler {
	return &MeHandler{users: users, log: log}
}

// Me возвращает текущего пользователя и статус согласия на обработку ПДн.
//
//	@Summary     Текущий пользователь
//	@Tags        profile
//	@Produce     json
//	@Success     200 {object} dto.MeResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     404 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/me [get]
func (h *MeHandler) Me(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := h.users.GetByID(c.Request().Context(), userID)
	if err != nil {
		h.log.Error("get user failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	consentAt, err := h.users.GetConsentAcceptedAt(c.Request().Context(), userID, repository.ConsentPersonalData)
	if err != nil {
		h.log.Error("get consent failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.MeResponse{
		User:                   dto.FromUser(user),
		PersonalDataAcceptedAt: consentAt,
	})
}

// SetRole выбирает роль пользователя и фиксирует согласие на обработку ПДн.
//
//	@Summary     Выбор роли (+ согласие на обработку ПДн)
//	@Tags        profile
//	@Accept      json
//	@Produce     json
//	@Param       request body dto.RoleRequest true "роль и согласие"
//	@Success     200 {object} dto.MeResponse
//	@Failure     400 {object} dto.ErrorResponse
//	@Failure     401 {object} dto.ErrorResponse
//	@Failure     500 {object} dto.ErrorResponse
//	@Security    BearerAuth
//	@Router      /api/v1/me/role [post]
func (h *MeHandler) SetRole(c echo.Context) error {
	userID, err := middleware.UserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.RoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := dto.ValidateRole(req.Role); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if !req.AcceptPersonalData {
		return echo.NewHTTPError(http.StatusBadRequest, "acceptPersonalData is required")
	}

	ctx := c.Request().Context()
	if err := h.users.UpdateRole(ctx, userID, req.Role); err != nil {
		h.log.Error("update role failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	if err := h.users.UpsertConsent(ctx, userID, repository.ConsentPersonalData); err != nil {
		h.log.Error("upsert consent failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	user, err := h.users.GetByID(ctx, userID)
	if err != nil {
		h.log.Error("get user failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	consentAt, err := h.users.GetConsentAcceptedAt(ctx, userID, repository.ConsentPersonalData)
	if err != nil {
		h.log.Error("get consent failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.MeResponse{
		User:                   dto.FromUser(user),
		PersonalDataAcceptedAt: consentAt,
	})
}
