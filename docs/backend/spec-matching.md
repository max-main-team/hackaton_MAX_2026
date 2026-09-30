# Spec: скоринг, TTL, еженедельные уведомления


## 1. Скоринг кандидата под вакансию (фича 1.3, алгоритмическая часть)

Файл: `backend/internal/scoring/scoring.go` (чистая функция, легко покрыть тестами).

```go
type Vacancy struct {
    RequiredSkills      string
    MinExperienceMonths int
    City                string
    Schedule            string
}

type Resume struct {
    Skills            string
    ExperienceMonths  int
    City              string
    Schedule          string
}

// Score возвращает 0..100 и разложение по компонентам.
func Score(v Vacancy, r Resume) (int, Breakdown)
```

Формула (все компоненты 0..1):

| Компонент   | Вес  | Расчёт                                                                                   |
|-------------|------|------------------------------------------------------------------------------------------|
| `skills`    | 0.5  | Доля токенов `required_skills`, найденных в `resume.skills`. Токены: lowercase, split по запятой, trim. Пустые требования → 1 |
| `city`      | 0.2  | Равны (lowercase) → 1; одно из полей пустое → 0.5; иначе 0                                |
| `schedule`  | 0.2  | Равны → 1; иначе 0                                                                        |
| `experience`| 0.1  | `min(exp / max(min_exp,1), 1)`; `min_exp == 0` → 1                                        |

Итог: `round(100 * Σ вес_i * компонент_i)`.

- Токенизация навыков — общий хелпер `splitSkills(s string) []string`
  (использовать и для вакансий, и для резюме).
- Тесты `scoring_test.go`: полное совпадение = 100, пустое резюме =
  ожидаемое значение по формуле, совпадение 2 навыков из 4.

## 2. TTL на ответ рекрутера (фича 1.2)

- Источник TTL: `vacancies.response_ttl_hours`.
- Дедлайн приглашения: `recruiter_actions.created_at + ttl` — хранить не надо,
  считать в SQL/Go при выдаче (`GET /my/invitations`, `GET /matches`).
- Статусы приглашения (см. spec-api.md): `pending | overdue | responded`.
- Cron не нужен: статус вычисляется при чтении. Кнопки «ответить» на фронте
  дизейблятся при `overdue`, бек дополнительно возвращает 409.

## 3. Еженедельный опрос кандидата (фича 1.1)

Файл: `backend/internal/worker/survey.go`, запускается из `main.go`
горутиной с `time.Ticker` (интервал — 1 час).

Логика тика:

1. `SELECT ... FROM resumes JOIN users ON users.id = resumes.user_id
   WHERE resumes.is_active = TRUE
     AND resumes.last_confirmed_at < now() - interval '7 days'`
2. Для каждого — отправить сообщение в MAX через Bot API
   (`MAX_BOT_TOKEN` в env): текст «Ваш подбор ещё актуален?» + deeplink
   на мини-апп: `https://max.ru/?startapp=confirm` (ссылку уточнить по
   документации MAX Bot API при интеграции).
3. Повторные отправки не спамить: служебная таблица не нужна для MVP —
   отправка идёт пока пользователь не подтвердит (подтверждение обновляет
   `last_confirmed_at` и выпадает из выборки). Максимум 1 сообщение в тик.

Отказоустойчивость: ошибка отправки одному пользователю не прерывает
рассылку; ошибки — в лог. Если `MAX_BOT_TOKEN` пуст — воркер не стартует,
в лог пишется warning (dev-режим).

Новые env: `MAX_BOT_TOKEN` (опционален в dev).

## Критерии приёмки

- [ ] `Score` покрыт тестами, `go test ./internal/scoring/` зелёный.
- [ ] Кандидат с подходящими навыками стоит выше неподходящего при
      равных прочих.
- [ ] `pending`-приглашение с прошедшим дедлайном отдаётся как `overdue`,
      ответ на него → 409.
- [ ] Воркер не отправляет ничего при пустом `MAX_BOT_TOKEN`.
- [ ] Подтверждение активности сбрасывает выборку рассылки (SQL-проверка).
