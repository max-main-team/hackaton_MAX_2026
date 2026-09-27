# Тикеты MVP (порядок реализации)

Каждый тикет — самостоятельный коммит (или небольшая серия). Перед сдачей:
бек — `go build ./... && go vet ./... && gofmt -l .`, фронт —
`npm run build && npm run lint`. Спецификации лежат в `docs/backend/spec-*.md`
и `docs/frontend/spec-ui.md`.

## Бекенд

| ID   | Тикет | Спека | Зависит от |
|------|-------|-------|------------|
| T-01 | `internal/auth`: ParseInitData + Verify (dev-режим без секрета), env `JWT_SECRET` (обязательный) и `MAX_APP_SECRET_KEY` | spec-auth.md | — |
| T-02 | JWT: генерация в `/auth`, middleware `RequireAuth`, приватная группа роутов | spec-auth.md | T-01 |
| T-03 | Обновить ответ `/auth` (`{token, user}`), фронт не ломать — договориться в одном коммите с T-15 | spec-auth.md | T-02 |
| T-04 | Миграции 000002/000003 + репозитории (company, resume, vacancy, matching) | spec-domain.md | — |
| T-05 | Профиль: `GET /me`, `POST /me/role` | spec-api.md | T-02, T-04 |
| T-06 | Резюме: `GET/PUT /my/resume` | spec-api.md | T-05 |
| T-07 | Компания и вакансии: `POST /companies`, `GET /my/companies`, `POST /companies/{id}/vacancies`, `GET .../vacancies`, `PATCH /vacancies/{id}` | spec-api.md | T-05 |
| T-08 | `internal/scoring` + юнит-тесты | spec-matching.md | — |
| T-09 | Выдача кандидатов: `GET /vacancies/{id}/candidates` (list/feed) | spec-api.md | T-07, T-08 |
| T-10 | Действия рекрутера: `POST .../action` | spec-api.md | T-09 |
| T-11 | Приглашения кандидата: `GET /my/invitations`, `POST /invitations/{id}/respond` (со статусами pending/overdue/responded) | spec-api.md, spec-matching.md | T-10 |
| T-12 | Воркер еженедельного опроса + `POST /my/resume/confirm-activity`, env `MAX_BOT_TOKEN` | spec-matching.md | T-06 |

## Фронтенд

| ID   | Тикет | Спека | Зависит от |
|------|-------|-------|------------|
| T-13 | `react-router-dom`, структура `routes/`, `lib/session.ts`, api-клиенты с Bearer | spec-ui.md | T-02, T-03 |
| T-14 | Логин + онбординг выбора роли, гард-редиректы | spec-ui.md | T-13 |
| T-15 | Экран резюме кандидата | spec-ui.md | T-14 |
| T-16 | Рекрутер: компания → вакансия (включая TTL-селект) → список вакансий | spec-ui.md | T-14 |
| T-17 | Лента свайпов + список кандидатов + `UserCard`/`Badge` | spec-ui.md | T-16 |
| T-18 | Приглашения кандидата: `Countdown`, ответ, матч-экран с контактом | spec-ui.md | T-17 |

## Финализация

| ID   | Тикет |
|------|-------|
| T-19 | Сидовые данные для демо (SQL-скрипт или миграция 000004 с демо-компаниями/вакансиями — пометить, что для прода не включать) |
| T-20 | Демо-прогон полного сценария из spec-ui.md «Критерии приёмки», фикс багов |

## Параллельность

- Бек: T-01 → T-02 → (T-03) идут цепочкой; T-04 и T-08 параллельно сразу.
- Фронт может стартовать после T-03: T-13 → T-14, дальше T-15 ∥ T-16.
- Критический путь: T-01 → T-02 → T-03 → T-13 → T-14 → T-17 → T-18 → T-20.
