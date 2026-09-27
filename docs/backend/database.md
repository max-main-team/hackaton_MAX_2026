# Backend — база данных

## Подключение

- Сервер: Postgres 17 (docker-compose из корня, хост-порт **5433**).
- БД: `maxapp`, пользователь/пароль: `postgres/postgres` (только локально).
- Доступ из кода: `pgxpool` (`internal/database/postgres.go`).

## Схема (текущая)

`users` — миграция `000001_init`:

| Колонка    | Тип         | Заметка                          |
|------------|-------------|----------------------------------|
| id         | BIGINT PK   | ID пользователя из MAX           |
| username   | TEXT NOT NULL DEFAULT '' |                  |
| first_name | TEXT NOT NULL DEFAULT '' |                  |
| last_name  | TEXT NOT NULL DEFAULT '' |                  |
| photo_url  | TEXT NOT NULL DEFAULT '' |                  |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT now()       |
| updated_at | TIMESTAMPTZ NOT NULL DEFAULT now()       |

Служебная `schema_migrations (version BIGINT PK, name TEXT, applied_at)` —
создаётся раннером.

## Миграции

- Файлы: `backend/migrations/000001_name.up.sql` и `.down.sql`.
- Встраиваются в бинарник (`embed.go`, директива `//go:embed *.sql`).
- Применяются автоматически при старте сервера (`database.Migrate`),
  каждая в транзакции вместе с записью в `schema_migrations`.
- `down`-миграции автоматически **не** выполняются — хранятся для отката
  вручную.
- Пропущенные/кривые имена → сервер не стартует (ошибка парсинга имени).

### Как добавить миграцию

1. Следующий номер: `000002_<краткое_имя>.up.sql` + `.down.sql`.
2. Up — только аддитивные изменения (новая таблица/колонка с дефолтом),
   чтобы деплой не ломал старую версию бинаря.
3. Закоммитить оба файла в одном коммите с изменениями кода.
