-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS sessions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id    uuid,
    family_id    uuid NOT NULL,
    parent_id    uuid,
    replaced_by  uuid,
    refresh_hash bytea NOT NULL,
    user_agent   text,
    ip           inet,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz,
    reason       text
);

ALTER TABLE sessions
    ADD CONSTRAINT fk_sessions_parent
        FOREIGN KEY (parent_id) REFERENCES sessions(id) ON DELETE SET NULL;

ALTER TABLE sessions
    ADD CONSTRAINT fk_sessions_replaced_by
        FOREIGN KEY (replaced_by) REFERENCES sessions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS ix_sessions_user_id      ON sessions (user_id);
CREATE INDEX IF NOT EXISTS ix_sessions_family_id    ON sessions (family_id);
CREATE INDEX IF NOT EXISTS ix_sessions_client_id    ON sessions (client_id);
CREATE INDEX IF NOT EXISTS ix_sessions_expires_at   ON sessions (expires_at);

CREATE INDEX IF NOT EXISTS ix_sessions_user_active
    ON sessions (user_id, expires_at)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS ix_sessions_user_active;
DROP INDEX IF EXISTS ix_sessions_expires_at;
DROP INDEX IF EXISTS ix_sessions_client_id;
DROP INDEX IF EXISTS ix_sessions_family_id;
DROP INDEX IF EXISTS ix_sessions_user_id;

ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_replaced_by;
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_parent;

DROP TABLE IF EXISTS sessions;
