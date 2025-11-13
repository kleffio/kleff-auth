package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	pg "github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres"
	"github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres/migrations"
	"github.com/kleffio/kleff-auth/internal/config"
)

func SetupDatabase(ctx context.Context, cfg *config.RuntimeConfig) (*pg.DB, error) {
	if cfg.Database.CreateIfMissing {
		if err := ensureDatabase(ctx, cfg); err != nil {
			return nil, fmt.Errorf("ensure db: %w", err)
		}
	}

	dsn := buildPostgresURL(
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := pg.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	if cfg.Database.RunMigrations {
		if err := migrations.Run(ctx, dsn); err != nil {
			db.Pool.Close()
			return nil, fmt.Errorf("migrations failed: %w", err)
		}
		log.Printf("Migrations applied")
	}

	log.Printf("Connected to DB %q on %s:%d", cfg.Database.Name, cfg.Database.Host, cfg.Database.Port)
	return db, nil
}

func ensureDatabase(ctx context.Context, cfg *config.RuntimeConfig) error {
	maintURL := buildPostgresURL(
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Host,
		cfg.Database.Port,
		"postgres",
	)

	conn, err := pgx.Connect(ctx, maintURL)
	if err != nil {
		return fmt.Errorf("connect (maintenance): %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	err = conn.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)`,
		cfg.Database.Name,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	sql := `CREATE DATABASE ` + pgx.Identifier{cfg.Database.Name}.Sanitize() +
		` OWNER ` + pgx.Identifier{cfg.Database.User}.Sanitize() +
		` TEMPLATE template0 ENCODING 'UTF8'`

	_, err = conn.Exec(ctx, sql)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42501" {
			return fmt.Errorf("role %q lacks CREATEDB", cfg.Database.User)
		}
		return err
	}

	log.Printf("Created database %q", cfg.Database.Name)
	return nil
}

func buildPostgresURL(user, pass, host string, port int, dbName string) string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   dbName,
	}

	u.User = url.UserPassword(user, pass)

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String()
}
