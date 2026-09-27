# T-04 — Миграция 000002 и репозитории

- **Статус:** ⬜
- **Спека:** `docs/backend/spec-domain.md`
- **Зависимости:** —
- **Разблокирует:** T-05, T-08

## Что сделать

- `migrations/000002_domain.up.sql` / `.down.sql`: `users.role`,
  `companies`, `company_members`, `resumes`, `vacancies`, индексы
  `idx_vacancies_company`, `idx_resumes_active`.
  (000003/recruiter_actions — НЕ в этой задаче, это T-10.)
- `internal/repository/resume.go`: Upsert, GetByUserID, ListActive,
  TouchConfirmedAt.
- `internal/repository/company.go`: Create, GetByID, AddMember, ListByUser,
  ListVacancies.
- `internal/repository/vacancy.go`: Create, GetByID, Update (ttl/is_active),
  ListByCompany.

## Критерии приёмки

- [ ] Свежая БД: сервер стартует, 000002 применяется; повторный старт — идемпотентен.
- [ ] `.down.sql` откатывает изменения (проверено psql вручную).
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
