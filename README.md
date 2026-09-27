# MAX Mini App

Мини-приложение для мессенджера MAX: монорепозиторий с бекендом и фронтендом.

## Стек

| Слой     | Технологии                                  |
|----------|---------------------------------------------|
| Бекенд   | Go 1.27, Echo, pgx/pgxpool, PostgreSQL      |
| Фронтенд | React 19, TypeScript, Vite                  |
| Инфра    | Docker Compose (Postgres), Dockerfile       |

## Структура

```
backend/
├── cmd/server/          # точка входа (запуск, graceful shutdown)
├── internal/
│   ├── config/          # конфиг из env / .env
│   ├── database/        # pgxpool + запуск миграций
│   ├── handler/         # HTTP-обработчики (health, auth)
│   ├── middleware/      # slog-логирование запросов
│   ├── repository/      # доступ к данным (pgx)
│   └── server/          # сборка Echo: middleware и роуты
├── migrations/          # SQL-миграции (embed, применяются при старте)
└── Dockerfile

frontend/
├── src/
│   ├── lib/max.ts       # типизация MAX Bridge (window.WebApp)
│   ├── lib/api.ts       # клиент бекенда
│   └── App.tsx          # пример авторизации через initData
├── Dockerfile           # node build → nginx
├── nginx.conf           # SPA + прокси /api → backend:8080
└── vite.config.ts       # dev-прокси /api → localhost:8080
```

## Запуск в проде (CI/CD)

Push в `main` триггерит `.github/workflows/ci-cd.yml`:

1. CI: бекенд (`build/vet/test/fmt`) + фронтенд (`lint/build`)
2. Сборка и пуш образов в GHCR (`hackaton_max_2026/backend|frontend:latest`)
3. Деплой на сервер по SSH: копирование `docker-compose.prod.yml` + `.env`
   (рендерится из GitHub Secrets), `docker compose pull && up -d`,
   проверка `GET /api/v1/health`

Сервер: `/opt/max-miniapp`. Секреты GitHub: `SSH_HOST`, `SSH_USER`,
`SSH_PRIVATE_KEY`, `POSTGRES_PASSWORD`, `JWT_SECRET`, `MAX_BOT_TOKEN`,
`GHCR_TOKEN` (PAT с `read:packages` для pull из GHCR).

⚠️ Мини-приложения MAX требуют `https` — для релиза нужен домен +
TLS-сертификат перед nginx (certbot) либо проксирующий сервис.

## Запуск в разработке

1. Поднять Postgres:

   ```bash
   docker compose up -d
   ```

2. Запустить бекенд (миграции применятся автоматически):

   ```bash
   cd backend
   cp .env.example .env   # при необходимости поправить
   make run
   ```

3. Запустить фронтенд (в другом терминале):

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Фронтенд доступен на http://localhost:5173, запросы `/api/*` проксируются на бекенд (`:8080`).

## API

Базовый префикс: `/api/v1`

| Метод | Путь      | Описание                                                    |
|-------|-----------|-------------------------------------------------------------|
| GET   | `/health` | Живость сервиса и БД                                        |
| POST  | `/auth`   | Авторизация по initData из MAX Bridge, upsert пользователя  |

### Пример

```bash
curl -X POST http://localhost:8080/api/v1/auth \
  -H 'Content-Type: application/json' \
  -d '{"initData": "user=%7B%22id%22%3A1%2C%22first_name%22%3A%22Ivan%22%7D&auth_date=1730000000"}'
```

## Интеграция с MAX

- **MAX Bridge** — глобальный объект `window.WebApp` (скрипт подключён в `frontend/index.html`).
  Типы и хелперы — `frontend/src/lib/max.ts`: `getInitData()`, `isInsideMax()`, `ready()`, `expand()` и т.д.
- `initData` отправляется на `POST /api/v1/auth`; бекенд создаёт/обновляет пользователя.
- **Важно:** перед продом нужно добавить проверку подписи initData токеном
  бота (`MAX_BOT_TOKEN`) — алгоритм и готовая Go-реализация:
  `docs/max/init-data.md`.
- Для стилизации под нативный UI MAX есть библиотека React-компонентов
  [`@maxhub/max-ui`](https://github.com/max-messenger/max-ui) — подключается по желанию.

## Полезные команды

| Команда                      | Где        | Что делает                    |
|------------------------------|------------|-------------------------------|
| `make run`                   | `backend/` | Запуск сервера                |
| `make build`                 | `backend/` | Сборка в `bin/server`         |
| `make test` / `make vet`     | `backend/` | Тесты / статический анализ    |
| `npm run dev`                | `frontend/`| Dev-сервер Vite               |
| `npm run build`              | `frontend/`| Прод-сборка в `dist/`         |
| `npm run lint`               | `frontend/`| oxlint                        |
