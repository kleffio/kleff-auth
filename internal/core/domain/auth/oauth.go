package auth

import "time"

type OAuthIdentity struct {
	TenantSlug string
	Provider   string

	ProviderUserID string

	Email    *string
	Username *string

	Attrs map[string]any
}

type OAuthProviderConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

type OAuthClient struct {
	ID           string
	TenantID     string
	ClientID     string
	Name         string
	RedirectURIs []string
	Providers    map[string]OAuthProviderConfig
}

func (c *OAuthClient) AllowsRedirect(uri string) bool {
	for _, u := range c.RedirectURIs {
		if u == uri {
			return true
		}
	}
	return false
}

type OAuthState struct {
	TenantID    string
	TenantSlug  string
	ClientID    string
	RedirectURI string
	Nonce       string
	IssuedAt    time.Time
}
