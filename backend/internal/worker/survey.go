package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	surveyInterval  = time.Hour
	surveyIdleDays  = 7
	sendPauseMillis = 700
	apiBaseURL      = "https://platform-api2.max.ru"
)

type SurveyWorker struct {
	pool     *pgxpool.Pool
	botToken string
	log      *slog.Logger
	interval time.Duration
}

func NewSurveyWorker(pool *pgxpool.Pool, botToken string, log *slog.Logger) *SurveyWorker {
	return &SurveyWorker{pool: pool, botToken: botToken, log: log, interval: time.Hour}
}

func (w *SurveyWorker) Start(ctx context.Context) {
	if w.botToken == "" {
		w.log.Warn("survey worker disabled: MAX_BOT_TOKEN is empty")
		return
	}
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		w.tick(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.tick(ctx)
			}
		}
	}()
	w.log.Info("survey worker started", slog.Duration("interval", w.interval))
}

func (w *SurveyWorker) tick(ctx context.Context) {
	rows, err := w.pool.Query(ctx, `
		SELECT u.id
		FROM resumes r
		JOIN users u ON u.id = r.user_id
		WHERE r.is_active = TRUE
		  AND r.last_confirmed_at < now() - interval '7 days'
	`)
	if err != nil {
		w.log.Error("survey: select users failed", slog.Any("err", err))
		return
	}

	var userIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}
	rows.Close()

	for _, userID := range userIDs {
		if ctx.Err() != nil {
			return
		}
		if err := w.sendSurvey(ctx, userID); err != nil {
			w.log.Error("survey send failed", slog.Int64("user_id", userID), slog.Any("err", err))
		} else {
			w.log.Info("survey sent", slog.Int64("user_id", userID))
		}
		time.Sleep(sendPauseMillis * time.Millisecond)
	}
}

type openAppButton struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type surveyMessage struct {
	Text        string `json:"text"`
	Attachments []any  `json:"attachments"`
	Notify      bool   `json:"notify"`
}

func (w *SurveyWorker) sendSurvey(ctx context.Context, userID int64) error {
	message := surveyMessage{
		Text: "Ваш подбор вакансий ещё актуален? Подтвердите в мини-приложении, чтобы компании продолжили вас видеть.",
		Attachments: []any{map[string]any{
			"type": "inline_keyboard",
			"payload": map[string]any{
				"buttons": [][]openAppButton{
					{{Type: "open_app", Text: "Открыть приложение"}},
				},
			},
		}},
		Notify: true,
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/messages?user_id=%d", apiBaseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", w.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("max api status %d", resp.StatusCode)
	}
	return nil
}
