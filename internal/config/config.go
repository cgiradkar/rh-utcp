package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Providers []ProviderConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port        string
	Environment string
	LogLevel    string
}

// ProviderConfig holds configuration for a single provider
type ProviderConfig struct {
	Name    string
	Type    string
	Enabled bool
	BaseURL string
	Auth    AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type         string
	Username     string
	Password     string
	APIKey       string
	Token        string
	ClientID     string
	ClientSecret string
	TokenURL     string
}

// Load loads configuration from environment and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.environment", "development")
	v.SetDefault("server.loglevel", "info")

	// Set config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/rh-utcp/")

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		// It's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Enable environment variables
	v.SetEnvPrefix("RHUTCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Build configuration from environment
	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnvOrDefault("PORT", v.GetString("server.port")),
			Environment: v.GetString("server.environment"),
			LogLevel:    v.GetString("server.loglevel"),
		},
		Providers: []ProviderConfig{},
	}

	// Load Jira provider if configured
	if jiraURL := os.Getenv("JIRA_BASE_URL"); jiraURL != "" {
		cfg.Providers = append(cfg.Providers, ProviderConfig{
			Name:    "jira",
			Type:    "jira",
			Enabled: true,
			BaseURL: jiraURL,
			Auth: AuthConfig{
				Type:     "basic",
				Username: os.Getenv("JIRA_USERNAME"),
				Password: os.Getenv("JIRA_PASSWORD"),
			},
		})
	}

	// Load Wiki provider if configured
	if wikiURL := os.Getenv("WIKI_BASE_URL"); wikiURL != "" {
		cfg.Providers = append(cfg.Providers, ProviderConfig{
			Name:    "wiki",
			Type:    "confluence",
			Enabled: true,
			BaseURL: wikiURL,
			Auth: AuthConfig{
				Type:   "api_key",
				APIKey: os.Getenv("WIKI_API_KEY"),
			},
		})
	}

	// Load GitLab provider if configured
	if gitlabURL := os.Getenv("GITLAB_BASE_URL"); gitlabURL != "" {
		cfg.Providers = append(cfg.Providers, ProviderConfig{
			Name:    "gitlab",
			Type:    "gitlab",
			Enabled: true,
			BaseURL: gitlabURL,
			Auth: AuthConfig{
				Type:  "personal_token",
				Token: os.Getenv("GITLAB_TOKEN"),
			},
		})
	}

	// Load providers from config file if any
	if v.IsSet("providers") {
		var fileProviders []ProviderConfig
		if err := v.UnmarshalKey("providers", &fileProviders); err != nil {
			return nil, fmt.Errorf("error unmarshaling providers: %w", err)
		}

		// Merge with environment-based providers
		for _, fp := range fileProviders {
			// Skip if already loaded from environment
			exists := false
			for _, ep := range cfg.Providers {
				if ep.Name == fp.Name {
					exists = true
					break
				}
			}
			if !exists {
				cfg.Providers = append(cfg.Providers, fp)
			}
		}
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Validate providers
	for _, p := range c.Providers {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("provider %s: %w", p.Name, err)
		}
	}

	return nil
}

// Validate validates a provider configuration
func (p *ProviderConfig) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	if p.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	if p.Enabled && p.BaseURL == "" {
		return fmt.Errorf("base URL is required for enabled provider")
	}

	// Validate auth based on type
	if p.Enabled {
		switch p.Auth.Type {
		case "basic":
			if p.Auth.Username == "" || p.Auth.Password == "" {
				return fmt.Errorf("username and password required for basic auth")
			}
		case "api_key":
			if p.Auth.APIKey == "" {
				return fmt.Errorf("API key required for api_key auth")
			}
		case "personal_token":
			if p.Auth.Token == "" {
				return fmt.Errorf("token required for personal_token auth")
			}
		case "oauth2":
			if p.Auth.ClientID == "" || p.Auth.ClientSecret == "" || p.Auth.TokenURL == "" {
				return fmt.Errorf("client_id, client_secret, and token_url required for oauth2 auth")
			}
		}
	}

	return nil
}

// GetProvider returns a provider configuration by name
func (c *Config) GetProvider(name string) (*ProviderConfig, bool) {
	for _, p := range c.Providers {
		if p.Name == name {
			return &p, true
		}
	}
	return nil, false
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
