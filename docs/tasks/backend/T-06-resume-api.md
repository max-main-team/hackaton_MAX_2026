# T-06 — Резюме: `GET/PUT /my/resume`

- **Статус:** ✅
- **Спека:** `docs/backend/spec-api.md` (раздел «Кандидат: резюме»)
- **Зависимости:** T-05
- **Разблокирует:** T-12 (опрос), T-15 (экран)

## Что сделать

- `internal/handler/resume.go`: оба эндпоинта в приватной группе.
- `GET /my/resume` — 200 или 404 `{"message":"resume not found"}`.
- `PUT /my/resume` — upsert; валидация: title непустой, `schedule` ∈
  `full|part|remote|hybrid`, `experience_months` 0..600; иначе 400.
- `updated_at = now()` при сохранении.

## Критерии приёмки

- [ ] PUT создаёт, повторный PUT обновляет (одна запись — UNIQUE user_id).
- [ ] `schedule: "night"` → 400; пустой title → 400.
- [ ] curl-сценарий из `docs/common/stage-1.md` (шаг 6) проходит.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
