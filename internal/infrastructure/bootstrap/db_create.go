package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kleffio/kleff-auth/internal/utils"
)

func EnsureDatabase(ctx context.Context) error {
	host := utils.GetEnv("DB_HOST", "localhost")
	port := utils.GetEnv("DB_PORT", "5432")
	user := utils.GetEnv("DB_USER", "postgres")
	pass := utils.GetEnv("DB_PASSWORD", "")
	name := utils.GetEnv("DB_NAME", "kleff_auth")

	maintURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", user, pass, host, port)

	conn, err := pgx.Connect(ctx, maintURL)

	if err != nil {
		return fmt.Errorf("connect (maintenance): %w", err)
	}

	defer conn.Close(ctx)

	var exists bool

	if err := conn.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)`, name,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if exists {
		return nil
	}

	sql := `CREATE DATABASE ` + pgx.Identifier{name}.Sanitize() + ` OWNER ` + pgx.Identifier{user}.Sanitize() + ` TEMPLATE template0 ENCODING 'UTF8'`

	if _, err := conn.Exec(ctx, sql); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "42501": // INSUFFICIENT_PRIVILEGE
				return fmt.Errorf("cannot create database %q: role %q lacks CREATEDB; create it via infra or grant CREATEDB (dev only)", name, user)
			case "42P04": // DuplicateDatabase
				return nil
			}
		}

		return fmt.Errorf("create database: %w", err)
	}

	return nil
}
