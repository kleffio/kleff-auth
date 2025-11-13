package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

const configRootEnv = "KLEFF_CONFIG_ROOT"

func getConfigRoot() (string, error) {
	root := os.Getenv(configRootEnv)
	if root == "" {
		root = "."
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve config root: %w", err)
	}

	return abs, nil
}

func LoadRuntimeConfig(path string) (*RuntimeConfig, error) {
	configRoot, err := getConfigRoot()
	if err != nil {
		return nil, err
	}

	rel := filepath.Clean(path)

	targetAbs, err := filepath.Abs(filepath.Join(configRoot, rel))
	if err != nil {
		return nil, fmt.Errorf("resolve config path %q: %w", path, err)
	}

	if !strings.HasPrefix(targetAbs, configRoot+string(os.PathSeparator)) && targetAbs != configRoot {
		return nil, fmt.Errorf("config path %q is outside allowed root %q", targetAbs, configRoot)
	}

	data, err := os.ReadFile(targetAbs)
	if err != nil {
		return nil, fmt.Errorf("read runtime config %q: %w", targetAbs, err)
	}

	data = []byte(os.ExpandEnv(string(data)))

	var cfg RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal runtime config %q: %w", targetAbs, err)
	}

	return &cfg, nil
}
