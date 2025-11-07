package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kleffio/kleff-auth/internal/domain"
)

type TenantRepoPort interface {
	ResolveTenantID(ctx context.Context, tenant string) (tenantID string, err error)
}

type UserRepoPort interface {
	CreateUser(ctx context.Context, tenantID string, email, username *string, passwordHash string, attrsJSON *string) (userID string, err error)
	GetUserByIdentifier(ctx context.Context, tenantID string, identifier string) (userID string, passwordHash string, email *string, username *string, err error)
	GetUserByID(ctx context.Context, tenantID string, userID string) (email *string, username *string, err error)
	UpdatePasswordHash(ctx context.Context, userID string, newHash string) error
}

type PasswordHasherPort interface {
	Hash(plain string) (hash string, err error)
	Verify(plain string, hash string) (ok bool, err error)
	NeedsRehash(hash string) bool
}

type TokenSignerPort interface {
	IssueAccess(sub, tid, scope, jti string, ttl time.Duration) (string, error)
	ParseAccess(token string) (sub string, tenantID string, err error)
	JWKS() []byte
}

type RefreshTokenPort interface {
	Generate() (raw []byte, hash []byte, err error)
	Verify(hash []byte, raw []byte) error
	Encode(sessionID uuid.UUID, secret []byte) string
	Parse(token string) (sessionID uuid.UUID, secret []byte, err error)
}

type SessionRepoPort interface {
	Create(ctx context.Context, s *domain.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ReplacedAlready(ctx context.Context, id uuid.UUID) (bool, error)
	MarkReplaced(ctx context.Context, oldID, newID uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID, reason string) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID, reason string) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

type TimeProvider interface {
	Now() time.Time
}
