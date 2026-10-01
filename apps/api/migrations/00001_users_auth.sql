-- +goose Up
CREATE TABLE users (
    id                UUID PRIMARY KEY,
    email             TEXT NOT NULL,
    name              TEXT NOT NULL,
    avatar_url        TEXT,
    role              TEXT NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'ADMIN')),
    email_verified_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email))
);
CREATE UNIQUE INDEX users_email_key ON users (email);

-- One row per way a user can sign in. Password hashes live here (not on
-- users) so Google/Apple identities slot in without schema churn.
CREATE TABLE auth_identities (
    id               UUID PRIMARY KEY,
    user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider         TEXT NOT NULL CHECK (provider IN ('PASSWORD', 'GOOGLE', 'APPLE')),
    provider_subject TEXT NOT NULL,
    password_hash    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT auth_identities_provider_subject_key UNIQUE (provider, provider_subject),
    CONSTRAINT auth_identities_password_hash CHECK ((provider = 'PASSWORD') = (password_hash IS NOT NULL))
);
CREATE INDEX auth_identities_user_id_idx ON auth_identities (user_id);

-- Opaque server-side sessions. Only the SHA-256 of the token is stored.
CREATE TABLE sessions (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sessions_token_hash_key UNIQUE (token_hash)
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- +goose Down
DROP TABLE sessions;
DROP TABLE auth_identities;
DROP TABLE users;
