# T-07 — Компания и вакансии

- **Статус:** ✅
- **Спека:** `docs/backend/spec-api.md` (раздел «Рекрутер: компания и вакансии»)
- **Зависимости:** T-05
- **Разблокирует:** T-09, T-16 (экраны), T-19 (сиды)

## Что сделать

- `internal/handler/company.go`, `internal/handler/vacancy.go`.
- `POST /companies` — создать + добавить создателя в company_members.
- `GET /my/companies` — компании текущего пользователя.
- `POST /companies/{id}/vacancies` — членство (иначе 403); валидация:
  title, schedule, `response_ttl_hours` 1..336.
- `GET /companies/{id}/vacancies` — с проверкой членства.
- `PATCH /vacancies/{id}` — частичное обновление ttl/is_active.

## Критерии приёмки

- [ ] Вакансию в чужой компании → 403.
- [ ] `response_ttl_hours: 0` → 400.
- [ ] PATCH меняет только переданные поля.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
