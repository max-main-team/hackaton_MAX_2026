# Docs — документация проекта

Описывают фактическое состояние кода в репозитории. При изменении кода —
обновляйте соответствующий файл в этом же коммите.

## Структура

```
docs/
├── common/               — общее
│   ├── idea.md           — общая идея продукта
│   └── features.md       — список фич: основные и второстепенные
├── backend/
│   ├── architecture.md   — слои, поток запроса, конфиг, запуск
│   ├── api.md            — спецификация HTTP-эндпоинтов (текущее состояние)
│   ├── database.md       — схема БД, конвенции миграций
│   ├── spec-auth.md      — SPECS: initData-подпись, JWT, RequireAuth
│   ├── spec-domain.md    — SPECS: миграции 000002/000003, репозитории
│   ├── spec-api.md       — SPECS: полный контракт API v1
│   ├── spec-matching.md  — SPECS: скоринг, TTL, еженедельный опрос
│   └── spec-pdf-parse.md — SPECS: парсинг резюме из PDF
├── max/                  — выжимка документации MAX (dev.max.ru)
│   ├── README.md         — индекс + ключевые факты
│   ├── mini-app-setup.md — бот, привязка мини-аппа, диплинки
│   ├── bridge.md         — справочник window.WebApp
│   ├── init-data.md      — структура initData + алгоритм валидации + Go
│   └── bot-api.md        — Bot API: сообщения, кнопки, лимиты
└── frontend/
    ├── architecture.md   — структура, прокси, команды
    ├── max-bridge.md     — интеграция с MAX (window.WebApp, initData)
    └── spec-ui.md        — SPECS: роуты, экраны, компоненты, API-биндинги
```

Файлы `architecture.md`/`api.md`/`database.md` описывают код **как есть**;
файлы `spec-*.md` — целевое поведение фич. Живой контракт API — Swagger:
`https://eclipse-sim.ru/api/docs` (+ `DATA-API.yaml` и `openapi.yaml`
в корне репозитория). Архитектура каталогов —
[`ARCHITECTURE.md`](../ARCHITECTURE.md) в корне.

## Кратко о проекте

Мини-приложение реверс-найма для мессенджера MAX. Монорепозиторий:

| Папка      | Стек                                        |
|------------|---------------------------------------------|
| `backend/` | Go 1.27, Echo v4, pgx v5 (pgxpool), Postgres |
| `frontend/`| React 19, TypeScript, Vite 8, oxlint         |

Локально: Postgres в Docker торчит на **5433** (5432 занят локальным Postgres
разработчика), бекенд на `:8080`, фронт на `:5173` с прокси `/api`.
