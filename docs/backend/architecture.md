# Backend — архитектура

## Стек

- Go 1.27 (`go.mod` → `max-miniapp/backend`)
- Echo v4 — HTTP-роутер и middleware
- pgx v5 + pgxpool — пул соединений с Postgres
- godotenv — загрузка `.env` (не обязателен, прод читает чистый env)
- Миграции — собственный раннер на `embed.FS` (без внешних зависимостей)

## Слои

```
cmd/server/main.go        запуск: config → pgxpool → migrate → echo → shutdown
internal/config/          Config struct из env
internal/database/        postgres.go (Connect), migrate.go (Migrate)
internal/server/          сборка Echo: middleware + роуты
internal/middleware/      SlogLogger — логирование запросов через slog
internal/handler/         HTTP-обработчики (health, auth)
internal/repository/      доступ к данным (SQL, pgx)
migrations/               *.sql, встроены в бинарник через embed
```

Поток запроса: `Echo → middleware (logger → recover → CORS) → handler →
repository → pgxpool → Postgres`. Слой `service` отсутствует — бизнес-логика
живёт в handler'ах; при росте вынести в `internal/service/`.

## Конфигурация (env)

| Переменная          | Дефолт                                    | Назначение                          |
|---------------------|-------------------------------------------|-------------------------------------|
| `ENV`               | `dev`                                     | `prod` включает JSON-логи           |
| `ADDR`              | `:8080`                                   | Адрес HTTP-сервера                  |
| `DATABASE_URL`      | `postgres://...@localhost:5433/maxapp`    | Строка подключения                  |
| `MAX_BOT_TOKEN`| пусто                                     | Секрет приложения MAX (см. api.md)  |

Пул: `MaxConns=10`, `MinConns=2`, `MaxConnLifetime=1h`. Graceful shutdown —
10 секунд (SIGINT/SIGTERM).

## Запуск

```bash
docker compose up -d        # из корня репо, Postgres на 5433
cd backend && make run      # миграции применяются при старте автоматически
```

## Конвенции

- Ошибки оборачиваются через `fmt.Errorf("...: %w", err)`.
- Handler возвращает `echo.NewHTTPError(code, message)` — сообщения клиенту
  на английском, без деталей внутренних ошибок (они уходят в лог).
- Логи: `slog`, текстовый формат в dev; запросы логирует только SlogLogger.
- Новый эндпоинт: handler в `internal/handler/`, регистрация в
  `internal/server/server.go` (`setupRoutes`), группа `/api/v1`.
- Комментарии в коде не пишем — документация живёт в `docs/`.
