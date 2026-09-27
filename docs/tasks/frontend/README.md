# Задачи фронтенда

Детали каждой задачи — в её файле. Статусы дублируются в
`docs/common/status.md`. Спецификация экранов: `docs/frontend/spec-ui.md`.

## Этап 1 — каркас и базовые экраны

| ID   | Задача | Статус | Зависит | Файл |
|------|--------|--------|---------|------|
| T-13 | Роутинг, `session.ts`, api-клиенты с Bearer (⚠️ вместе с T-03 бека) | ⬜ | T-03 | [T-13-routing-session.md](./T-13-routing-session.md) |
| T-14 | Логин + онбординг выбора роли, гарды | ⬜ | T-13 | [T-14-login-onboarding.md](./T-14-login-onboarding.md) |
| T-15 | Экран резюме кандидата | ⬜ | T-14 | [T-15-resume-screen.md](./T-15-resume-screen.md) |
| T-16 | Рекрутер: компания → вакансия (TTL) → список вакансий | ⬜ | T-14 | [T-16-company-vacancy-screens.md](./T-16-company-vacancy-screens.md) |

## Этап 2 — матчинг UI

| ID   | Задача | Статус | Зависит | Файл |
|------|--------|--------|---------|------|
| T-17 | Свайп-лента + список кандидатов (`UserCard`, `Badge`) | ⬜ | T-16 | [T-17-feed-list.md](./T-17-feed-list.md) |
| T-18 | Приглашения: `Countdown`, ответ, матч-экран | ⬜ | T-17 | [T-18-invitations-screen.md](./T-18-invitations-screen.md) |
