# MAX Mini App — реверс-найм в мессенджере MAX

Мини-приложение, где работу ищут наоборот: **кандидат один раз заполняет резюме,
а компании сами находят его** под свои вакансии и присылают приглашения с дедлайном.
Принял приглашение — получил матч и контакт рекрутера.

**Прод:** <https://eclipse-sim.ru> · **Swagger:** <https://eclipse-sim.ru/api/docs>

| Слой     | Технологии                                        |
|----------|---------------------------------------------------|
| Бекенд   | Go 1.27, Echo v4, pgx v5 (pgxpool), PostgreSQL 17 |
| Фронтенд | React 19, TypeScript, Vite 8                      |
| AI       | GLM через z.ai (скоринг пар + парсинг PDF-резюме) |
| Инфра    | Docker Compose, GHCR, CI/CD на GitHub Actions     |

## Как это работает

1. **Кандидат** — заполняет резюме (вручную или загрузкой PDF: текст извлекаем,
   AI раскладывает по полям), выбирает город/формат/зарплату.
2. **Компания** — создаёт компанию и вакансии, смотрит подбор: кандидаты
   ранжируются скорингом (навыки 50% · город 20% · формат 20% · опыт 10%,
   дальше AI-обогащение) или весь список кандидатов с поиском.
3. **Приглашение** — рекрутер приглашает за один клик; у кандидата есть TTL
   на ответ (24–336 часов). Взаимное «да» = матч с контактом рекрутера.
4. **Бот MAX** — приветствует, отвечает на `/help` и `/status` (резюме,
   приглашения, квоты), напоминает о подтверждении актуальности резюме.
5. Дополнительно: карта вакансий, верификация компании, рефералки
   (кандидат→кандидат и компания→компания с квотами инвайтов).

## Быстрый старт (локально)

Нужны Docker, Go 1.27+, Node 22+.

```bash
git clone https://github.com/max-main-team/hackaton_MAX_2026.git
cd hackaton_MAX_2026

# 1. Postgres (порт 5433, чтобы не конфликтовать с локальным 5432)
docker compose up -d

# 2. Бекенд (миграции применятся сами при старте)
cd backend
cp .env.example .env        # заполните JWT_SECRET (любая длинная строка)
make run                    # слушает :8080

# 3. Фронтенд (второй терминал)
cd frontend
npm install
npm run dev                 # http://localhost:5173, /api/* проксируется на :8080
```

Откройте <http://localhost:5173> — если вы не внутри MAX, нажмите
**«Демо-вход»**: будет создан тестовый пользователь (ID 700000000–799999999),
выберите роль и пользуйтесь. Без MAX можно проверить весь продукт.

### Переменные окружения (backend/.env)

| Переменная     | Обязательна | Назначение                                            |
|----------------|-------------|-------------------------------------------------------|
| `JWT_SECRET`   | да          | Подпись JWT (без неё сервер не стартует)              |
| `DATABASE_URL` | да          | Postgres (по умолчанию `localhost:5433/maxapp`)       |
| `ADDR`         | нет         | Адрес слушателя (`:8080`)                             |
| `MAX_BOT_TOKEN`| нет         | Токен бота MAX: подпись initData, бот, рассылки       |
| `AI_API_KEY`   | нет         | Ключ z.ai: AI-скоринг и парсинг PDF (без ключа — алго-фолбэк) |
| `AI_BASE_URL`, `AI_MODEL` | нет | Эндпоинт и модель LLM                     |

### Проверки и тесты

```bash
cd backend  && make test && make vet && gofmt -l .   # юнит + интеграционные*
cd frontend && npm run lint && npm run build
```

\* интеграционные тесты (`handler`, `worker`) используют БД из `DATABASE_URL`;
без доступной БД они автоматически пропускаются. Live-проверка AI включается
переменной `AI_LIVE_TEST=1`.

## API

Базовый префикс `/api/v1`, авторизация — `Authorization: Bearer <JWT>`
(выдаётся на `POST /auth` по `initData` из MAX Bridge, подпись проверяется
токеном бота). Живой контракт: <https://eclipse-sim.ru/api/docs>.

Основные ручки:

| Метод и путь | Назначение |
|---|---|
| `GET /health` | живость сервиса и БД |
| `POST /auth`, `POST /auth/demo` | вход по initData / демо-вход |
| `GET /me`, `POST /me/role` | профиль и выбор роли |
| `GET/PUT /my/resume` | резюме кандидата (+версии) |
| `POST /my/resume/parse`, `/parse-file` | AI-парсинг текста / PDF-файла |
| `POST /my/resume/confirm-activity` | «подбор актуален» (еженедельный опрос) |
| `POST /companies`, `GET /my/companies` | создание и список компаний |
| `GET/POST /companies/{id}/vacancies`, `PATCH /vacancies/{id}` | вакансии |
| `GET /vacancies/{id}/candidates` | подбор со скорингом (list/feed) |
| `GET /candidates` | все кандидаты компании, поиск, пагинация |
| `POST /vacancies/{id}/candidates/{uid}/action` | invite/skip |
| `GET /my/invitations`, `POST /invitations/{id}/respond` | приглашения и ответ |
| `GET /vacancies/map` | вакансии для карты |
| `GET /my/referrals`, `/my/company-referrals` | рефералки |
| `POST /companies/{id}/verify` | верификация компании |

## Бот MAX

Команды: `/start` — приветствие, `/help` — как устроен сервис,
`/status` — персональный статус (резюме и приглашения у кандидата;
компания, верификация и квоты у рекрутера). Бот также рассылает
еженедельный опрос «подбор актуален?» и регистрирует команды в платформе
при старте. Код: `backend/internal/worker/`.

## DATA-API

В корне лежат `DATA-API.yaml` (обязательные HTTP-проверки в формате
DATA-API 1.0) и `openapi.yaml` (OpenAPI 3.1, конвертируется из Swagger
скриптом). Регенерация и проверка:

```bash
cd backend && make swag
python3 scripts/gen_openapi3.py                 # свежий openapi.yaml с прода
# валидатор организаторов: https://gitverse.ru/stasnorman/example-data-api
python3 validate_data_api.py DATA-API.yaml      # → «ПОДТВЕРЖДЕНИЕ», exit 0
```

## Структура репозитория

```
backend/
├── cmd/server/          # точка входа (config → db → migrate → echo → воркеры)
├── internal/
│   ├── auth/            # initData (парсинг/подпись), JWT
│   ├── config/          # env-конфиг
│   ├── database/        # pgxpool + миграции (embed)
│   ├── dto/             # запросы/ответы API
│   ├── geo/             # координаты городов
│   ├── handler/         # HTTP-обработчики
│   ├── middleware/      # логирование, RequireAuth
│   ├── repository/      # SQL-слой (pgx)
│   ├── scoring/         # алго-скоринг, AI-клиент, нормализация текста резюме
│   ├── server/          # сборка Echo: роуты и middleware
│   └── worker/          # бот (long polling) и еженедельный опрос
├── migrations/          # *.sql, применяются при старте
└── docs/                # сгенерированный Swagger

frontend/src/
├── api/                 # клиенты бекенда и типы
├── components/          # UI-компоненты (таббар, карточки, экран)
├── lib/                 # MAX Bridge, pdf-извлечение, сессия
└── routes/              # экраны по ролям (candidate/, recruiter/, карта)
```

Подробности — [`ARCHITECTURE.md`](./ARCHITECTURE.md) и [`docs/`](./docs/README.md).

## CI/CD и прод

Push в `main` триггерит `.github/workflows/ci-cd.yml`:

1. CI: бекенд (`build`/`vet`/`test`/`gofmt`) + фронтенд (`lint`/`build`)
2. Образы → GHCR (`hackaton_max_2026/backend|frontend:latest`)
3. Деплой по SSH на `/opt/max-miniapp` (Cloud.ru VPS), `docker compose pull && up -d`,
   проверка `GET /api/v1/health`

Схема прода: `интернет → nginx хоста (TLS, certbot) → frontend (nginx :8081)
→ backend :8080 → postgres`.

## Известное ограничение — платформа MAX

Тестовый бот хакатона ограничен самой платформой: нажатие «Начать» отклоняется
(`error.user.blocked.send — User is restricted`), диалог юзер↔бот не создаётся.
Мини-апп внутри MAX при этом полностью работает (авторизация по initData ✅).
После разблокировки бота организаторами бот-команды и рассылки заработают без
изменений кода. Для проверки продукта без MAX используйте «Демо-вход».
