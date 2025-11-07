package postgres

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/kleffio/kleff-auth/internal/domain"
)

type SessionRepo struct{ db *DB }

func NewSessionRepo(db *DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	const q = `
INSERT INTO sessions (
  id, user_id, client_id, family_id, parent_id, refresh_hash,
  user_agent, ip, expires_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING created_at, last_used_at, revoked_at, reason, replaced_by;`

	var (
		createdAt time.Time
		lastUsed  time.Time
		revokedAt *time.Time
		reason    *string
		replaced  *uuid.UUID
	)

	ip := net.IP(s.IP)

	return r.db.Pool.QueryRow(ctx, q,
		s.ID, s.UserID, s.ClientID, s.FamilyID, s.ParentID, s.RefreshHash,
		s.UserAgent, ip, s.ExpiresAt,
	).Scan(&createdAt, &lastUsed, &revokedAt, &reason, &replaced)
}

func (r *SessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	const q = `
SELECT id, user_id, client_id, family_id, parent_id, replaced_by,
       refresh_hash, user_agent, ip, created_at, last_used_at, expires_at, revoked_at, reason
FROM sessions WHERE id = $1;`

	var s domain.Session
	var ip net.IP

	if err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.UserID, &s.ClientID, &s.FamilyID, &s.ParentID, &s.ReplacedBy,
		&s.RefreshHash, &s.UserAgent, &ip, &s.CreatedAt, &s.LastUsedAt,
		&s.ExpiresAt, &s.RevokedAt, &s.Reason,
	); err != nil {
		return nil, err
	}

	s.IP = ip

	return &s, nil
}

func (r *SessionRepo) ReplacedAlready(ctx context.Context, id uuid.UUID) (bool, error) {
	const q = `SELECT replaced_by IS NOT NULL FROM sessions WHERE id = $1;`

	var b bool

	if err := r.db.Pool.QueryRow(ctx, q, id).Scan(&b); err != nil {
		return false, err
	}

	return b, nil
}

func (r *SessionRepo) MarkReplaced(ctx context.Context, oldID, newID uuid.UUID) error {
	const q = `
UPDATE sessions
SET replaced_by = $2, revoked_at = now(), reason = 'rotated'
WHERE id = $1;`

	ct, err := r.db.Pool.Exec(ctx, q, oldID, newID)

	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("session not found")
	}

	return nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID, reason string) error {
	const q = `
UPDATE sessions
SET revoked_at = now(), reason = $2
WHERE id = $1 AND revoked_at IS NULL;`

	_, err := r.db.Pool.Exec(ctx, q, id, reason)

	return err
}

func (r *SessionRepo) RevokeFamily(ctx context.Context, familyID uuid.UUID, reason string) error {
	const q = `
UPDATE sessions
SET revoked_at = now(), reason = $2
WHERE family_id = $1 AND revoked_at IS NULL;`

	_, err := r.db.Pool.Exec(ctx, q, familyID, reason)

	return err
}

func (r *SessionRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sessions SET last_used_at = now() WHERE id = $1;`

	_, err := r.db.Pool.Exec(ctx, q, id)

	return err
}
