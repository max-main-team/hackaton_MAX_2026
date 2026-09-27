CREATE TABLE resume_scores (
    id         BIGSERIAL PRIMARY KEY,
    resume_id  BIGINT NOT NULL REFERENCES resumes (id) ON DELETE CASCADE,
    vacancy_id BIGINT NOT NULL REFERENCES vacancies (id) ON DELETE CASCADE,
    algo_score INTEGER NOT NULL,
    ai_score   INTEGER,
    ai_comment TEXT    NOT NULL DEFAULT '',
    ai_model   TEXT    NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (resume_id, vacancy_id)
);
