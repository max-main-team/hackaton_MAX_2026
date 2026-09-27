# T-02 — JWT и middleware `RequireAuth`

- **Статус:** ⬜
- **Спека:** `docs/backend/spec-auth.md` (разделы 2, 4)
- **Зависимости:** T-01
- **Разблокирует:** T-03, T-05

## Что сделать

- `internal/auth/jwt.go`: `Issue(userID, secret, now)` (HS256, sub/iat/exp=+168h),
  `Parse(token, secret) (int64, error)`.
- `internal/auth/jwt_test.go`: roundtrip; битая подпись и просроченный → ошибка.
- `internal/middleware/auth.go`: `RequireAuth(secret)` — Bearer, 401 при
  ошибке, `c.Set("user_id", id)`; хелпер `UserIDFromContext`.
- `internal/server/server.go`: приватная группа `private.Use(RequireAuth(...))`.

## Критерии приёмки

- [ ] Тест roundtrip зелёный (`go test ./internal/auth/`).
- [ ] Запрос без/с битым/с просроченным токеном к приватной ручке → 401.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
