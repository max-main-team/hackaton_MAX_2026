package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/auth"
	"max-miniapp/backend/internal/config"
	"max-miniapp/backend/internal/dto"
	"max-miniapp/backend/internal/repository"
)

const initDataMaxAge = time.Hour

type AuthHandler struct {
	users *repository.UserRepo
	cfg   *config.Config
	log   *slog.Logger
}

func NewAuthHandler(users *repository.UserRepo, cfg *config.Config, log *slog.Logger) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg, log: log}
}

// @Summary     Авторизация по initData
// @Description Принимает window.WebApp.initData, проверяет подпись (токен бота),
// @Description  срок давности, апсертит пользователя и выдаёт JWT (7 дней).
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dto.AuthRequest true "initData из MAX Bridge"
// @Success     200 {object} dto.AuthResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /api/v1/auth [post]
func (h *AuthHandler) Auth(c echo.Context) error {
	var req dto.AuthRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.InitData == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "initData is required")
	}

	user, authDate, err := auth.ParseInitData(req.InitData)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid initData")
	}
	if time.Since(authDate) > initDataMaxAge {
		return echo.NewHTTPError(http.StatusUnauthorized, "initData expired")
	}

	if h.cfg.MaxBotToken != "" {
		if err := auth.Verify(req.InitData, h.cfg.MaxBotToken); err != nil {
			h.log.Warn("initData signature check failed", slog.Any("err", err))
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid initData signature")
		}
	}

	saved, err := h.users.UpsertUser(c.Request().Context(), repository.User{
		ID:           user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PhotoURL:     user.PhotoURL,
		LanguageCode: user.LanguageCode,
	})
	if err != nil {
		h.log.Error("upsert user failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	token, err := auth.Issue(saved.ID, h.cfg.JWTSecret, time.Now())
	if err != nil {
		h.log.Error("issue token failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	if startParam := auth.StartParam(req.InitData); strings.HasPrefix(startParam, "ref_") {
		if refID, err := strconv.ParseInt(strings.TrimPrefix(startParam, "ref_"), 10, 64); err == nil && refID != saved.ID {
			if err := h.users.SetReferrerIfEmpty(c.Request().Context(), saved.ID, refID); err != nil {
				h.log.Warn("set referrer failed", slog.Any("err", err))
			}
		}
	}

	if startParam := auth.StartParam(req.InitData); strings.HasPrefix(startParam, "refc_") {
		if refCompanyID, err := strconv.ParseInt(strings.TrimPrefix(startParam, "refc_"), 10, 64); err == nil {
			if err := h.users.SetPendingRefCompany(c.Request().Context(), saved.ID, refCompanyID); err != nil {
				h.log.Warn("set pending ref company failed", slog.Any("err", err))
			}
		}
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Token:        token,
		User:         dto.FromUser(saved),
		ReferralCode: fmt.Sprintf("ref_%d", saved.ID),
	})
}

// @Summary     Демо-вход (только dev)
// @Description Создаёт/возвращает тестового пользователя и JWT. В prod отключён.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dto.DemoAuthRequest false "параметры тестового пользователя"
// @Success     200 {object} dto.AuthResponse
// @Failure     403 {object} dto.ErrorResponse
// @Router      /api/v1/auth/demo [post]
func (h *AuthHandler) DemoAuth(c echo.Context) error {
	var req dto.DemoAuthRequest
	_ = c.Bind(&req)

	// только тестовые 700000000–799999999
	if req.UserID != 0 && (req.UserID < 700000000 || req.UserID > 799999999) {
		req.UserID = 0
	}

	id := req.UserID
	if id == 0 {
		id = 700000000 + time.Now().UnixMilli()%100000
	}
	firstName := req.FirstName
	if firstName == "" {
		firstName = "Demo"
	}

	saved, err := h.users.UpsertUser(c.Request().Context(), repository.User{
		ID:           id,
		Username:     fmt.Sprintf("demo_%d", id),
		FirstName:    firstName,
		LanguageCode: "ru",
	})
	if err != nil {
		h.log.Error("demo upsert user failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	token, err := auth.Issue(saved.ID, h.cfg.JWTSecret, time.Now())
	if err != nil {
		h.log.Error("issue token failed", slog.Any("err", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Token:        token,
		User:         dto.FromUser(saved),
		ReferralCode: fmt.Sprintf("ref_%d", saved.ID),
	})
}
