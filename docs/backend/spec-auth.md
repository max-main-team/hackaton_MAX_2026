# Spec: аутентификация и безопасность

Тикеты: T-01, T-02, T-03 (см. `docs/tasks/`).

## Цель

Заменить «дырявый» `POST /auth` на полноценную схему: проверка подписи
initData → JWT → защищённые эндпоинты.

## 1. Проверка подписи initData

Файл: `backend/internal/auth/initdata.go`

```go
// ParseInitData разбирает initData и возвращает пользователя и auth_date.
func ParseInitData(raw string) (InitDataUser, time.Time, error)

// Verify проверяет подпись initData секретом приложения.
// Алгоритм HMAC — взять из официальной документации MAX (dev.max.ru),
// раздел Mini Apps → подпись initData.
func Verify(raw, appSecret string) error
```

Правила:

- Парсинг: `url.ParseQuery`, поле `user` — JSON (`id`, `first_name`,
  `last_name`, `username`, `photo_url`), поле `auth_date` — unix-секунды.
- Отклонять `auth_date` старше 24 часов (`ErrInitDataExpired`).
- **Dev-режим:** если `MAX_APP_SECRET_KEY` пуст — проверку подписи
  пропустить и писать warning в лог при старте. Это режим хакатона;
  в проде ключ обязателен.

## 2. JWT

Файл: `backend/internal/auth/jwt.go`, middleware: `backend/internal/middleware/auth.go`

- Библиотека: `github.com/golang-jwt/jwt/v5`.
- Алгоритм HS256, секрет из env `JWT_SECRET` (обязателен всегда; при
  пустом значении сервер не стартует — добавить проверку в `config.Load`
  с понятной ошибкой).
- Claims: `sub` = user.id, `iat`, `exp = iat + 7 дней`.
- Middleware `RequireAuth`: заголовок `Authorization: Bearer <token>`,
  при ошибке — `401 {"message":"unauthorized"}`.
- В context кладётся `user_id` (helper `UserIDFromContext(c) (int64, error)`).

Новые env (в `.env.example` тоже):

| Переменная          | Обязательность                          |
|---------------------|-----------------------------------------|
| `JWT_SECRET`        | всегда                                  |
| `MAX_APP_SECRET_KEY`| прод (в dev может быть пустым)          |

## 3. Обновлённый POST /api/v1/auth

Запрос: `{"initData": "..."}` (как сейчас).

Логика handler'а:

1. `ParseInitData` → 401 при ошибке.
2. `Verify` (если ключ задан) → 401 при невалидной подписи.
3. Upsert пользователя (как сейчас).
4. Ответ `200`:

```json
{
  "token": "eyJ...",
  "user": {
    "id": 42,
    "username": "ivan",
    "first_name": "Ivan",
    "last_name": "Petrov",
    "photo_url": "",
    "role": null,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

Поле `role` появляется после миграции 000002 (`null` до выбора роли).

## 4. Новые защищённые эндпоинты

Все эндпоинты из `spec-api.md`, кроме `/health` и `/auth`, регистрируются
в группе с `RequireAuth`:

```go
private := api.Group("")
private.Use(middleware.RequireAuth(jwtSecret))
```

## Критерии приёмки

- [ ] `POST /auth` с невалидной подписью → 401 (при заданном секретe).
- [ ] `POST /auth` с протухшим `auth_date` (>24ч) → 401.
- [ ] Запрос без токена к защищённому эндпоинту → 401.
- [ ] Запрос с битым/просроченным токеном → 401.
- [ ] Пустой `JWT_SECRET` → сервер не стартует с понятной ошибкой.
- [ ] `go build`, `go vet`, `gofmt -l` — чисто.
