package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RuntimeConfig struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`

	Database struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		Name            string `yaml:"name"`
		User            string `yaml:"user"`
		Pass            string `yaml:"pass"`
		RunMigrations   bool   `yaml:"run_migrations"`
		CreateIfMissing bool   `yaml:"create_if_missing"`
	} `yaml:"database"`

	JWT struct {
		Issuer string `yaml:"issuer"`
	} `yaml:"jwt"`

	OAuth struct {
		StateKey string `yaml:"state_key"`
	} `yaml:"oauth"`

	Clients []OAuthClientConfig `yaml:"clients"`
}

type OAuthClientConfig struct {
	TenantSlug   string                         `yaml:"tenant_slug"`
	ClientID     string                         `yaml:"client_id"`
	DisplayName  string                         `yaml:"display_name"`
	RedirectURIs []string                       `yaml:"redirect_uris"`
	Providers    map[string]OAuthProviderConfig `yaml:"providers"`
}

type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

func LoadRuntimeConfig(path string) (*RuntimeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data = []byte(os.ExpandEnv(string(data)))

	var cfg RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
