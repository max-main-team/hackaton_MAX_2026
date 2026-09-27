# Spec: модель домена (миграции)

Тикеты: T-04 (см. `docs/tasks/`).

Все изменения — только новые миграции `backend/migrations/`. Формат:
`0000NN_имя.up.sql` + `.down.sql`, применяются автоматически при старте.

## 000002 — user_fields (уже применена в базовом коммите)

- `ALTER TABLE users ADD COLUMN language_code TEXT NOT NULL DEFAULT ''`
  (из initData, для локализации).
- Таблица согласий:

```sql
CREATE TABLE user_consents (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id),
    consent     TEXT        NOT NULL,   -- 'personal_data' | ...
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, consent)
);
```

## 000003 — роли, компании, резюме, вакансии

```sql
ALTER TABLE users ADD COLUMN role TEXT;
-- возможные значения: 'candidate' | 'recruiter' | NULL (роль не выбрана)

CREATE TABLE companies (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    verified    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE company_members (
    company_id BIGINT NOT NULL REFERENCES companies (id),
    user_id    BIGINT NOT NULL REFERENCES users (id),
    PRIMARY KEY (company_id, user_id)
);

CREATE TABLE resumes (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL UNIQUE REFERENCES users (id),
    title             TEXT   NOT NULL,
    skills            TEXT   NOT NULL DEFAULT '',
    experience_months INTEGER NOT NULL DEFAULT 0,
    about             TEXT   NOT NULL DEFAULT '',
    city              TEXT   NOT NULL DEFAULT '',
    schedule          TEXT   NOT NULL DEFAULT 'full',
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    last_confirmed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE vacancies (
    id                  BIGSERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL REFERENCES companies (id),
    title               TEXT   NOT NULL,
    description         TEXT   NOT NULL DEFAULT '',
    required_skills     TEXT   NOT NULL DEFAULT '',
    min_experience_months INTEGER NOT NULL DEFAULT 0,
    city                TEXT   NOT NULL DEFAULT '',
    schedule            TEXT   NOT NULL DEFAULT 'full',
    response_ttl_hours  INTEGER NOT NULL DEFAULT 48,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Конвенция `schedule`: `'full' | 'part' | 'remote' | 'hybrid'` — общая для
резюме и вакансий, валидируется в handler'е (белый список).

## 000004 — приглашения и ответы

```sql
CREATE TABLE recruiter_actions (
    id                BIGSERIAL PRIMARY KEY,
    vacancy_id        BIGINT NOT NULL REFERENCES vacancies (id),
    recruiter_user_id BIGINT NOT NULL REFERENCES users (id),
    candidate_user_id BIGINT NOT NULL REFERENCES users (id),
    action            TEXT   NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (vacancy_id, candidate_user_id)
);

CREATE TABLE candidate_responses (
    id           BIGSERIAL PRIMARY KEY,
    action_id    BIGINT NOT NULL UNIQUE REFERENCES recruiter_actions (id),
    response     TEXT   NOT NULL,
    responded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- `action`: `'invite' | 'skip'`. Один рекрутер — одно действие на пару
  (вакансия, кандидат), повторный свайп невозможен (UNIQUE).
- `response`: `'accept' | 'decline'`.
- Дедлайн TTL считается на лету: `created_at + vacancies.response_ttl_hours`.
- «Матч» = recruiter_actions.action='invite' + candidate_responses.response='accept'.

## Индексы (в тех же миграциях)

```sql
CREATE INDEX idx_vacancies_company ON vacancies (company_id);
CREATE INDEX idx_recruiter_actions_candidate ON recruiter_actions (candidate_user_id);
CREATE INDEX idx_recruiter_actions_vacancy ON recruiter_actions (vacancy_id);
CREATE INDEX idx_resumes_active ON resumes (is_active);
```

## Репозитории

Новые файлы в `internal/repository/`:

| Файл            | Методы                                                                                       |
|-----------------|----------------------------------------------------------------------------------------------|
| `company.go`    | Create, GetByID, AddMember, ListByUser, ListVacancies                                        |
| `resume.go`     | Upsert, GetByUserID, ListActive, TouchConfirmedAt                                            |
| `vacancy.go`    | Create, GetByID, Update (ttl/is_active), ListByCompany                                       |
| `matching.go`   | CreateAction, ExistsAction, ListByCandidate, Get, CreateResponse, ListByVacancy (для ленты)  |

## Критерии приёмки

- [ ] Сервер со свежей БД стартует, миграции 000003–000004 применяются.
- [ ] Повторный старт не пытается применить их повторно.
- [ ] `down`-миграции откатывают изменения (проверить вручную через psql).
- [ ] UNIQUE-ограничение: повторный `recruiter_actions` на ту же пару даёт
      ошибку БД, репозиторий возвращает уже существующую запись.
