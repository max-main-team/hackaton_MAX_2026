package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/auth"
)

const contextKey = "user_id"

func RequireAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			userID, err := auth.Parse(token, secret)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			c.Set(contextKey, userID)
			return next(c)
		}
	}
}

func UserIDFromContext(c echo.Context) (int64, error) {
	id, ok := c.Get(contextKey).(int64)
	if !ok {
		return 0, errors.New("user_id is not in context")
	}
	return id, nil
}
