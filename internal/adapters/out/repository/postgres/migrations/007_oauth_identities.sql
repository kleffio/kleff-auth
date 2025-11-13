-- +goose Up
CREATE TABLE IF NOT EXISTS oauth_identities (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider          text NOT NULL,
    provider_user_id  text NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT ux_oauth_identities_provider_user UNIQUE(provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS ix_oauth_identities_user_id ON oauth_identities(user_id);
CREATE INDEX IF NOT EXISTS ix_oauth_identities_provider ON oauth_identities(provider);

-- +goose Down
DROP INDEX IF EXISTS ix_oauth_identities_provider;
DROP INDEX IF EXISTS ix_oauth_identities_user_id;
DROP TABLE IF EXISTS oauth_identities;