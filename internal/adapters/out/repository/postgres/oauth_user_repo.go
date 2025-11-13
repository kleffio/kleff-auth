package postgres

import (
	"context"
	"encoding/json"
	"errors"

	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
)

type OAuthUserRepo struct{ db *DB }

func NewOAuthUserRepo(db *DB) *OAuthUserRepo {
	return &OAuthUserRepo{db: db}
}

func (r *OAuthUserRepo) GetUserByOAuth(
	ctx context.Context,
	provider, providerUserID string,
) (tenantID string, userID string, email, username *string, err error) {
	const q = `
SELECT 
	u.tenant_id::text,
	u.id::text,
	u.email,
	u.username
FROM oauth_identities oi
JOIN users u ON oi.user_id = u.id
WHERE oi.provider = $1 AND oi.provider_user_id = $2
LIMIT 1;`

	err = r.db.Pool.QueryRow(ctx, q, provider, providerUserID).Scan(
		&tenantID, &userID, &email, &username,
	)
	if err != nil {
		return "", "", nil, nil, err
	}

	return tenantID, userID, email, username, nil
}

func (r *OAuthUserRepo) CreateUserFromOAuth(
	ctx context.Context,
	identity *domain.OAuthIdentity,
) (tenantID string, userID string, err error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return "", "", err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const getTenantQ = `SELECT id::text FROM tenants WHERE slug = $1;`
	err = tx.QueryRow(ctx, getTenantQ, identity.TenantSlug).Scan(&tenantID)
	if err != nil {
		return "", "", errors.New("tenant not found")
	}

	attrsJSON, _ := json.Marshal(identity.Attrs)
	const createUserQ = `
INSERT INTO users (tenant_id, email, username, password_hash, attrs)
VALUES ($1::uuid, $2, $3, '', $4::jsonb)
RETURNING id::text;`

	err = tx.QueryRow(ctx, createUserQ,
		tenantID, identity.Email, identity.Username, attrsJSON,
	).Scan(&userID)
	if err != nil {
		return "", "", err
	}

	const linkOAuthQ = `
INSERT INTO oauth_identities (user_id, provider, provider_user_id)
VALUES ($1::uuid, $2, $3);`

	_, err = tx.Exec(ctx, linkOAuthQ, userID, identity.Provider, identity.ProviderUserID)
	if err != nil {
		return "", "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return "", "", err
	}

	return tenantID, userID, nil
}
