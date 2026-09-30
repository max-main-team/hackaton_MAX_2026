# Spec: API v1 (полный контракт)


Общее: база `/api/v1`, JSON, ошибки `{"message": "..."}`. Все эндпоинты
кроме `/health` и `/auth` — за `RequireAuth` (см. `spec-auth.md`).
`user_id` всегда берётся из токена, из тела запроса не приходит.

## Аутентификация и профиль

### POST /auth — см. spec-auth.md

### GET /me
Ответ `200`: объект пользователя + `personal_data_accepted_at`
(`null` или timestamp из `user_consents`).

### POST /me/role
```json
{"role": "candidate", "acceptPersonalData": true}
```
- `role` ∈ `candidate | recruiter`, иначе 400.
- `acceptPersonalData = true` обязателен (иначе 400) → upsert в
  `user_consents ('personal_data')`.
- Повторная смена роли разрешена (для демо), ответ — обновлённый пользователь
  + `personal_data_accepted_at`.

## Кандидат: резюме

Одно резюме на пользователя (UNIQUE в БД).

### GET /my/resume
- `200` — резюме или `404 {"message":"resume not found"}`.

### PUT /my/resume
```json
{
  "title": "Backend разработчик",
  "skills": "go, postgres, docker",
  "experience_months": 36,
  "about": "Пишу сервисы на Go",
  "city": "Москва",
  "schedule": "remote"
}
```
- Валидация: `title` непустой (400), `schedule` из белого списка (400),
  `experience_months` 0..600.
- `200` — сохранённое резюме (все поля + `is_active`, `last_confirmed_at`).
- Upsert: создаёт или обновляет, `updated_at = now()`.

### POST /my/resume/confirm-activity
Ответ на еженедельный опрос (фича 1.1).
```json
{"active": true}
```
- `active: true` → `last_confirmed_at = now()`, `is_active = true`.
- `active: false` → `is_active = false`.
- `200` — обновлённое резюме.

## Рекрутер: компания и вакансии

### POST /companies
```json
{"name": "Ромашка", "description": "ИТ-компания"}
```
- Создаёт компанию и добавляет создателя в `company_members`.
- `200` — компания: `{id, name, description, verified: false, created_at}`.
- `name` непустой (400). Пользователь может состоять в нескольких компаниях
  (для MVP UI показывает первую).

### GET /my/companies
`200` — массив компаний текущего пользователя.

### POST /companies/{id}/vacancies
Тело:
```json
{
  "title": "Go-разработчик",
  "description": "Платёжный сервис",
  "required_skills": "go, postgres",
  "min_experience_months": 12,
  "city": "Москва",
  "schedule": "hybrid",
  "response_ttl_hours": 24
}
```
- Проверки: пользователь состоит в компании (403), `title` непустой,
  `schedule` валиден, `response_ttl_hours` 1..336 (400).
- `200` — вакансия со всеми полями.

### GET /companies/{id}/vacancies
`200` — массив вакансий компании (проверить членство, иначе 403).

### PATCH /vacancies/{id}
```json
{"response_ttl_hours": 72, "is_active": false}
```
- Частичное обновление: применять только переданные поля.
- Проверка членства (403). `200` — обновлённая вакансия.

## Матчинг: выдача кандидатов

### GET /vacancies/{id}/candidates?mode=list&limit=20&offset=0
- Проверка членства (403), вакансия активна (409).
- Только кандидаты с `role='candidate'`, активным резюме и без
  существующего `recruiter_actions` по этой вакансии.
- Скоринг — см. `spec-matching.md`; сортировка по score desc.
- `mode=list` — постранично; `mode=feed` — тот же список без пагинации
  (лента перемешивается на фронте).
- `200`:
```json
{
  "items": [
    {
      "score": 87,
      "breakdown": {"skills": 60, "city": 15, "schedule": 12, "experience": 0},
      "user": {"id": 42, "first_name": "Ivan", "photo_url": ""},
      "resume": {"title": "Backend разработчик", "skills": "go, postgres",
                 "experience_months": 36, "city": "Москва", "schedule": "remote",
                 "about": "..."}
    }
  ],
  "total": 134
}
```

### POST /vacancies/{id}/candidates/{candidateUserId}/action
```json
{"action": "invite"}
```
- `action` ∈ `invite | skip` (400).
- Если запись уже есть — вернуть существующую (200), не дублировать.
- `200` — запись действия: `{id, action, created_at}`.

## Приглашения: кандидат

### GET /my/invitations
Только `action='invite'` по резюме кандидата. Каждая запись:

```json
{
  "id": 7,
  "status": "pending",
  "deadline_at": "2026-09-29T18:00:00Z",
  "hours_left": 21.5,
  "company": {"id": 1, "name": "Ромашка", "verified": true},
  "vacancy": {"id": 3, "title": "Go-разработчик", "city": "Москва"},
  "response": null
}
```

- `status`: `pending` (ответа нет, дедлайн в будущем) | `overdue`
  (ответа нет, дедлайн прошёл) | `responded`.
- Сортировка: сначала pending по дедлайну asc, потом overdue, потом responded.

### POST /invitations/{id}/respond
```json
{"response": "accept"}
```
- Отвечать можно только на своё приглашение (404/403) и только пока
  `status = pending` (409 при overdue/responded).
- `response` ∈ `accept | decline` (400).
- `accept` создаёт матч: `200` — `{id, response, company, vacancy,
  recruiter_contact}`. В MVP `recruiter_contact` — username рекрутера
  в MAX (из users.username) для связи в чате.

## Критерии приёмки (на эндпоинт-группу)

- [ ] Все ручки доступны только с валидным токеном.
- [ ] Чужие компании/вакансии/приглашения → 403/404.
- [ ] Невалидные поля → 400 с человекочитаемым message.
- [ ] Контракты совпадают с примерами выше (фронт пишет типы по ним).
