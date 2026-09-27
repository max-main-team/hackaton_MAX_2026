ALTER TABLE users ADD COLUMN language_code TEXT NOT NULL DEFAULT '';

CREATE TABLE user_consents (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id),
    consent     TEXT        NOT NULL,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, consent)
);
