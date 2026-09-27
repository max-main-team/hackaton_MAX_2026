# T-05 — Профиль: `GET /me`, `POST /me/role`

- **Статус:** ✅
- **Спека:** `docs/backend/spec-api.md` (раздел «Аутентификация и профиль»)
- **Зависимости:** T-02, T-04
- **Разблокирует:** T-06, T-07

## Что сделать

- `internal/handler/me.go`: оба эндпоинта в приватной группе.
- `GET /me` — пользователь по `user_id` из токена + `personal_data_accepted_at`
  (из `user_consents` через `GetConsentAcceptedAt`).
- `POST /me/role` — тело `{"role": "candidate"|"recruiter", "acceptPersonalData": true}`;
  невалидная роль или `acceptPersonalData != true` → 400;
  `UPDATE users SET role` + `UpsertConsent(user_id, 'personal_data')`.

## Критерии приёмки

- [ ] Без токена → 401; валидный токен → 200 с пользователем и `personal_data_accepted_at`.
- [ ] `role: "admin"` → 400; без `acceptPersonalData` → 400.
- [ ] Повторный вызов обновляет `accepted_at`, дубля в `user_consents` нет (UNIQUE).
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
