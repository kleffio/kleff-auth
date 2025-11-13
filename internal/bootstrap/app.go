package bootstrap

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpad "github.com/kleffio/kleff-auth/internal/adapters/in/http"
	hashargon "github.com/kleffio/kleff-auth/internal/adapters/out/hash/argon2"
	oauthadapter "github.com/kleffio/kleff-auth/internal/adapters/out/oauth"
	pg "github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres"
	tokeneddsa "github.com/kleffio/kleff-auth/internal/adapters/out/token/eddsa"
	tokenrefresh "github.com/kleffio/kleff-auth/internal/adapters/out/token/refresh"
	"github.com/kleffio/kleff-auth/internal/config"
	app "github.com/kleffio/kleff-auth/internal/core/service/auth"
)

type App struct {
	Server *http.Server
	DB     *pg.DB
}

// NewApp wires DB, crypto, repos, service, and HTTP server.
func NewApp(ctx context.Context) (*App, error) {
	// Load env once here (for ${VAR} in config.yaml)
	config.LoadEnv()

	// --- YAML config --- //

	runtimeCfg, err := config.LoadRuntimeConfig("config.yaml")
	if err != nil {
		return nil, err
	}

	if v, ok := os.LookupEnv("DB_PASS"); ok && v != "" {
		runtimeCfg.Database.Pass = v
	}

	// --- DB bootstrap --- //

	db, err := SetupDatabase(ctx, runtimeCfg)
	if err != nil {
		return nil, err
	}

	// --- Config seeding (tenants + oauth clients) --- //

	log.Printf("Running config seeder...")
	if err := SeedFromConfig(ctx, db, runtimeCfg); err != nil {
		db.Pool.Close()
		return nil, err
	}

	// --- Crypto --- //

	issuer := runtimeCfg.JWT.Issuer
	if issuer == "" {
		issuer = "http://localhost:8080"
	}

	signer, err := tokeneddsa.NewInMemorySigner(issuer)
	if err != nil {
		db.Pool.Close()
		return nil, err
	}

	hasher := hashargon.NewArgon2id()
	refreshCodec := &tokenrefresh.Codec{Hasher: hasher}

	// --- OAuth --- //

	oauthSecretKey := runtimeCfg.OAuth.StateKey
	if oauthSecretKey == "" {
		oauthSecretKey = "12345678901234567890123456789012"
	}

	stateCodec, err := oauthadapter.NewStateCodec(oauthSecretKey)
	if err != nil {
		db.Pool.Close()
		return nil, err
	}

	oauthProvider := oauthadapter.NewProvider()
	oauthClientRepo := pg.NewOAuthClientRepo(db)
	oauthUserRepo := pg.NewOAuthUserRepo(db)

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

		OAuthProviders: oauthProvider,
		OAuthClients:   oauthClientRepo,
		OAuthState:     stateCodec,
		OAuthUsers:     oauthUserRepo,

		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	}

	// --- HTTP server --- //

	addr := runtimeCfg.Server.Address
	if addr == "" {
		addr = ":8080"
	}

	mux := httpad.NewRouter(svc)

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
