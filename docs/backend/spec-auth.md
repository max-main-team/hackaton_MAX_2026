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
- Подпись проверяется по алгоритму из `docs/max/init-data.md`:
  `hash = hex(HMAC_SHA256(secret_key, launch_params))`, где
  `secret_key = HMAC_SHA256("WebAppData", MAX_BOT_TOKEN)`, а
  `launch_params` — пары `key=value` (без `hash`), отсортированные по
  ключу и склеенные через `\n`. Сравнение — `hmac.Equal`.
- Отклонять `auth_date` старше 1 часа (рекомендация MAX,
  `ErrInitDataExpired`).
- **Dev-режим:** если `MAX_BOT_TOKEN` пуст — проверку подписи
  пропустить и писать warning в лог при старте. Это режим хакатона;
  в проде токен обязателен.

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

| Переменная     | Обязательность                                       |
|----------------|------------------------------------------------------|
| `JWT_SECRET`   | всегда                                               |
| `MAX_BOT_TOKEN`| прод (в dev может быть пустым); также нужен воркеру T-12 |

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
- [ ] `POST /auth` с протухшим `auth_date` (>1 часа) → 401.
- [ ] Запрос без токена к защищённому эндпоинту → 401.
- [ ] Запрос с битым/просроченным токеном → 401.
- [ ] Пустой `JWT_SECRET` → сервер не стартует с понятной ошибкой.
- [ ] `go build`, `go vet`, `gofmt -l` — чисто.

## Сравнение с референс-реализацией (max-main-team/backend_hackaton_MAX)

Референс: uni-bot расписание, тоже MAX mini app.

| Аспект | Референс | У нас |
|--------|----------|-------|
| Валидация подписи | HMAC-SHA256("WebAppData", botToken) → подпись строк, hex-compare | та же схема |
| Формат запроса | `POST /auth/login`, form-urlencoded (raw initData как form body) | `POST /auth` JSON c initData-строкой — эквивалентно |
| Свежесть auth_date | не проверяется — initData можно переигрывать бесконечно | отсечка > 1 часа |
| Сравнение хешей | `==` (не constant-time) | `hmac.Equal` |
| JWT | HS256, кастомные Claims с данными юзера внутри (имя, аватар) | HS256, минимальные claims (sub/iat/exp) — данные всегда свежие из БД |
| Refresh-токены | access 24ч + refresh 24 дня в БД + cookie HttpOnly + `POST /auth/refresh` | один access на 7 дней |
| Middleware | Bearer → ParseToken → юзер в контексте | то же |
| Upsert юзера | create/update при логине | то же |
| DTO | models/dto + http/dto разделены | internal/dto — тот же принцип |
| Swagger | swag-аннотации | то же |

Если 7-дневного access-токена окажется мало (жалобы на вылеты) — берём их
паттерн refresh-токенов: `refresh_repo` + cookie HttpOnly + `POST /auth/refresh`.
