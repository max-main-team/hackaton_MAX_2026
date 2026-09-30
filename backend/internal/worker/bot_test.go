package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteCommand(t *testing.T) {
	assert.Equal(t, "start", routeCommand("/start"))
	assert.Equal(t, "start", routeCommand("  Start  "))
	assert.Equal(t, "help", routeCommand("/help"))
	assert.Equal(t, "help", routeCommand("ПОМОЩЬ"))
	assert.Equal(t, "help", routeCommand("как это работает"))
	assert.Equal(t, "status", routeCommand("/status"))
	assert.Equal(t, "status", routeCommand("Статус"))
	assert.Equal(t, "unknown", routeCommand("привет"))
	assert.Equal(t, "unknown", routeCommand("   "))
}

func TestStatusMessageNewUser(t *testing.T) {
	msg := statusMessage(botUserStatus{})
	text, _ := msg["text"].(string)
	assert.Contains(t, text, "не видел")
}

func TestStatusMessageCandidate(t *testing.T) {
	msg := statusMessage(botUserStatus{Known: true, Role: "candidate", ResumeTitle: "Go-разработчик", ResumeActive: true, Pending: 2})
	text, _ := msg["text"].(string)
	assert.Contains(t, text, "Go-разработчик")
	assert.Contains(t, text, "активно")
	assert.Contains(t, text, "2")

	msg = statusMessage(botUserStatus{Known: true, Role: "candidate", ResumeTitle: "Go-разработчик"})
	text, _ = msg["text"].(string)
	assert.Contains(t, text, "на паузе")
}

func TestStatusMessageRecruiter(t *testing.T) {
	msg := statusMessage(botUserStatus{Known: true, Role: "recruiter", CompanyName: "Ромашка", Verified: true, InviteQuota: 6, InviteUsed: 2})
	text, _ := msg["text"].(string)
	assert.Contains(t, text, "Ромашка")
	assert.Contains(t, text, "верифицирована: да")
	assert.Contains(t, text, "2 из 6")
}

func TestKeyboardMessageButtons(t *testing.T) {
	msg := greetingMessage()
	atts, _ := json.Marshal(msg["attachments"])
	assert.Contains(t, string(atts), "open_app")
	assert.Contains(t, string(atts), "callback")
	assert.Equal(t, true, msg["notify"])
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
	t.Cleanup(pool.Close)
	return pool
}

func TestFetchUserStatusIntegration(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	userID := 740000000 + time.Now().UnixNano()%100000
	_, err := pool.Exec(ctx, `INSERT INTO users (id, username, first_name, role) VALUES ($1, $2, $3, 'candidate')
		ON CONFLICT (id) DO UPDATE SET role = 'candidate'`, userID, "it_bot", "Anna")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO resumes (user_id, title, is_active) VALUES ($1, 'Frontend', TRUE)
		ON CONFLICT (user_id) DO UPDATE SET title = 'Frontend', is_active = TRUE`, userID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM resume_versions WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM resumes WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	w := NewBotWorker(pool, "", discardLogger())
	s := w.fetchUserStatus(ctx, userID)
	assert.True(t, s.Known)
	assert.Equal(t, "candidate", s.Role)
	assert.Equal(t, "Frontend", s.ResumeTitle)
	assert.True(t, s.ResumeActive)
	assert.Equal(t, 0, s.Pending)

	unknown := w.fetchUserStatus(ctx, 1)
	assert.False(t, unknown.Known)
	msg := statusMessage(unknown)
	text, _ := msg["text"].(string)
	assert.Contains(t, text, "не видел")
}

func TestBotCommandsShape(t *testing.T) {
	cmds := botCommands()
	require.Len(t, cmds, 3)
	names := []string{}
	for _, c := range cmds {
		names = append(names, c.Name)
		assert.NotEmpty(t, c.Description)
	}
	assert.ElementsMatch(t, []string{"start", "help", "status"}, names)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
