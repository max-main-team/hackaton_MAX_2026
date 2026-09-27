package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/config"
	"max-miniapp/backend/internal/repository"
)

type AuthHandler struct {
	users *repository.UserRepo
	cfg   *config.Config
	log   *slog.Logger
}

func NewAuthHandler(users *repository.UserRepo, cfg *config.Config, log *slog.Logger) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg, log: log}
}

type authRequest struct {
	InitData string `json:"initData"`
}

type maxUser struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	PhotoURL     string `json:"photo_url"`
	LanguageCode string `json:"language_code"`
}

func (h *AuthHandler) Auth(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.InitData == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "initData is required")
	}

	user, err := parseInitData(req.InitData)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid initData")
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

	return c.JSON(http.StatusOK, saved)
}

func parseInitData(initData string) (maxUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return maxUser{}, fmt.Errorf("parse initData: %w", err)
	}

	var user maxUser
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil {
		return maxUser{}, fmt.Errorf("parse user from initData: %w", err)
	}
	if user.ID == 0 {
		return maxUser{}, errors.New("user id is missing in initData")
	}
	return user, nil
}
