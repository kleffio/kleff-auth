package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kleffio/kleff-auth/internal/infrastructure/bootstrap"
	"github.com/kleffio/kleff-auth/internal/utils"

	cryptoad "github.com/kleffio/kleff-auth/internal/adapters/crypto"
	httpad "github.com/kleffio/kleff-auth/internal/adapters/http"
	pg "github.com/kleffio/kleff-auth/internal/adapters/postgres"
	app "github.com/kleffio/kleff-auth/internal/application/auth"
	migration "github.com/kleffio/kleff-auth/internal/infrastructure/db"
)

func main() {
	utils.LoadEnv()
	ctx := context.Background()

	//-- Bootstrap --\\

	if utils.GetEnv("DB_CREATE", "false") == "true" {
		if err := bootstrap.EnsureDatabase(ctx); err != nil {
			log.Fatalf("ensure db: %v", err)
		}
	}

	//-- DB Connections --\\

	db, err := pg.ConnectFromEnv(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Pool.Close()

	if utils.GetEnv("MIGRATE_ON_START", "false") == "true" {
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			utils.GetEnv("DB_USER", "postgres"),
			utils.GetEnv("DB_PASSWORD", ""),
			utils.GetEnv("DB_HOST", "localhost"),
			utils.GetEnv("DB_PORT", "5432"),
			utils.GetEnv("DB_NAME", "kleff_auth"),
		)
		if err := migration.Run(ctx, dsn); err != nil {
			log.Fatalf("db: %v", err)
		}
		log.Printf("Migrations applied (or already up-to-date)")
	}

	var dbname string
	_ = db.Pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbname)
	log.Printf("Connected to DB: %s", dbname)

	//-- Crypto --\\

	issuer := utils.GetEnv("JWT_ISSUER", "http://localhost:8080")
	signer, err := cryptoad.NewInMemorySigner(issuer)
	if err != nil {
		log.Fatalf("signer: %v", err)
	}

	hasher := cryptoad.NewArgon2id()
	refreshCodec := &cryptoad.RefreshCodec{Hasher: hasher}

	//-- Repos --\\

	tenantRepo := pg.NewTenantRepo(db)
	userRepo := pg.NewUserRepo(db)
	sessionRepo := pg.NewSessionRepo(db)

	//-- Application service --\\

	svc := &app.Service{
		Tenants:  tenantRepo,
		Users:    userRepo,
		Hash:     hasher,
		Tokens:   signer,
		Sessions: sessionRepo,
		Refresh:  refreshCodec,

		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	}

	//-- HTTP server --\\

	mux := httpad.NewRouter(svc)

	addr := utils.GetEnv("HTTP_ADDR", ":8080")
	log.Printf("authd listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
