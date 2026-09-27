# Spec: модель домена (миграции)

Тикеты: T-04 (см. `docs/tasks/`).

Все изменения — только новые миграции `backend/migrations/`. Формат:
`0000NN_имя.up.sql` + `.down.sql`, применяются автоматически при старте.

Решения по схеме (утверждены владельцем бека):
- храним максимум данных о кандидате, минимизация ПДн не приоритет;
- согласие на обработку данных — гейт онбординга, фиксируется в `user_consents`;
- позиция в компании (`company_members.position`) изменяема, задаётся при создании;
- история резюме (`resume_versions`) хранится полностью и **видна рекрутерам** —
  антифрад-подход: честные правки выглядят невинно, начальный вымысел раскрыт.

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
-- 'candidate' | 'recruiter' | NULL (роль не выбрана)
ALTER TABLE users ADD COLUMN birth_date DATE;
ALTER TABLE users ADD COLUMN gender TEXT NOT NULL DEFAULT '';
-- '' = не указан; 'male' | 'female' | 'other'

CREATE TABLE companies (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    website     TEXT        NOT NULL DEFAULT '',
    logo_url    TEXT        NOT NULL DEFAULT '',
    address     TEXT        NOT NULL DEFAULT '',
    verified    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE company_members (
    company_id BIGINT NOT NULL REFERENCES companies (id),
    user_id    BIGINT NOT NULL REFERENCES users (id),
    position   TEXT   NOT NULL DEFAULT 'employee',
    -- 'owner' | 'hr' | 'employee'; изменяемая, задаётся при создании компании
    PRIMARY KEY (company_id, user_id)
);

CREATE TABLE resumes (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL UNIQUE REFERENCES users (id),
    title               TEXT   NOT NULL,
    skills              TEXT   NOT NULL DEFAULT '',      -- "go, postgres, docker"
    experience_months   INTEGER NOT NULL DEFAULT 0,
    about               TEXT   NOT NULL DEFAULT '',
    education           TEXT   NOT NULL DEFAULT '',
    links               JSONB  NOT NULL DEFAULT '[]'::jsonb,
    -- [{"type": "github", "url": "https://..."}, ...]
    city                TEXT   NOT NULL DEFAULT '',
    work_format         TEXT   NOT NULL DEFAULT 'onsite',
    -- 'onsite' | 'hybrid' | 'remote'
    employment_type     TEXT   NOT NULL DEFAULT 'full_time',
    -- 'full_time' | 'part_time' | 'contract' | 'internship'
    salary_min          INTEGER,
    salary_max          INTEGER,
    source              TEXT   NOT NULL DEFAULT 'manual',
    -- 'manual' | 'file_parse'
    source_text         TEXT   NOT NULL DEFAULT '',
    -- полный текст загруженного/распарсенного резюме
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    last_confirmed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE resume_versions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id),
    snapshot   JSONB   NOT NULL,   -- полный слепок resumes на момент сохранения
    source     TEXT    NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- хранится полностью; рекрутерам доступна история версий резюме кандидата

CREATE TABLE vacancies (
    id                    BIGSERIAL PRIMARY KEY,
    company_id            BIGINT NOT NULL REFERENCES companies (id),
    title                 TEXT   NOT NULL,
    description           TEXT   NOT NULL DEFAULT '',
    required_skills       TEXT   NOT NULL DEFAULT '',
    min_experience_months INTEGER NOT NULL DEFAULT 0,
    city                  TEXT   NOT NULL DEFAULT '',
    work_format           TEXT   NOT NULL DEFAULT 'onsite',
    employment_type       TEXT   NOT NULL DEFAULT 'full_time',
    salary_min            INTEGER,
    salary_max            INTEGER,
    response_ttl_hours    INTEGER NOT NULL DEFAULT 48,
    is_active             BOOLEAN NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vacancies_company ON vacancies (company_id);
CREATE INDEX idx_resumes_active ON resumes (is_active);
CREATE INDEX idx_resume_versions_user ON resume_versions (user_id);
CREATE INDEX idx_company_members_user ON company_members (user_id);
```

## 000004 — приглашения и ответы

```sql
CREATE TABLE recruiter_actions (
    id                BIGSERIAL PRIMARY KEY,
    vacancy_id        BIGINT NOT NULL REFERENCES vacancies (id),
    recruiter_user_id BIGINT NOT NULL REFERENCES users (id),
    candidate_user_id BIGINT NOT NULL REFERENCES users (id),
    action            TEXT   NOT NULL,   -- 'invite' | 'skip'
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (vacancy_id, candidate_user_id)
);

CREATE TABLE candidate_responses (
    id           BIGSERIAL PRIMARY KEY,
    action_id    BIGINT NOT NULL UNIQUE REFERENCES recruiter_actions (id),
    response     TEXT   NOT NULL,   -- 'accept' | 'decline'
    responded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- `action`: `'invite' | 'skip'`. Один рекрутер — одно действие на пару
  (вакансия, кандидат), повторный свайп невозможен (UNIQUE).
- `response`: `'accept' | 'decline'`.
- Дедлайн TTL считается на лету: `created_at + vacancies.response_ttl_hours`.
- «Матч» = recruiter_actions.action='invite' + candidate_responses.response='accept'.

## Репозитории

Новые файлы в `internal/repository/`:

| Файл            | Методы                                                                                       |
|-----------------|----------------------------------------------------------------------------------------------|
| `company.go`    | Create (с member-позицией), GetByID, IsMember, ListByUser, UpdateMemberPosition              |
| `resume.go`     | Upsert (+снапшот в resume_versions в одной транзакции), GetByUserID, ListActive, TouchConfirmedAt, ListVersions |
| `vacancy.go`    | Create, GetByID, Update (ttl/is_active), ListByCompany                                       |
| `matching.go`   | CreateAction, ExistsAction, ListByCandidate, Get, CreateResponse, ListByVacancy (для ленты)  |

## Критерии приёмки

- [ ] Сервер со свежей БД стартует, миграции 000003–000004 применяются.
- [ ] Повторный старт не пытается применить их повторно.
- [ ] `down`-миграции откатывают изменения (проверить вручную через psql).
- [ ] UNIQUE-ограничение: повторный `recruiter_actions` на ту же пару даёт
      ошибку БД, репозиторий возвращает уже существующую запись.
