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
  "created_at": "2026-09-27T16:34:06.061663+03:00",
  "updated_at": "2026-09-27T16:34:06.073838+03:00"
}
```

Ошибки: `400` — нет тела / нет `initData`; `401` — `initData` не парсится или
`user.id` отсутствует; `500` — ошибка БД.

## ⚠️ Безопасность (сделать до прода)

Подпись `initData` **не проверяется**: любой клиент может прислать чужой
`user.id`. План:

1. Взять токен бота, к которому привязан мини-апп (платформа business.max.ru)
   → `MAX_BOT_TOKEN`; он же нужен воркеру рассылки (T-12).
2. Проверять подпись `initData` (HMAC по алгоритму из документации MAX,
   поля сортируются, исключая `signature`).
3. После проверки — выдавать JWT/сессию и защищать остальные эндпоинты
   middleware'ом авторизации.

Документация: https://dev.max.ru
