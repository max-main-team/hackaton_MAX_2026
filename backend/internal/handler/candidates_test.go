package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"max-miniapp/backend/internal/repository"
)

func newCandidatesContext(t *testing.T, userID int64, qs string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/candidates?"+qs, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	return c, rec
}

func TestAllCandidatesIntegration(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := repository.NewResumeRepo(pool)
	userRepo := repository.NewUserRepo(pool)
	h := NewCandidatesHandler(repo, userRepo, testLogger())

	suffix := time.Now().UnixNano() % 1000000
	mark := fmt.Sprintf("mk%d", suffix)
	recruiterID := int64(741000000 + suffix%1000)
	candA := int64(742000000 + suffix%1000)
	candB := int64(743000000 + suffix%1000)

	for _, u := range []struct {
		id   int64
		role string
		name string
	}{{recruiterID, "recruiter", "Rec"}, {candA, "candidate", "Anna"}, {candB, "candidate", "Boris"}} {
		_, err := userRepo.UpsertUser(ctx, repository.User{ID: u.id, Username: fmt.Sprintf("it_cand_%d", u.id), FirstName: u.name})
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `UPDATE users SET role = $2 WHERE id = $1`, u.id, u.role)
		require.NoError(t, err)
	}
	seed := func(userID int64, title, skills, city string, active bool) {
		_, err := repo.UpsertResume(ctx, repository.Resume{
			UserID: userID, Title: title, Skills: skills, City: city,
			WorkFormat: "remote", EmploymentType: "full_time",
			Links: "[]", Source: "manual", IsActive: active,
		})
		require.NoError(t, err)
	}
	titleA := "GoDev" + mark
	titleB := "Front" + mark
	seed(candA, titleA, "go, "+mark, "Санкт-Петербург", true)
	seed(candB, titleB, "react, typescript", "Казань", true)

	t.Cleanup(func() {
		for _, id := range []int64{recruiterID, candA, candB} {
			_, _ = pool.Exec(ctx, `DELETE FROM resume_versions WHERE user_id = $1`, id)
			_, _ = pool.Exec(ctx, `DELETE FROM resumes WHERE user_id = $1`, id)
			_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
		}
	})

	t.Run("recruiter sees all", func(t *testing.T) {
		c, rec := newCandidatesContext(t, recruiterID, "q="+mark+"&limit=100")
		require.NoError(t, h.AllCandidates(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		s := rec.Body.String()
		assert.Contains(t, s, titleA)
		assert.Contains(t, s, titleB)
		assert.NotContains(t, s, `"final_score"`)
	})

	t.Run("search filters", func(t *testing.T) {
		c, rec := newCandidatesContext(t, recruiterID, "q="+url.QueryEscape("go, "+mark))
		require.NoError(t, h.AllCandidates(c))
		s := rec.Body.String()
		assert.Contains(t, s, titleA)
		assert.NotContains(t, s, titleB)
	})

	t.Run("inactive hidden", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE resumes SET is_active = FALSE WHERE user_id = $1`, candB)
		require.NoError(t, err)
		c, rec := newCandidatesContext(t, recruiterID, "q="+mark+"&limit=100")
		require.NoError(t, h.AllCandidates(c))
		assert.NotContains(t, rec.Body.String(), titleB)
		_, _ = pool.Exec(ctx, `UPDATE resumes SET is_active = TRUE WHERE user_id = $1`, candB)
	})

	t.Run("candidate forbidden", func(t *testing.T) {
		c, _ := newCandidatesContext(t, candA, "")
		err := h.AllCandidates(c)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)
	})

	t.Run("pagination", func(t *testing.T) {
		c, rec := newCandidatesContext(t, recruiterID, "q="+mark+"&limit=1&offset=1")
		require.NoError(t, h.AllCandidates(c))
		s := rec.Body.String()
		assert.Equal(t, 1, strings.Count(s, `"resume":`))
		assert.Contains(t, s, `"total":2`)
	})
}
