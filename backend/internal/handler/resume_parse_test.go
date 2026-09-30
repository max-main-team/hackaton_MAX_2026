package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"max-miniapp/backend/internal/database"
	"max-miniapp/backend/internal/repository"
	"max-miniapp/backend/internal/scoring"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5433/maxapp?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("no local db: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("no local db: %v", err)
	}
	require.NoError(t, database.Migrate(ctx, pool))
	t.Cleanup(pool.Close)
	return pool
}

func stubAI(t *testing.T, content string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": content}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

const parseFixtureText = "Михаил Соковых\r\nBackend-разработчик\r\n\r\nОпыт 5 лет: Go, PostgreSQL, Docker.\r\nЖиву в Санкт-Петербурге, удалёнка.\r\nСтраница 2 из 4\r\nЗарплата от 250к.\r\n"

func newParseContext(t *testing.T, userID int64, body string) echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/resume/parse", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	return c
}

func waitParseStatus(t *testing.T, pool *pgxpool.Pool, userID int64, want string) repository.Resume {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last repository.Resume
	for time.Now().Before(deadline) {
		r, err := repository.NewResumeRepo(pool).GetByUserID(context.Background(), userID)
		if err == nil {
			last = r
			if r.ParseStatus == want {
				return r
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("parse_status %q not reached, last=%q", want, last.ParseStatus)
	return last
}

func TestParseIntegrationDraftFlow(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	draft := `{"title": "Backend-разработчик", "skills": "go, go, postgresql, docker", "experience_months": 60,
		"about": "Сервисы на Go", "education": "ИТМО", "city": "Санкт-Петербург",
		"work_format": "удалёнка", "employment_type": "полный день",
		"salary_min": 200000, "salary_max": 300000, "notes": "всё нашёл"}`
	h := NewResumeHandler(repository.NewResumeRepo(pool), scoring.NewAIClient("k", stubAI(t, draft).URL, "m"), testLogger())

	userID := 720000000 + time.Now().UnixNano()%1000000
	_, err := pool.Exec(ctx, `INSERT INTO users (id, username, first_name) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`, userID, "it_parse", "Mikhail")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM resume_versions WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM resumes WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	c := newParseContext(t, userID, fmt.Sprintf(`{"text": %q}`, parseFixtureText))
	require.NoError(t, h.Parse(c))
	assert.Equal(t, http.StatusAccepted, c.Response().Status)

	r := waitParseStatus(t, pool, userID, "done")

	assert.Equal(t, "Backend-разработчик", r.Title)
	assert.Equal(t, "go, postgresql, docker", r.Skills)
	assert.Equal(t, 60, r.ExperienceMonths)
	assert.Equal(t, "Санкт-Петербург", r.City)
	assert.Equal(t, "remote", r.WorkFormat)
	assert.Equal(t, "full_time", r.EmploymentType)
	assert.NotNil(t, r.SalaryMin)
	assert.Equal(t, 200000, *r.SalaryMin)
	assert.Equal(t, true, r.IsActive)
	assert.Equal(t, "file_parse", r.Source)

	assert.NotContains(t, r.SourceText, "\r")
	assert.NotContains(t, r.SourceText, "Страница 2 из 4")
	assert.Contains(t, r.SourceText, "PostgreSQL")
}

func TestParseIntegrationAIFailure(t *testing.T) {
	pool := testPool(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	h := NewResumeHandler(repository.NewResumeRepo(pool), scoring.NewAIClient("k", srv.URL, "m"), testLogger())

	userID := 720000000 + time.Now().UnixNano()%1000000
	_, err := pool.Exec(context.Background(), `INSERT INTO users (id, username, first_name) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`, userID, "it_parse_fail", "Mikhail")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM resume_versions WHERE user_id = $1`, userID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM resumes WHERE user_id = $1`, userID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	c := newParseContext(t, userID, fmt.Sprintf(`{"text": %q}`, parseFixtureText))
	require.NoError(t, h.Parse(c))
	assert.Equal(t, http.StatusAccepted, c.Response().Status)

	r := waitParseStatus(t, pool, userID, "failed")
	assert.Equal(t, "file_parse", r.Source)
	assert.Contains(t, r.SourceText, "PostgreSQL")
	assert.Equal(t, "", r.Title)
	assert.False(t, r.IsActive)
}

func TestParseIntegrationValidation(t *testing.T) {
	pool := testPool(t)
	h := NewResumeHandler(repository.NewResumeRepo(pool), scoring.NewAIClient("k", "http://127.0.0.1:1", "m"), testLogger())

	userID := int64(730000001)
	c := newParseContext(t, userID, `{"text": ""}`)
	require.Error(t, h.Parse(c))
}
