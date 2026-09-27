# Этап 1 — Домен и безопасность: детальный план

Расшифровка этапа 1 из `plan.md` в пошаговый план имплементации.
Контракты эндпоинтов — в `backend/spec-api.md`, схема БД —
`backend/spec-domain.md`, авторизация — `backend/spec-auth.md`,
экраны — `frontend/spec-ui.md`. Этот документ задаёт порядок работ,
границы этапа и критерии готовности.

## Результат этапа (что можно показать в конце)

1. Логин в мини-аппе через MAX → выбор роли.
2. Кандидат заполняет резюме (одно на пользователя).
3. Рекрутер создаёт компанию и вакансию (с TTL-полем).
4. Всё это — под JWT, с проверкой подписи initData.

Всё, что связано с подбором (скоринг, лента, приглашения, уведомления) —
НЕ в этом этапе.

## Границы этапа

| Входит (бек)                                   | Тикеты   |
|------------------------------------------------|----------|
| `internal/auth`: initData + JWT, `RequireAuth` | T-01–T-03|
| Миграция 000002 (роли, компании, резюме, вакансии) + репозитории | T-04 (частично: без matching.go) |
| `GET /me`, `POST /me/role`                     | T-05     |
| `GET/PUT /my/resume`                           | T-06     |
| Компания и CRUD вакансий                       | T-07     |

| Входит (фронт)                                 | Тикеты   |
|------------------------------------------------|----------|
| Роутинг, `session.ts`, api-модули с Bearer     | T-13     |
| Логин + онбординг роли                         | T-14     |
| Экран резюме                                   | T-15     |
| Компания → вакансия → список вакансий          | T-16     |

| Не входит (этапы 2–3)                          |
|------------------------------------------------|
| Миграция 000003 (recruiter_actions), scoring, выдача кандидатов, свайпы, TTL-статусы, воркер опроса, сиды |

## Порядок работ

### Шаг 1 — конфиг (бек)

Файлы: `internal/config/config.go`, `.env.example`.

- `JWTSecret string` из `JWT_SECRET` — **обязательный**: `Load()` возвращает
  конфиг только с непустым секретом, иначе `panic` с сообщением
  «JWT_SECRET is required» (упростим: паника при старте допустима).
- `MaxBotToken` — остаётся опциональным (dev-режим).
- Обновить `.env.example`.

Проверка: без `JWT_SECRET` сервер не стартует, с ним — стартует.

### Шаг 2 — пакет `internal/auth` (бек)

Файлы: `internal/auth/initdata.go`, `internal/auth/jwt.go`, `internal/auth/jwt_test.go`.

`initdata.go`:
- `ParseInitData(raw string) (InitDataUser, time.Time, error)` — как
  существующий `parseInitData` из handler'а, перенести сюда + `auth_date`.
- `Verify(raw, appSecret string) error` — алгоритм HMAC по документации
  MAX; до уточнения алгоритма вернуть `nil` с комментарием-заглушкой
  **запрещено** оставлять молча: функция должна логировать через
  возвращаемую ошибку `ErrNotImplemented`, а вызов в dev-режиме
  (пустой секрет) её не доходит до проверки. Итоговое правило:
  секрет пуст → проверка пропускается (warning при старте);
  секрет задан → `Verify` обязан реально проверять.
- Отклонять `auth_date` старше 24 часов.

`jwt.go`:
- `Issue(userID int64, secret string, now time.Time) (string, error)` —
  HS256, `sub`, `iat`, `exp = now + 168h`.
- `Parse(token, secret string) (int64, error)` — возвращает `userID`.
- Тест: roundtrip `Issue → Parse`, битый подпись/просроченный → ошибка.

### Шаг 3 — middleware и роуты (бек)

Файлы: `internal/middleware/auth.go`, `internal/server/server.go`.

- `RequireAuth(secret string) echo.MiddlewareFunc`: `Authorization: Bearer`,
  ошибки → 401; успех → `c.Set("user_id", id)`.
- Хелпер `auth.UserIDFromContext(c) (int64, error)`.
- В `setupRoutes`: приватная группа `private := api.Group("")`,
  `private.Use(middleware.RequireAuth(cfg.JWTSecret))`.

### Шаг 4 — новый `/auth` (бек)

Файл: `internal/handler/auth.go`.

1. `ParseInitData` → 401 при ошибке/протухшем auth_date.
2. Секрет задан → `Verify` → 401 при ошибке.
3. Upsert пользователя (как сейчас).
4. `Issue` токен → ответ `{"token": "...", "user": {...}}`.

⚠️ Синхронно с шагом 9: старый фронт ждёт плоский объект пользователя —
бек и фронт меняются одним коммитом (или T-03+T-13 вместе).

### Шаг 5 — миграция 000002 (бек)

Файлы: `migrations/000002_domain.up.sql` / `.down.sql`.

Только из `spec-domain.md`: `users.role`, `companies`, `company_members`,
`resumes`, `vacancies` + индексы `idx_vacancies_company`, `idx_resumes_active`.
Миграция 000003 (recruiter_actions) — этап 2, не делать сейчас.

Репозитории: `repository/resume.go`, `repository/company.go`,
`repository/vacancy.go` (методы по таблице из spec-domain.md;
`matching.go` — этап 2).

### Шаг 6 — профиль и резюме (бек)

Файлы: `internal/handler/me.go`, `internal/handler/resume.go`, регистрация в `server.go`.

- `GET /me` — пользователь из токена.
- `POST /me/role` — валидация `candidate|recruiter`, `UPDATE users SET role`.
- `GET /my/resume` — 200/404.
- `PUT /my/resume` — валидация по spec-api.md (title, schedule белый список
  `full|part|remote|hybrid`, experience 0..600), upsert.

Проверка curl'ом (dev, без секрета — initData любой валидной структуры):

```bash
TOKEN=$(curl -s localhost:8080/api/v1/auth -H 'Content-Type: application/json' \
  -d '{"initData":"user=%7B%22id%22%3A1%7D&auth_date=9999999999"}' | jq -r .token)
curl -s localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
curl -s -X PUT localhost:8080/api/v1/my/resume -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Go dev","skills":"go, sql","experience_months":24,"city":"Москва","schedule":"remote"}'
```

### Шаг 7 — компания и вакансии (бек)

Файлы: `internal/handler/company.go`, `internal/handler/vacancy.go`.

- `POST /companies` — создать + добавить в members.
- `GET /my/companies`.
- `POST /companies/{id}/vacancies` — проверка членства (403), валидация
  (title, schedule, ttl 1..336).
- `GET /companies/{id}/vacancies`, `PATCH /vacancies/{id}` (ttl/is_active).

### Шаг 8 — каркас фронта

Файлы: `src/lib/session.ts`, `src/api/*.ts`, `src/routes/*`, `src/main.tsx`.

- `npm i react-router-dom`.
- `session.ts`: `getToken/setToken/clearToken` (localStorage `max_token`),
  `authHeaders()`, обработка 401 → `clearToken` + редирект `/`.
- api-модули: `auth.ts` (`auth`, `me`, `setRole`), `resume.ts`,
  `company.ts` — типы snake_case по spec-api.md.
- `main.tsx` → `RouterProvider`; удалить старый `App.tsx` (логин-кнопка
  переезжает в `routes/login.tsx`).

### Шаг 9 — экраны

- `login.tsx`: авто-`auth(getInitData())` → токен → редирект по `role`
  (`null` → `/onboarding`).
- `onboarding.tsx`: две карточки «Я кандидат» / «Я рекрутер» → `setRole`.
- `candidate/resume.tsx`: форма по полям резюме, schedule — селект,
  сохранение `PUT /my/resume`, состояния loading/error/saved.
- `recruiter/company.tsx`: если `GET /my/companies` пуст — форма создания;
  иначе карточка + кнопка «Вакансии».
- `recruiter/vacancy-new.tsx`: поля вакансии + TTL-селект
  (12/24/48/72/168, дефолт 48) → создание → переход к списку.
- `recruiter/vacancies.tsx`: список вакансий компании, бейдж active/черновик.

### Шаг 10 — сквозная проверка и DoD

Сценарий двумя ролями (dev, curl или два окна браузера с разными
localStorage):

1. Пользователь A: логин → роль candidate → резюме сохранено,
   `GET /my/resume` возвращает его.
2. Пользователь B: логин → роль recruiter → компания → вакансия с TTL 24.
3. Невалидные кейсы: без токена → 401; чужая компания → 403; bad schedule → 400.

## Критерии приёмки этапа

- [ ] Сервер не стартует без `JWT_SECRET`; стартует с пустым `MAX_BOT_TOKEN` (warning в логе).
- [ ] `/auth` отдаёт `{token, user}`; защищённые ручки без токена → 401.
- [ ] Миграция 000002 применяется, повторный старт — идемпотентен.
- [ ] Сценарий из шага 10 проходит целиком.
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
- [ ] `npm run build && npm run lint` — чисто.
- [ ] Docs обновлены: `backend/api.md` (новые ручки), `database.md`
      (000002), `frontend/architecture.md` (роуты) — тем же коммитом.

## Риски / открытые вопросы

- Алгоритм подписи initData — уточнить по dev.max.ru до прода; в dev
  работаем без секрета (см. spec-auth.md).
- Смена роли после привязки к компании — MVP разрешает всё, ограничить
  после хакатона.
- Демо на двух ролях в одном браузере: использовать два профиля или
  curl-сценарий из шага 10.
