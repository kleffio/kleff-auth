// internal/infrastructure/bootstrap/app.go
package bootstrap

import (
	"context"
	"log"
	"net/http"
	"time"

	httpad "github.com/kleffio/kleff-auth/internal/adapters/in/http"
	cryptoad "github.com/kleffio/kleff-auth/internal/adapters/out/crypto"
	pg "github.com/kleffio/kleff-auth/internal/adapters/out/postgres"
	app "github.com/kleffio/kleff-auth/internal/application/auth"
	"github.com/kleffio/kleff-auth/internal/utils"
)

type App struct {
	Server *http.Server
	DB     *pg.DB
}

// NewApp wires DB, crypto, repos, service, and HTTP server.
func NewApp(ctx context.Context) (*App, error) {
	// Load env once here
	utils.LoadEnv()

	// --- DB bootstrap --- //

	db, err := SetupDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// --- Crypto --- //

	issuer := utils.GetEnv("JWT_ISSUER", "http://localhost:8080")
	signer, err := cryptoad.NewInMemorySigner(issuer)
	if err != nil {
		db.Pool.Close()
		return nil, err
	}

	hasher := cryptoad.NewArgon2id()
	refreshCodec := &cryptoad.RefreshCodec{Hasher: hasher}

	// --- Repos --- //
	tenantRepo := pg.NewTenantRepo(db)
	userRepo := pg.NewUserRepo(db)
	sessionRepo := pg.NewSessionRepo(db)

	// --- Application service --- //
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

	// --- HTTP server --- //
	mux := httpad.NewRouter(svc)
	addr := utils.GetEnv("HTTP_ADDR", ":8080")

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("authd listening on %s", addr)

	return &App{
		Server: srv,
		DB:     db,
	}, nil
}

// Run starts the HTTP server.
func (a *App) Run() error {
	return a.Server.ListenAndServe()
}

// Close cleanly shuts down shared resources like the DB pool.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Pool.Close()
	}
}
