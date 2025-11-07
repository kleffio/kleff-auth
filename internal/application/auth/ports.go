package auth

import (
	"context"
	"time"
)

type TenantRepoPort interface {
	ResolveTenantID(ctx context.Context, slug string) (string, error)
}

type UserRepoPort interface {
	CreateUser(ctx context.Context, tenantID string, email, username *string, passHash string, attrsJSON []byte) (userID string, err error)
	GetUserByIdentifier(ctx context.Context, tenantID, identifier string) (userID, passwordHash string, email, username *string, err error)
	UpdatePasswordHash(ctx context.Context, userID, newHash string) error
}

type PasswordHasherPort interface {
	Hash(plain string) (string, error)
	Verify(plain, encoded string) (bool, error)
	NeedsRehash(encoded string) bool
}

type TokenSignerPort interface {
	IssueAccess(sub, tid, scope, jti string, ttl time.Duration) (string, error)
	NewRefresh() (raw string, hash string, err error)
	JWKS() []byte
}
