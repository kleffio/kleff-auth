package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	httpad "github.com/kleffio/kleff-auth/internal/adapters/in/http"
	cryptoad "github.com/kleffio/kleff-auth/internal/adapters/out/crypto"
	pg "github.com/kleffio/kleff-auth/internal/adapters/out/postgres"
	app "github.com/kleffio/kleff-auth/internal/application/auth"
	"github.com/kleffio/kleff-auth/internal/utils"
)

// App represents the fully-wired application.
type App struct {
	Server   *http.Server
	Shutdown func(ctx context.Context) error
}

// NewApp wires the DB, crypto, repositories, service and HTTP server.
func NewApp(ctx context.Context) (*App, error) {
	// Load .env
	utils.LoadEnv()

	// --- DB bootstrap (create + migrate + connect) --- //
	db, err := SetupDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// --- Crypto --- //

	issuer := utils.GetEnv("JWT_ISSUER", "http://localhost:8080")
	signer, err := cryptoad.NewInMemorySigner(issuer)
	if err != nil {
		db.Pool.Close()
		return nil, fmt.Errorf("signer: %w", err)
	}

	hasher := cryptoad.NewArgon2id()
	refreshCodec := &cryptoad.RefreshCodec{Hasher: hasher}

	// --- Repositories --- //

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

	appInstance := &App{
		Server: srv,
		Shutdown: func(_ context.Context) error {
			db.Pool.Close()
			return nil
		},
	}

	return appInstance, nil
}

// Run is a convenience helper to start the HTTP server.
func (a *App) Run() error {
	log.Printf("authd listening on %s", a.Server.Addr)
	return a.Server.ListenAndServe()
}
