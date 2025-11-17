-- +goose Up
CREATE TABLE IF NOT EXISTS oauth_clients (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_id     text NOT NULL,
    name          text NOT NULL,
    redirect_uris jsonb NOT NULL DEFAULT '[]'::jsonb,
    providers     jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at    timestamptz NOT NULL DEFAULT NOW(),
    updated_at    timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT ux_oauth_clients_tenant_client UNIQUE(tenant_id, client_id)
);

CREATE INDEX IF NOT EXISTS ix_oauth_clients_tenant_id ON oauth_clients(tenant_id);

CREATE TRIGGER oauth_clients_set_timestamp
    BEFORE UPDATE ON oauth_clients
    FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

-- +goose Down
DROP TRIGGER IF EXISTS oauth_clients_set_timestamp ON oauth_clients;
DROP INDEX IF EXISTS ix_oauth_clients_tenant_id;
DROP TABLE IF EXISTS oauth_clients;