# T-08 — Скоринг `internal/scoring`

- **Статус:** ✅
- **Спека:** `docs/backend/spec-matching.md` (раздел 1)
- **Зависимости:** T-04
- **Разблокирует:** T-09

## Что сделать

- `internal/scoring/scoring.go`: `Score(v Vacancy, r Resume) (int, Breakdown)`.
- Формула: skills 0.5 (доля токенов required_skills в resume.skills,
  lowercase/trim/запятая; пустые требования → 1), city 0.2 (равны 1 /
  одно пустое 0.5 / разные 0), schedule 0.2 (равны 1 / иначе 0),
  experience 0.1 (`min(exp/max(min_exp,1),1)`, min_exp=0 → 1).
- Итог: `round(100 * Σ вес·компонент)`; общий хелпер `splitSkills`.
- `scoring_test.go`: совпадение = 100; пустое резюме — по формуле;
  2 навыка из 4.

## Критерии приёмки

- [ ] `go test ./internal/scoring/` зелёный.
- [ ] Диапазон результата 0..100, `Breakdown` в сумме даёт итог.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
