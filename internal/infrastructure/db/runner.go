package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var FS embed.FS

func Run(ctx context.Context, dsn string) error {
	goose.SetBaseFS(FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open sql: %w", err)
	}

	defer db.Close()

	const embeddedDir = "migrations"

	if err := goose.Up(db, embeddedDir); err != nil && err != goose.ErrNoCurrentVersion {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
