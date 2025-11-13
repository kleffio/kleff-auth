package postgres

import (
	"context"
	"encoding/json"

	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
)

type OAuthClientRepo struct{ db *DB }

func NewOAuthClientRepo(db *DB) *OAuthClientRepo {
	return &OAuthClientRepo{db: db}
}

func (r *OAuthClientRepo) GetOAuthClient(ctx context.Context, tenantID, clientID, provider string) (*domain.OAuthClient, error) {
	const q = `
SELECT id::text, tenant_id::text, client_id, name, redirect_uris, providers
FROM oauth_clients
WHERE tenant_id = $1::uuid AND client_id = $2;`

	var (
		id            string
		tid           string
		cid           string
		name          string
		redirectJSON  []byte
		providersJSON []byte
	)

	err := r.db.Pool.QueryRow(ctx, q, tenantID, clientID).Scan(
		&id, &tid, &cid, &name, &redirectJSON, &providersJSON,
	)
	if err != nil {
		return nil, err
	}

	var redirectURIs []string
	if err := json.Unmarshal(redirectJSON, &redirectURIs); err != nil {
		return nil, err
	}

	var providers map[string]domain.OAuthProviderConfig
	if err := json.Unmarshal(providersJSON, &providers); err != nil {
		return nil, err
	}

	return &domain.OAuthClient{
		ID:           id,
		TenantID:     tid,
		ClientID:     cid,
		Name:         name,
		RedirectURIs: redirectURIs,
		Providers:    providers,
	}, nil
}
