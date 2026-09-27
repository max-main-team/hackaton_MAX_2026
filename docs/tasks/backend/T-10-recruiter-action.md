# T-10 — Действия рекрутера (миграция 000003)

- **Статус:** ⬜
- **Спека:** `docs/backend/spec-api.md`, `spec-domain.md` (000003)
- **Зависимости:** T-09
- **Разблокирует:** T-11, T-17

## Что сделать

- `migrations/000003_actions.up.sql` / `.down.sql`: `recruiter_actions`
  (UNIQUE vacancy_id+candidate_user_id), `candidate_responses`, индексы.
- `internal/repository/matching.go`: CreateAction, ExistsAction, Get,
  CreateResponse, ListByCandidate, ListByVacancy.
- `POST /vacancies/{id}/candidates/{candidateUserId}/action` —
  `{"action": "invite"|"skip"}`; повторный вызов → существующая запись (200).
- Проверки: членство (403), кандидат существует (404), action валиден (400).

## Критерии приёмки

- [ ] Повторный action на ту же пару не создаёт дубль (UNIQUE отрабатывает).
- [ ] Чужая вакансия → 403.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
