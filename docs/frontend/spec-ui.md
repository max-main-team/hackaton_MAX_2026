# Spec: фронтенд MVP (экраны и структура)


## Новые зависимости

- `react-router-dom` (роутинг).

## Структура после MVP

```
src/
├── main.tsx              RouterProvider
├── App.tsx               удалить (разносится по роутам)
├── api/
│   ├── client.ts         был lib/api.ts: fetch-обёртка + Bearer-токен
│   ├── auth.ts           auth(), me(), setRole()
│   ├── resume.ts         getResume(), saveResume(), confirmActivity()
│   ├── company.ts        createCompany(), myCompanies(), createVacancy(), ...
│   └── matching.ts       candidates(), action(), invitations(), respond()
├── lib/
│   ├── max.ts            как сейчас
│   └── session.ts        хранение JWT: localStorage 'max_token', helper authHeaders()
├── routes/
│   ├── login.tsx         авто-логин через initData → редирект
│   ├── onboarding.tsx    выбор роли
│   ├── candidate/
│   │   ├── resume.tsx    форма резюме
│   │   └── invitations.tsx список приглашений с дедлайном
│   └── recruiter/
│       ├── company.tsx   создание/просмотр компании
│       ├── vacancies.tsx список вакансий
│       ├── vacancy-new.tsx форма вакансии (+TTL)
│       ├── feed.tsx      свайп-лента кандидатов
│       └── candidate-list.tsx табличный список
└── components/
    ├── UserCard.tsx      карточка кандидата (для ленты и списка)
    ├── Badge.tsx         бейдж «верифицирована»
    └── Countdown.tsx     «осталось N ч» / «просрочено»
```

## Роуты

| Путь                       | Экран            | Доступ      |
|----------------------------|------------------|-------------|
| `/`                        | редирект по роли | авторизован |
| `/onboarding`              | выбор роли       | авторизован |
| `/resume`                  | моё резюме       | candidate   |
| `/invitations`             | приглашения      | candidate   |
| `/company`                 | компания         | recruiter   |
| `/company/vacancies`       | вакансии         | recruiter   |
| `/company/vacancies/new`   | новая вакансия   | recruiter   |
| `/vacancy/:id/feed`        | свайп-лента      | recruiter   |
| `/vacancy/:id/list`        | список           | recruiter   |

Гард-логика: без токена → `/`; без роли → `/onboarding`; не та роль →
редирект на свой домашний экран.

## Ключевые экраны

### Логин (`/`)
При монтировании: `api.auth(getInitData())` → токен в `session.ts` →
редирект по роли. Вне MAX (`getInitData() === ''`) — экран с кнопкой
«Демо-вход»: тот же `POST /auth` с пустой строкой уйдёт на бекенд-заглушку
(бекенд в dev без секрета пропускает подпись) — иначе показать текст
«Открой мини-апп через MAX».

### Свайп-лента (`/vacancy/:id/feed`, фича 1.5)
- Данные: `GET /vacancies/{id}/candidates?mode=feed`.
- Карточка `UserCard`: имя, фото, title, навыки тегами, score-бейдж.
- Свайп влево/кнопка ✕ → `POST .../action {action: "skip"}`.
- Свайп вправо/кнопка ♥ → `invite`. После — следующая карточка.
- Хаптика: `HapticFeedback.impactOccurred('light')` на свайп.
- Пустая лента → «Кандидаты закончились».

### Приглашения (`/invitations`, фича 1.2 для кандидата)
- Данные: `GET /my/invitations`.
- `Countdown` для `pending` (тик каждую минуту), красный «Просрочено»
  для `overdue` — кнопки «Принять/Отклонить» задизейблены.
- `accept` → экран успеха (матч: название компании + контакт рекрутера).

### Список кандидатов (`/vacancy/:id/list`, фича 1.4)
- `GET /vacancies/{id}/candidates?mode=list&limit=20&offset=0`: таблица (имя, title, город, score),
  сортировка по score, клик → разворот карточки с кнопками invite/skip.

### Настройка TTL (в форме вакансии)
- Поле «Срок ответа, часов» (селект: 12 / 24 / 48 / 72 / 168,
  дефолт 48) — `POST /companies/:id/vacancies` и `PATCH /vacancies/:id`.

## Правила

- Типы API — snake_case как в бекенде, хранить рядом с api-модулем.
- Каждый запрос: состояния loading / error / empty отрисовать явно.
- Токен добавлять во все запросы через `session.ts`
  (`Authorization: Bearer`), 401 → очистить токен → `/`.
- Комментарии в коде не пишем.

## Критерии приёмки

- [ ] `npm run build`, `npm run lint` зелёные.
- [ ] Полный сценарий демо проходит: логин → роль → резюме → (второй
      пользователь) компания → вакансия → лента → invite → у кандидата
      приглашение с дедлайном → accept → матч-экран.
- [ ] Все экраны в тёмной теме MAX, без горизонтального скролла в 360px.
