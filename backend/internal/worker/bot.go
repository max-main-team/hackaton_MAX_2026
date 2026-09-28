package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const updatesURL = apiBaseURL + "/updates"

type BotWorker struct {
	pool     *pgxpool.Pool
	botToken string
	log      *slog.Logger
	marker   int64
	client   *http.Client
}

func NewBotWorker(pool *pgxpool.Pool, botToken string, log *slog.Logger) *BotWorker {
	return &BotWorker{
		pool:     pool,
		botToken: botToken,
		log:      log,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Start запускает long polling событий бота; без MAX_BOT_TOKEN — не стартует.
func (w *BotWorker) Start(ctx context.Context) {
	if w.botToken == "" {
		w.log.Warn("bot worker disabled: MAX_BOT_TOKEN is empty")
		return
	}
	go func() {
		w.log.Info("bot worker started (long polling)")
		for {
			if ctx.Err() != nil {
				return
			}
			if err := w.poll(ctx); err != nil {
				w.log.Error("bot poll failed", slog.Any("err", err))
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (w *BotWorker) poll(ctx context.Context) error {
	url := updatesURL
	if w.marker > 0 {
		url = fmt.Sprintf("%s?marker=%d&time=30", updatesURL, w.marker)
	} else {
		url += "?time=30"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", w.botToken)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("updates status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Updates []struct {
			UpdateType string `json:"update_type"`
			Payload    struct {
				CallbackID string `json:"callback_id"`
				Message    struct {
					Text      string `json:"text"`
					Timestamp int64  `json:"timestamp"`
					Recipient struct {
						ChatID int64 `json:"chat_id"`
					} `json:"recipient"`
				} `json:"message"`
				User struct {
					ID int64 `json:"id"`
				} `json:"user"`
			} `json:"payload"`
		} `json:"updates"`
		Marker json.Number `json:"marker"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode updates: %w", err)
	}

	if m, err := parsed.Marker.Int64(); err == nil && m > 0 {
		w.marker = m
	}

	for _, u := range parsed.Updates {
		switch u.UpdateType {
		case "message_created":
			// отвечаем только на свежие сообщения, чтобы не спамить при рестартах
			if u.Payload.Message.Timestamp > 0 {
				msgTime := time.UnixMilli(u.Payload.Message.Timestamp)
				if time.Since(msgTime) > 2*time.Minute {
					continue
				}
			}
			w.replyGreeting(ctx, u.Payload.Message.Recipient.ChatID)
		case "message_callback":
			w.answerCallback(ctx, u.Payload.CallbackID)
		}
	}
	return nil
}

type keyboardButton struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (w *BotWorker) replyGreeting(ctx context.Context, chatID int64) {
	message := map[string]any{
		"text": "Привет! 👋\n\nЯ помогаю находить работу по-новому: компании сами ищут тебя.\n\nЗаполни резюме в мини-приложении — и получай приглашения от компаний СПб.",
		"attachments": []map[string]any{
			{
				"type": "inline_keyboard",
				"payload": map[string]any{
					"buttons": [][]keyboardButton{
						{{Type: "open_app", Text: "🚀 Открыть мини-приложение"}},
						{{Type: "callback", Text: "❓ Как это работает"}},
					},
				},
			},
		},
		"notify": true,
	}
	if err := w.sendMessage(ctx, fmt.Sprintf("chat_id=%d", chatID), message); err != nil {
		w.log.Error("bot reply failed", slog.Int64("chat_id", chatID), slog.Any("err", err))
	}
}

func (w *BotWorker) answerCallback(ctx context.Context, callbackID string) {
	body, _ := json.Marshal(map[string]any{
		"text": "Открой мини-приложение кнопкой выше — там всё происходит 👆",
	})
	url := fmt.Sprintf("%s/answers?callback_id=%s", apiBaseURL, callbackID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", w.botToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		w.log.Error("callback answer failed", slog.Any("err", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.log.Error("callback answer status", slog.Int("status", resp.StatusCode), slog.String("body", strings.TrimSpace(string(body))[:min(80, len(body))]))
	}
}

func (w *BotWorker) sendMessage(ctx context.Context, query string, message any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/messages?%s", apiBaseURL, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", w.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("messages status %d: %s", resp.StatusCode, strings.TrimSpace(string(body))[:min(80, len(body))])
	}
	return nil
}
