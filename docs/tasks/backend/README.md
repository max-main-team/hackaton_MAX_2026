# Задачи бекенда

Детали каждой задачи — в её файле. Статусы дублируются в
`docs/common/status.md`. Спецификации: `docs/backend/spec-*.md`.

## Этап 1 — домен и безопасность

| ID   | Задача | Статус | Зависит | Файл |
|------|--------|--------|---------|------|
| T-01 | `internal/auth`: initData (Parse/Verify, dev-режим), `JWT_SECRET` | ✅ | — | [T-01-auth-initdata.md](./T-01-auth-initdata.md) |
| T-02 | JWT: Issue/Parse + middleware `RequireAuth`, приватная группа | ✅ | T-01 | [T-02-jwt-requireauth.md](./T-02-jwt-requireauth.md) |
| T-03 | Новый ответ `/auth` `{token, user}` (⚠️ вместе с T-13) | ✅ | T-02 | [T-03-auth-response.md](./T-03-auth-response.md) |
| T-04 | Миграция 000003 (домен) + репозитории resume/company/vacancy | ✅ | — | [T-04-migrations-repos.md](./T-04-migrations-repos.md) |
| T-05 | `GET /me`, `POST /me/role` | ✅ | T-02, T-04 | [T-05-me-role.md](./T-05-me-role.md) |
| T-06 | Резюме: `GET/PUT /my/resume` | ✅ | T-05 | [T-06-resume-api.md](./T-06-resume-api.md) |
| T-07 | Компания и вакансии: create/list/patch | ✅ | T-05 | [T-07-company-vacancy-api.md](./T-07-company-vacancy-api.md) |

## Этап 2 — матчинг и взаимодействие

| ID   | Задача | Статус | Зависит | Файл |
|------|--------|--------|---------|------|
| T-08 | `internal/scoring` + тесты | ⬜ | T-04 | [T-08-scoring.md](./T-08-scoring.md) |
| T-09 | Выдача кандидатов `GET /vacancies/{id}/candidates` (list/feed) | ⬜ | T-07, T-08 | [T-09-candidates-api.md](./T-09-candidates-api.md) |
| T-10 | Действия рекрутера `POST .../action` (миграция 000003) | ⬜ | T-09 | [T-10-recruiter-action.md](./T-10-recruiter-action.md) |
| T-11 | Приглашения: `GET /my/invitations`, `POST /invitations/{id}/respond` | ⬜ | T-10 | [T-11-invitations.md](./T-11-invitations.md) |

## Этап 3 — уведомления и сиды

| ID   | Задача | Статус | Зависит | Файл |
|------|--------|--------|---------|------|
| T-12 | Воркер еженедельного опроса + `POST /my/resume/confirm-activity` | ⬜ | T-06 | [T-12-survey-worker.md](./T-12-survey-worker.md) |
| T-19 | Сидовые данные для демо | ⬜ | T-07 | [T-19-demo-seeds.md](./T-19-demo-seeds.md) |

## Новые фичи (верификация, рефералка, карта)

| ID | Задача | Статус | Зависит | Файл |
|----|--------|--------|---------|------|
| T-21 | Верификация компании через бота (verified=true) | ⬜ | T-07 | [T-21-company-verification.md](./T-21-company-verification.md) |
| T-23 | Рефералка: start_param, /my/referrals | ⬜ | T-03 | [T-23-referral-api.md](./T-23-referral-api.md) |
| T-25 | Карта: координаты городов, GET /vacancies/map | ⬜ | T-07 | [T-25-map-api.md](./T-25-map-api.md) |
| T-27 | AI-скоринг | ❄️ | T-08 | [T-27-ai-scoring.md](./T-27-ai-scoring.md) |
