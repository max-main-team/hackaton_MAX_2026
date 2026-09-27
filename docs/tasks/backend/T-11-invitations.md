# T-11 — Приглашения кандидата и TTL

- **Статус:** ✅
- **Спека:** `docs/backend/spec-api.md` (раздел «Приглашения»), `spec-matching.md` (раздел 2)
- **Зависимости:** T-10
- **Разблокирует:** T-18 (экран приглашений)

## Что сделать

- `GET /my/invitations` — только `action='invite'` по текущему кандидату;
  статус вычисляется на лету: `pending` (дедлайн в будущем),
  `overdue` (прошёл, нет ответа), `responded`; `hours_left`;
  company (verified!) и vacancy в ответе; сортировка pending→overdue→responded.
- `POST /invitations/{id}/respond` — `{"response": "accept"|"decline"}`;
  только своё (404/403); только из `pending` (иначе 409).
- `accept` → матч: ответ включает `recruiter_contact` (username рекрутера).
- Дедлайн: `created_at + vacancies.response_ttl_hours`, без хранения.

## Критерии приёмки

- [ ] Приглашение с прошедшим дедлайном → `overdue`, ответ → 409.
- [ ] В ответе присутствуют company.verified и vacancy.title.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
