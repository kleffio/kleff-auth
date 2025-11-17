package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5"
	pg "github.com/kleffio/kleff-auth/internal/adapters/out/repository/postgres"
	"github.com/kleffio/kleff-auth/internal/config"
)

func SeedFromConfig(ctx context.Context, db *pg.DB, cfg *config.RuntimeConfig) error {
	if cfg == nil {
		return nil
	}

	for _, c := range cfg.Clients {
		if err := upsertClientFromConfig(ctx, db, &c); err != nil {
			return err
		}
	}

	return nil
}

func upsertClientFromConfig(ctx context.Context, db *pg.DB, c *config.OAuthClientConfig) error {
	if c.TenantSlug == "" || c.ClientID == "" || c.DisplayName == "" {
		return fmt.Errorf("invalid client config: tenant_slug, client_id and display_name are required")
	}
	if len(c.RedirectURIs) == 0 {
		return fmt.Errorf("client %s: at least one redirect_uri is required", c.ClientID)
	}

	var tenantID string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, name)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id::text;
	`, c.TenantSlug, c.TenantSlug).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("ensure tenant '%s': %w", c.TenantSlug, err)
	}

	providers := make(map[string]interface{}, len(c.Providers))
	for name, pcfg := range c.Providers {
		providers[name] = map[string]interface{}{
			"client_id":     pcfg.ClientID,
			"client_secret": pcfg.ClientSecret,
			"redirect_url":  pcfg.RedirectURL,
			"scopes":        pcfg.Scopes,
		}
	}

	redirectJSON, _ := json.Marshal(c.RedirectURIs)
	providersJSON, _ := json.Marshal(providers)

	query := `
INSERT INTO oauth_clients (tenant_id, client_id, name, redirect_uris, providers)
VALUES ($1::uuid, $2, $3, $4::jsonb, $5::jsonb)
ON CONFLICT (tenant_id, client_id) 
DO UPDATE SET 
	name = EXCLUDED.name,
	redirect_uris = EXCLUDED.redirect_uris,
	providers = oauth_clients.providers || EXCLUDED.providers,
	updated_at = NOW()
RETURNING id;`

	var id string
	err = db.Pool.QueryRow(ctx, query, tenantID, c.ClientID, c.DisplayName, redirectJSON, providersJSON).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert oauth client '%s': %w", c.ClientID, err)
	}

	log.Printf("Seeded OAuth client %q for tenant %q (id=%s)", c.ClientID, c.TenantSlug, id)
	return nil
}
