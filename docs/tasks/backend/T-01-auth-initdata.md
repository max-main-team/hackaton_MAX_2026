# T-01 — Пакет `internal/auth`: initData

- **Статус:** ✅
- **Спека:** `docs/backend/spec-auth.md` (разделы 1–2)
- **Зависимости:** —
- **Разблокирует:** T-02

## Что сделать

- `internal/auth/initdata.go`: `ParseInitData(raw)` (user + auth_date),
  `Verify(raw, appSecret)` — HMAC по документации MAX; отсечка
  `auth_date` старше 24 часов.
- Dev-режим: `MAX_BOT_TOKEN` пуст → проверка подписи пропускается,
  warning в логе при старте; секрет задан → проверка обязательна.
- `config.Load()`: новый обязательный `JWTSecret` (паника при пустом).
- Обновить `backend/.env.example`.

## Критерии приёмки

- [ ] Без `JWT_SECRET` сервер не стартует с понятной ошибкой.
- [ ] `ParseInitData` возвращает пользователя и auth_date; протухший → ошибка.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
