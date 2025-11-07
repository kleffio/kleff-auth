package postgres

import (
	"context"
	"errors"
)

type UserRepo struct{ db *DB }

func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) CreateUser(ctx context.Context, tenantID string, email, username *string, passHash string, attrsJSON []byte) (string, error) {
	const q = `
INSERT INTO users (tenant_id, email, username, password_hash, attrs)
VALUES ($1, $2, $3, $4, COALESCE($5::jsonb, '{}'::jsonb))
RETURNING id::text`

	var id string

	if err := r.db.Pool.QueryRow(ctx, q, tenantID, email, username, passHash, string(attrsJSON)).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (r *UserRepo) GetUserByIdentifier(ctx context.Context, tenantID, identifier string) (userID, passwordHash string, email, username *string, err error) {
	const q = `
SELECT id::text, password_hash, email, username
FROM users
WHERE tenant_id=$1 AND (email=$2 OR username=$2)
LIMIT 1`

	var em, un *string

	if err = r.db.Pool.QueryRow(ctx, q, tenantID, identifier).Scan(&userID, &passwordHash, &em, &un); err != nil {
		return "", "", nil, nil, errors.New("not_found")
	}

	return userID, passwordHash, em, un, nil
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID, newHash string) error {
	const q = `UPDATE users SET password_hash=$2, updated_at=now() WHERE id=$1`

	ct, err := r.db.Pool.Exec(ctx, q, userID, newHash)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("user_not_found")
	}

	return nil
}
