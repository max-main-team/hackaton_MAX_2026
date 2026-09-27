ALTER TABLE users ADD COLUMN role TEXT;
ALTER TABLE users ADD COLUMN birth_date DATE;
ALTER TABLE users ADD COLUMN gender TEXT NOT NULL DEFAULT '';

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
    PRIMARY KEY (company_id, user_id)
);

CREATE TABLE resumes (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL UNIQUE REFERENCES users (id),
    title               TEXT   NOT NULL,
    skills              TEXT   NOT NULL DEFAULT '',
    experience_months   INTEGER NOT NULL DEFAULT 0,
    about               TEXT   NOT NULL DEFAULT '',
    education           TEXT   NOT NULL DEFAULT '',
    links               JSONB  NOT NULL DEFAULT '[]'::jsonb,
    city                TEXT   NOT NULL DEFAULT '',
    work_format         TEXT   NOT NULL DEFAULT 'onsite',
    employment_type     TEXT   NOT NULL DEFAULT 'full_time',
    salary_min          INTEGER,
    salary_max          INTEGER,
    source              TEXT   NOT NULL DEFAULT 'manual',
    source_text         TEXT   NOT NULL DEFAULT '',
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    last_confirmed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE resume_versions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id),
    snapshot   JSONB   NOT NULL,
    source     TEXT    NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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
