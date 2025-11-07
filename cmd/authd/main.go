package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kleffio/kleff-auth/internal/utils"

	cryptoad "github.com/kleffio/kleff-auth/internal/adapters/crypto"
	httpad "github.com/kleffio/kleff-auth/internal/adapters/http"
	pg "github.com/kleffio/kleff-auth/internal/adapters/postgres"
	app "github.com/kleffio/kleff-auth/internal/application/auth"
)

func main() {
	utils.LoadEnv()

	ctx := context.Background()

	//-- DB --\\

	db, err := pg.ConnectFromEnv(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Pool.Close()

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

	//-- App --\\

	svc := &app.Service{
		Tenants:   pg.NewTenantRepo(db),
		Users:     pg.NewUserRepo(db),
		Hash:      hasher,
		Tokens:    signer,
		AccessTTL: 15 * time.Minute,
	}

	mux := httpad.NewServer(svc)
	addr := utils.GetEnv("HTTP_ADDR", ":8080")
	log.Printf("authd listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
