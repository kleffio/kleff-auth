package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pg "github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres"
	"github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres/migrations"
	"github.com/kleffio/kleff-auth/internal/config"
)

// SetupDatabase is the one-stop DB bootstrap:
//  1. optionally creates the database (DB_CREATE)
//  2. connects using ConnectFromEnv
//  3. optionally runs migrations (MIGRATE_ON_START)
//  4. logs the current database name
func SetupDatabase(ctx context.Context) (*pg.DB, error) {
	// Ensure DB exists, if requested
	if config.GetEnv("DB_CREATE", "false") == "true" {
		if err := ensureDatabase(ctx); err != nil {
			return nil, fmt.Errorf("ensure db: %w", err)
		}
	}

	// Connect to DB via adapter
	db, err := pg.ConnectFromEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	// Run migrations, if requested
	if config.GetEnv("MIGRATE_ON_START", "false") == "true" {
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			config.GetEnv("DB_USER", "postgres"),
			config.GetEnv("DB_PASSWORD", ""),
			config.GetEnv("DB_HOST", "localhost"),
			config.GetEnv("DB_PORT", "5432"),
			config.GetEnv("DB_NAME", "kleff_auth"),
		)

		if err := migrations.Run(ctx, dsn); err != nil {
			// clean up on failure
			db.Pool.Close()
			return nil, fmt.Errorf("db migrate: %w", err)
		}
		log.Printf("Migrations applied (or already up-to-date)")
	}

	// Log current database name (nice debugging info)
	var dbname string
	_ = db.Pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbname)
	log.Printf("Connected to DB: %s", dbname)

	return db, nil
}

// ensureDatabase contains the old logic from db_create.go,
// but kept private and local to the bootstrap package.
func ensureDatabase(ctx context.Context) error {
	host := config.GetEnv("DB_HOST", "localhost")
	port := config.GetEnv("DB_PORT", "5432")
	user := config.GetEnv("DB_USER", "postgres")
	pass := config.GetEnv("DB_PASSWORD", "")
	name := config.GetEnv("DB_NAME", "kleff_auth")

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

	sql := `CREATE DATABASE ` + pgx.Identifier{name}.Sanitize() +
		` OWNER ` + pgx.Identifier{user}.Sanitize() +
		` TEMPLATE template0 ENCODING 'UTF8'`

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
