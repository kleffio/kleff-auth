CREATE TABLE IF NOT EXISTS users (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email          citext,
    username       citext,
    password_hash  text NOT NULL,
    attrs          jsonb NOT NULL DEFAULT '{}'::jsonb,
    email_verified_at timestamptz,
    created_at     timestamptz NOT NULL DEFAULT NOW(),
    updated_at     timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_tenant_email
    ON users (tenant_id, email) WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_tenant_username
    ON users (tenant_id, username) WHERE username IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_users_tenant_id ON users (tenant_id);

CREATE TRIGGER users_set_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
