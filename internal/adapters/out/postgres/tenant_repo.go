package postgres

import (
	"context"
	"errors"
)

type TenantRepo struct{ db *DB }

func NewTenantRepo(db *DB) *TenantRepo { return &TenantRepo{db: db} }

func (r *TenantRepo) ResolveTenantID(ctx context.Context, slug string) (string, error) {
	const q = `SELECT id::text FROM tenants WHERE slug=$1`

	var id string
	err := r.db.Pool.QueryRow(ctx, q, slug).Scan(&id)

	if err != nil {
		return "", errors.New("unknown tenant")
	}

	return id, nil
}
