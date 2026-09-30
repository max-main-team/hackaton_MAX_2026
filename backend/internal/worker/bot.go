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

func (w *BotWorker) Start(ctx context.Context) {
	if w.botToken == "" {
		w.log.Warn("bot worker disabled: MAX_BOT_TOKEN is empty")
		return
	}
	w.registerCommands(ctx)
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

type botCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func botCommands() []botCommand {
	return []botCommand{
		{Name: "start", Description: "Приветствие и кнопка приложения"},
		{Name: "help", Description: "Как работает сервис"},
		{Name: "status", Description: "Мой статус: резюме и приглашения"},
	}
}

func (w *BotWorker) registerCommands(ctx context.Context) {
	body, _ := json.Marshal(map[string]any{"commands": botCommands()})
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, apiBaseURL+"/me/commands", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", w.botToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		w.log.Warn("register bot commands failed", slog.Any("err", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.log.Warn("register bot commands status", slog.Int("status", resp.StatusCode), slog.String("body", truncateForLog(body)))
		return
	}
	w.log.Info("bot commands registered")
}

func truncateForLog(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 80 {
		s = s[:80]
	}
	return s
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

	rawBody, _ := io.ReadAll(resp.Body)
	w.log.Debug("bot poll raw", slog.Int("len", len(rawBody)), slog.String("body", string(rawBody)))

	var parsed struct {
		Updates []struct {
			UpdateType string `json:"update_type"`
			Payload    struct {
				CallbackID string `json:"callback_id"`
				Chat       struct {
					ChatID int64 `json:"chat_id"`
				} `json:"chat"`
				User struct {
					ID        int64  `json:"id"`
					FirstName string `json:"first_name"`
				} `json:"user"`
				Message struct {
					Text      string `json:"text"`
					Timestamp int64  `json:"timestamp"`
					Recipient struct {
						ChatID int64 `json:"chat_id"`
					} `json:"recipient"`
				} `json:"message"`
			} `json:"payload"`
		} `json:"updates"`
		Marker json.Number `json:"marker"`
	}
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return fmt.Errorf("decode updates: %w", err)
	}

	if m, err := parsed.Marker.Int64(); err == nil && m > 0 {
		w.marker = m
	}
	w.log.Info("bot poll", slog.Int("updates", len(parsed.Updates)), slog.Int64("marker", w.marker))
	for _, u := range parsed.Updates {
		w.log.Info("bot update",
			slog.String("type", u.UpdateType),
			slog.Int64("ts", u.Payload.Message.Timestamp),
			slog.Int64("chat_id", u.Payload.Message.Recipient.ChatID),
			slog.String("text", u.Payload.Message.Text))
	}

	for _, u := range parsed.Updates {
		switch u.UpdateType {
		case "bot_started":
			if uid := u.Payload.User.ID; uid != 0 {
				if err := w.sendMessageToUser(ctx, uid, greetingMessage()); err != nil {
					w.log.Error("bot_started greeting failed", slog.Int64("user_id", uid), slog.Any("err", err))
				} else {
					w.log.Info("bot_started greeting sent", slog.Int64("user_id", uid))
				}
			}
		case "message_created":
			if ts := u.Payload.Message.Timestamp; ts > 0 {
				var msgTime time.Time
				if ts > 1_000_000_000_000 {
					msgTime = time.UnixMilli(ts)
				} else {
					msgTime = time.Unix(ts, 0)
				}
				if time.Since(msgTime) > 2*time.Minute {
					continue
				}
			}
			w.handleMessage(ctx, u.Payload.Message.Recipient.ChatID, u.Payload.User.ID, u.Payload.Message.Text)
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

type botUserStatus struct {
	Known        bool
	Role         string
	FirstName    string
	ResumeTitle  string
	ResumeActive bool
	Pending      int
	CompanyName  string
	Verified     bool
	InviteQuota  int
	InviteUsed   int
}

func (w *BotWorker) handleMessage(ctx context.Context, chatID, userID int64, text string) {
	if err := w.replyCommand(ctx, chatID, userID, text); err != nil {
		w.log.Error("bot reply failed", slog.Int64("chat_id", chatID), slog.Any("err", err))
	} else {
		w.log.Info("bot reply sent", slog.Int64("chat_id", chatID), slog.String("text", text))
	}
}

func routeCommand(text string) string {
	cmd := strings.ToLower(strings.TrimSpace(text))
	cmd = strings.TrimPrefix(cmd, "/")
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return "unknown"
	}
	switch fields[0] {
	case "start":
		return "start"
	case "help", "помощь", "как":
		return "help"
	case "status", "статус":
		return "status"
	default:
		return "unknown"
	}
}

func (w *BotWorker) replyCommand(ctx context.Context, chatID, userID int64, text string) error {
	var msg map[string]any
	switch routeCommand(text) {
	case "start":
		msg = greetingMessage()
	case "help":
		msg = helpMessage()
	case "status":
		status := w.fetchUserStatus(ctx, userID)
		msg = statusMessage(status)
	default:
		msg = fallbackMessage()
	}
	return w.sendMessage(ctx, fmt.Sprintf("chat_id=%d", chatID), msg)
}

func (w *BotWorker) fetchUserStatus(ctx context.Context, userID int64) botUserStatus {
	var s botUserStatus
	err := w.pool.QueryRow(ctx, `
		SELECT u.id IS NOT NULL, COALESCE(u.role, ''), COALESCE(u.first_name, ''),
		       COALESCE(r.title, ''), COALESCE(r.is_active, FALSE),
		       (SELECT count(*) FROM recruiter_actions ra
		        WHERE ra.candidate_user_id = u.id AND ra.action = 'invite'
		          AND NOT EXISTS (SELECT 1 FROM candidate_responses cr WHERE cr.action_id = ra.id)),
		       COALESCE(c.name, ''), COALESCE(c.verified, FALSE),
		       COALESCE(c.invite_quota, 0), COALESCE(c.invite_used, 0)
		FROM users u
		LEFT JOIN resumes r ON r.user_id = u.id
		LEFT JOIN company_members cm ON cm.user_id = u.id
		LEFT JOIN companies c ON c.id = cm.company_id
		WHERE u.id = $1
		LIMIT 1
	`, userID).Scan(
		&s.Known, &s.Role, &s.FirstName,
		&s.ResumeTitle, &s.ResumeActive,
		&s.Pending,
		&s.CompanyName, &s.Verified,
		&s.InviteQuota, &s.InviteUsed,
	)
	if err != nil {
		return botUserStatus{}
	}
	return s
}

func statusMessage(s botUserStatus) map[string]any {
	var b strings.Builder
	switch {
	case !s.Known:
		b.WriteString("Я вас ещё не видел в мини-приложении 🙈\n\nОткройте его, выберите роль — и я покажу статус здесь.")
	case s.Role == "candidate":
		b.WriteString("Ваш статус 👤\n\n")
		if s.ResumeTitle == "" {
			b.WriteString("Резюме пока не заполнено.")
		} else if s.ResumeActive {
			b.WriteString(fmt.Sprintf("Резюме «%s» активно — компании его видят.", s.ResumeTitle))
		} else {
			b.WriteString(fmt.Sprintf("Резюме «%s» на паузе — компании его не видят.", s.ResumeTitle))
		}
		if s.Pending > 0 {
			b.WriteString(fmt.Sprintf("\n\nПриглашений ждёт ответа: %d.", s.Pending))
		}
	default:
		b.WriteString("Ваш статус 🏢\n\n")
		if s.CompanyName == "" {
			b.WriteString("Компания ещё не создана — откройте мини-приложение и добавьте её.")
		} else {
			verified := "нет"
			if s.Verified {
				verified = "да"
			}
			b.WriteString(fmt.Sprintf("Компания: %s (верифицирована: %s).\nПриглашений использовано: %d из %d.",
				s.CompanyName, verified, s.InviteUsed, s.InviteQuota))
		}
	}
	return keyboardMessage(b.String())
}

func helpMessage() map[string]any {
	return keyboardMessage("Как это работает 💡\n\n" +
		"Это реверс-найм: не вы откликаетесь на вакансии, а компании находят вас.\n\n" +
		"1. Заполните резюме в мини-приложении (можно загрузить PDF — распарсим сами).\n" +
		"2. Компании видят вас в подборке под свои вакансии.\n" +
		"3. Придёт приглашение с дедлайном — примите или отклоните.\n" +
		"4. При согласии открывается матч и контакт рекрутера.\n\n" +
		"Команды: /status — мой статус, /help — эта справка.")
}

func fallbackMessage() map[string]any {
	return keyboardMessage("Я на связи! Напишите /help — расскажу, как всё устроено, или /status — покажу ваш статус.")
}

func keyboardMessage(text string) map[string]any {
	return map[string]any{
		"text": text,
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
}

func greetingMessage() map[string]any {
	return keyboardMessage("Привет! 👋\n\nЯ помогаю находить работу по-новому: компании сами ищут тебя.\n\nЗаполни резюме в мини-приложении — и получай приглашения от компаний СПб.")
}

func (w *BotWorker) sendMessageToUser(ctx context.Context, userID int64, message map[string]any) error {
	return w.sendMessage(ctx, fmt.Sprintf("user_id=%d", userID), message)
}

func (w *BotWorker) answerCallback(ctx context.Context, callbackID string) {
	body, _ := json.Marshal(map[string]any{
		"text": "Реверс-найм: вы заполняете резюме один раз, а компании сами находят вас и присылают приглашения. Кнопка ниже откроет мини-приложение 👇",
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
		w.log.Error("callback answer status", slog.Int("status", resp.StatusCode), slog.String("body", truncateForLog(body)))
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
		return fmt.Errorf("messages status %d: %s", resp.StatusCode, truncateForLog(body))
	}
	return nil
}
