package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kleffio/kleff-auth/internal/utils"
)

type DB struct {
	Pool *pgxpool.Pool
}

func ConnectFromEnv(ctx context.Context) (*DB, error) {
	host := utils.GetEnv("DB_HOST", "localhost")
	port := utils.GetEnv("DB_PORT", "5432")
	user := utils.GetEnv("DB_USER", "postgres")
	pass := utils.GetEnv("DB_PASSWORD", "kleff")
	name := utils.GetEnv("DB_NAME", "kleff_auth")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}
