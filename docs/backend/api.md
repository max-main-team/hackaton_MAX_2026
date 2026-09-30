# Backend — API

Базовый префикс: `/api/v1`. Формат — JSON. Ошибки —
`{"message": "..."}` (стандартный формат Echo).

## GET /api/v1/health

Живость сервиса и БД.

Ответ `200`:
```json
{"status": "ok", "db": "ok"}
```

Если пинг БД не прошёл — `503` и `"db": "unavailable"`.

## POST /api/v1/auth

Авторизация пользователя мини-приложения. Тело:

```json
{"initData": "user=%7B%22id%22%3A42%2C%22first_name%22%3A%22Ivan%22%7D&auth_date=1730000000"}
```

`initData` — urlencoded-строка из MAX Bridge (`window.WebApp.initData`),
пары `key=value` через `&`, поле `user` — JSON с полями `id`, `first_name`,
`last_name`, `username`, `photo_url`.

Алгоритм handler'а:
1. Распарсить `initData` как query-string, достать `user`.
2. `user.id == 0` → `401`.
3. Upsert в таблицу `users` (повторный вход обновляет поля и `updated_at`,
   `created_at` сохраняется).
4. Вернуть пользователя.

Ответ `200`:
```json
{
  "id": 42,
  "username": "ivan",
  "first_name": "Ivan",
  "last_name": "Petrov",
  "photo_url": "",
  "language_code": "ru",
  "created_at": "2026-09-27T16:34:06.061663+03:00",
  "updated_at": "2026-09-27T16:34:06.073838+03:00"
}
```

Ошибки: `400` — нет тела / нет `initData`; `401` — `initData` не парсится или
`user.id` отсутствует; `500` — ошибка БД.

## GET /api/v1/candidates

Все активные резюме кандидатов для компаний — без привязки к вакансии
и скорингу. Bearer JWT, роль `recruiter` (иначе `403`).

Query: `q` (подстрока в title/skills/city), `limit` (1..100, по умолчанию 20),
`offset` ≥ 0.

Ответ `200`:
```json
{
  "total": 2,
  "items": [
    {"user": {"id": 42, "first_name": "Anna", "role": "candidate", "...": "..."},
     "resume": {"title": "Go-разработчик", "skills": "go, postgres", "city": "Санкт-Петербург", "...": "..."}}
  ]
}
```

Полный контракт всех ручек — живой Swagger: `https://eclipse-sim.ru/api/docs`.

## Безопасность

Подпись `initData` проверяется на бекенде HMAC-токеном бота
(`MAX_BOT_TOKEN`), алгоритм и Go-реализация — `docs/max/init-data.md`,
код — `internal/auth/initdata.go`. Если `MAX_BOT_TOKEN` не задан
(только локальная разработка), проверка отключена и пишется warning.

Дальше: JWT (`internal/auth/jwt.go`, 7 дней) + `RequireAuth` на всех
приватных ручках; демо-вход (`POST /auth/demo`) создаёт только
тестовых пользователей с ID 700000000–799999999.

Документация MAX: https://dev.max.ru
