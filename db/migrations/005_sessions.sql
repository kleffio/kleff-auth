-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id    uuid,
    refresh_hash text NOT NULL UNIQUE,
    family_id    uuid,
    user_agent   text,
    ip           text,
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz
);

CREATE INDEX IF NOT EXISTS ix_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS ix_sessions_family_id ON sessions (family_id);

-- +goose Down
DROP INDEX IF EXISTS ix_sessions_family_id;
DROP INDEX IF EXISTS ix_sessions_user_id;
DROP TABLE IF EXISTS sessions;
