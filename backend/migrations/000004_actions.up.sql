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

CREATE INDEX idx_recruiter_actions_candidate ON recruiter_actions (candidate_user_id);
CREATE INDEX idx_recruiter_actions_vacancy ON recruiter_actions (vacancy_id);
