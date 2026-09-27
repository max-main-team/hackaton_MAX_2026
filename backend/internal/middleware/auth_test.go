package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"max-miniapp/backend/internal/auth"
)

func TestRequireAuthAllowsValidToken(t *testing.T) {
	token, err := auth.Issue(42, "secret", time.Now())
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var gotID int64
	handler := RequireAuth("secret")(func(c echo.Context) error {
		id, err := UserIDFromContext(c)
		require.NoError(t, err)
		gotID = id
		return c.NoContent(http.StatusOK)
	})

	require.NoError(t, handler(c))
	assert.Equal(t, int64(42), gotID)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAuth("secret")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestRequireAuthRejectsBadToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAuth("secret")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}
