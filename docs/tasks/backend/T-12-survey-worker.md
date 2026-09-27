# T-12 — Воркер еженедельного опроса

- **Статус:** ⬜
- **Спека:** `docs/backend/spec-matching.md` (раздел 3)
- **Зависимости:** T-06
- **Разблокирует:** демонстрацию фичи 1.1

## Что сделать

- `internal/worker/survey.go`: горутина с `time.Ticker` (1 час), запуск из `main.go`.
- Выборка: `resumes.is_active = TRUE AND last_confirmed_at < now() - 7 days`.
- Отправка через MAX Bot API (env `MAX_BOT_TOKEN`): сообщение
  «Подбор ещё актуален?» + deeplink в мини-апп (экран подтверждения).
  Точный метод Bot API сверить с документацией MAX при интеграции.
- `MAX_BOT_TOKEN` пуст → воркер не стартует, warning.
- `POST /my/resume/confirm-activity` — `{"active": true|false}`:
  true → `last_confirmed_at=now(), is_active=true`; false → `is_active=false`.
- Ошибка отправки одному пользователю не прерывает рассылку.

## Критерии приёмки

- [ ] Пустой токен → в логе warning, рассылки нет.
- [ ] После `confirm-activity` пользователь выпадает из выборки (SQL-проверка).
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
