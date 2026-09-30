# Архитектура репозитория

Описание всех каталогов репозитория и правил, куда что класть.
Файл обновляется в том же коммите, что и изменение структуры.

## Дерево каталогов

```
hackaton_MAX_2026/
├── ARCHITECTURE.md            ← этот файл
├── README.md                  — быстрый старт: запуск, команды
├── docker-compose.yml         — Postgres 17 для локальной разработки (порт 5433)
├── docker-compose.prod.yml    — прод-стек: postgres + backend + frontend (nginx :80)
│
├── .github/workflows/ci-cd.yml — CI/CD: тесты → образы в GHCR → деплой по SSH
│
├── backend/                   — Go 1.27 + Echo v4 + pgx v5 (pgxpool)
│   ├── cmd/server/            — точка входа: main.go (config → db → migrate → echo)
│   ├── internal/
│   │   ├── config/            — env-конфиг (Load)
│   │   ├── database/          — pgxpool (postgres.go), раннер миграций (migrate.go)
│   │   ├── auth/              — [этап 1] initData, JWT
│   │   ├── scoring/           — [этап 2] скоринг кандидат↔вакансия
│   │   ├── worker/            — [этап 3] фоновые задачи (еженедельный опрос)
│   │   ├── handler/           — HTTP-обработчики Echo
│   │   ├── middleware/        — slog-логирование, RequireAuth [этап 1]
│   │   ├── repository/        — SQL-слой: user, resume, company, vacancy, matching
│   │   └── server/            — сборка Echo: middleware, роуты, Start/Shutdown
│   ├── migrations/            — *.sql, embed; применяются при старте сервера
│   ├── Dockerfile             — multi-stage образ бекенда
│   ├── .env.example, Makefile
│
├── frontend/                  — React 19 + TypeScript + Vite 8
│   ├── index.html             — подключает скрипт MAX Bridge
│   ├── Dockerfile             — node build → nginx (SPA + proxy /api)
│   ├── nginx.conf             — конфиг nginx для прод-образа
│   └── src/
│       ├── main.tsx           — точка входа, RouterProvider [этап 1]
│       ├── routes/            — экраны по ролям: login, onboarding, candidate/, recruiter/
│       ├── api/               — клиенты бекенда (client, auth, resume, company, matching)
│       ├── lib/               — max.ts (MAX Bridge), session.ts (JWT) [этап 1]
│       └── components/        — переиспользуемые UI-компоненты
│
└── docs/                      — вся документация
    ├── README.md              — индекс docs
    ├── common/                — общее
    │   ├── idea.md            — идея продукта (проблема → решение → ЦА)
    │   └── features.md        — список фич: основные и второстепенные
    ├── backend/               — доки бекенда
    │   ├── architecture.md    — как есть: слои, конфиг, запуск
    │   ├── api.md, database.md — как есть: текущие ручки и схема БД
    │   └── spec-*.md          — спецификации фич
    └── frontend/              — доки фронтенда (architecture, max-bridge, spec-ui)
```

## Потоки данных

```
MAX клиент → мини-апп (frontend, window.WebApp.initData)
    → fetch /api/* (в dev: vite proxy :5173 → :8080)
    → Echo middleware (logger → recover → CORS → RequireAuth)
    → handler → repository → pgxpool → Postgres (docker, :5433)
```

Авторизация: `POST /auth` с initData → JWT → `Authorization: Bearer` во
всех остальных запросах.

## Правила

| Что делаем                            | Куда кладём                                   |
|---------------------------------------|-----------------------------------------------|
| Новый код бекенда                     | `backend/internal/<слой>/`                    |
| Изменение схемы БД                    | `backend/migrations/0000NN_*.up.sql` + `.down.sql` |
| Новый экран/роут фронта               | `frontend/src/routes/<роль>/`                 |
| Факт о текущем коде                   | `docs/{backend,frontend}/*.md` (не спеки)     |
| Целевое поведение фичи                | `docs/{backend,frontend}/spec-*.md`           |
| Изменил структуру каталогов           | обнови этот файл                              |
