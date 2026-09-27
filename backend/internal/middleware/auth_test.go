package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"max-miniapp/backend/internal/auth"
)

func TestRequireAuthAllowsValidToken(t *testing.T) {
	token, err := auth.Issue(42, "secret", time.Now())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequireAuth("secret")(func(c echo.Context) error {
		id, err := UserIDFromContext(c)
		if err != nil {
			t.Fatalf("user id from context: %v", err)
		}
		if id != 42 {
			t.Fatalf("want 42, got %d", id)
		}
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
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
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 HTTPError, got %v", err)
	}
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
	if err == nil {
		t.Fatal("expected error for bad token, got nil")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 HTTPError, got %v", err)
	}
}
