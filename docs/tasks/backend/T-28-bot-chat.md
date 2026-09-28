# T-28 — Логика бота в MAX (чат)

- **Статус:** ⬜
- **Спека:** docs/max/bot-api.md (GET /updates, POST /messages, POST /answers)
- **Зависимости:** T-12 (общий MAX_BOT_TOKEN)

## Что сделать

- `internal/worker/bot.go`: long polling `GET /updates` (marker + time=30).
- `message_created` → приветствие + inline-клавиатура:
  кнопка `open_app` («Открыть мини-приложение») и callback «Как это работает».
- `message_callback` → `POST /answers` с пояснением.
- Анти-спам: отвечать только на сообщения свежее 2 минут (без спама при рестартах).
- Воркер стартует из main.go вместе с survey-воркером.

## Критерии приёмки

- [ ] Без MAX_BOT_TOKEN воркер не стартует (warning).
- [ ] `GET /updates` с токеном отвечает 200 (проверено curl).
- [ ] go build/vet/gofmt — чисто.
