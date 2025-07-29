package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save current environment
	oldJiraURL := os.Getenv("JIRA_BASE_URL")
	oldJiraUser := os.Getenv("JIRA_USERNAME")
	oldJiraPass := os.Getenv("JIRA_PASSWORD")
	oldWikiURL := os.Getenv("WIKI_BASE_URL")
	oldWikiKey := os.Getenv("WIKI_API_KEY")
	oldGitLabURL := os.Getenv("GITLAB_BASE_URL")
	oldGitLabToken := os.Getenv("GITLAB_TOKEN")
	oldPort := os.Getenv("PORT")

	// Restore environment after test
	defer func() {
		os.Setenv("JIRA_BASE_URL", oldJiraURL)
		os.Setenv("JIRA_USERNAME", oldJiraUser)
		os.Setenv("JIRA_PASSWORD", oldJiraPass)
		os.Setenv("WIKI_BASE_URL", oldWikiURL)
		os.Setenv("WIKI_API_KEY", oldWikiKey)
		os.Setenv("GITLAB_BASE_URL", oldGitLabURL)
		os.Setenv("GITLAB_TOKEN", oldGitLabToken)
		os.Setenv("PORT", oldPort)
	}()

	t.Run("Default configuration", func(t *testing.T) {
		// Clear all environment variables
		os.Unsetenv("JIRA_BASE_URL")
		os.Unsetenv("WIKI_BASE_URL")
		os.Unsetenv("GITLAB_BASE_URL")
		os.Unsetenv("PORT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Check defaults
		if cfg.Server.Port != "8080" {
			t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
		}

		if cfg.Server.Environment != "development" {
			t.Errorf("Expected default environment 'development', got %s", cfg.Server.Environment)
		}

		if cfg.Server.LogLevel != "info" {
			t.Errorf("Expected default log level 'info', got %s", cfg.Server.LogLevel)
		}

		if len(cfg.Providers) != 0 {
			t.Errorf("Expected no providers, got %d", len(cfg.Providers))
		}
	})

	t.Run("Load from environment", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "9090")
		os.Setenv("JIRA_BASE_URL", "https://jira.test.com")
		os.Setenv("JIRA_USERNAME", "testuser")
		os.Setenv("JIRA_PASSWORD", "testpass")
		os.Setenv("WIKI_BASE_URL", "https://wiki.test.com")
		os.Setenv("WIKI_API_KEY", "wiki-key-123")
		os.Setenv("GITLAB_BASE_URL", "https://gitlab.test.com")
		os.Setenv("GITLAB_TOKEN", "gitlab-token-456")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Check server config
		if cfg.Server.Port != "9090" {
			t.Errorf("Expected port 9090, got %s", cfg.Server.Port)
		}

		// Check providers
		if len(cfg.Providers) != 3 {
			t.Fatalf("Expected 3 providers, got %d", len(cfg.Providers))
		}

		// Check Jira provider
		jiraProvider, found := cfg.GetProvider("jira")
		if !found {
			t.Fatal("Jira provider not found")
		}

		if jiraProvider.Type != "jira" {
			t.Errorf("Expected Jira type 'jira', got %s", jiraProvider.Type)
		}

		if jiraProvider.BaseURL != "https://jira.test.com" {
			t.Errorf("Expected Jira URL 'https://jira.test.com', got %s", jiraProvider.BaseURL)
		}

		if jiraProvider.Auth.Type != "basic" {
			t.Errorf("Expected Jira auth type 'basic', got %s", jiraProvider.Auth.Type)
		}

		if jiraProvider.Auth.Username != "testuser" {
			t.Errorf("Expected Jira username 'testuser', got %s", jiraProvider.Auth.Username)
		}

		// Check Wiki provider
		wikiProvider, found := cfg.GetProvider("wiki")
		if !found {
			t.Fatal("Wiki provider not found")
		}

		if wikiProvider.Type != "confluence" {
			t.Errorf("Expected Wiki type 'confluence', got %s", wikiProvider.Type)
		}

		if wikiProvider.Auth.Type != "api_key" {
			t.Errorf("Expected Wiki auth type 'api_key', got %s", wikiProvider.Auth.Type)
		}

		if wikiProvider.Auth.APIKey != "wiki-key-123" {
			t.Errorf("Expected Wiki API key 'wiki-key-123', got %s", wikiProvider.Auth.APIKey)
		}

		// Check GitLab provider
		gitlabProvider, found := cfg.GetProvider("gitlab")
		if !found {
			t.Fatal("GitLab provider not found")
		}

		if gitlabProvider.Type != "gitlab" {
			t.Errorf("Expected GitLab type 'gitlab', got %s", gitlabProvider.Type)
		}

		if gitlabProvider.Auth.Type != "personal_token" {
			t.Errorf("Expected GitLab auth type 'personal_token', got %s", gitlabProvider.Auth.Type)
		}

		if gitlabProvider.Auth.Token != "gitlab-token-456" {
			t.Errorf("Expected GitLab token 'gitlab-token-456', got %s", gitlabProvider.Auth.Token)
		}
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid configuration",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "jira",
						Type:    "jira",
						Enabled: true,
						BaseURL: "https://jira.example.com",
						Auth: AuthConfig{
							Type:     "basic",
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing server port",
			config: Config{
				Server: ServerConfig{
					Port: "",
				},
			},
			wantErr: true,
			errMsg:  "server port is required",
		},
		{
			name: "Provider missing name",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name: "",
						Type: "jira",
					},
				},
			},
			wantErr: true,
			errMsg:  "provider name is required",
		},
		{
			name: "Provider missing type",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name: "test",
						Type: "",
					},
				},
			},
			wantErr: true,
			errMsg:  "provider type is required",
		},
		{
			name: "Enabled provider missing base URL",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "test",
						Type:    "jira",
						Enabled: true,
						BaseURL: "",
					},
				},
			},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name: "Basic auth missing credentials",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "test",
						Type:    "jira",
						Enabled: true,
						BaseURL: "https://example.com",
						Auth: AuthConfig{
							Type:     "basic",
							Username: "",
							Password: "pass",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "username and password required",
		},
		{
			name: "API key auth missing key",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "test",
						Type:    "wiki",
						Enabled: true,
						BaseURL: "https://example.com",
						Auth: AuthConfig{
							Type:   "api_key",
							APIKey: "",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "API key required",
		},
		{
			name: "Personal token auth missing token",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "test",
						Type:    "gitlab",
						Enabled: true,
						BaseURL: "https://example.com",
						Auth: AuthConfig{
							Type:  "personal_token",
							Token: "",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "token required",
		},
		{
			name: "OAuth2 missing client ID",
			config: Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Providers: []ProviderConfig{
					{
						Name:    "test",
						Type:    "oauth",
						Enabled: true,
						BaseURL: "https://example.com",
						Auth: AuthConfig{
							Type:         "oauth2",
							ClientID:     "",
							ClientSecret: "secret",
							TokenURL:     "https://auth.example.com/token",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "client_id, client_secret, and token_url required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	cfg := &Config{
		Providers: []ProviderConfig{
			{Name: "jira", Type: "jira"},
			{Name: "wiki", Type: "confluence"},
			{Name: "gitlab", Type: "gitlab"},
		},
	}

	tests := []struct {
		name     string
		provider string
		found    bool
	}{
		{"Find existing Jira", "jira", true},
		{"Find existing Wiki", "wiki", true},
		{"Find existing GitLab", "gitlab", true},
		{"Find non-existent", "bitbucket", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, found := cfg.GetProvider(tt.provider)

			if found != tt.found {
				t.Errorf("Expected found=%v, got %v", tt.found, found)
			}

			if found && provider.Name != tt.provider {
				t.Errorf("Expected provider name %s, got %s", tt.provider, provider.Name)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Save and restore environment
	oldValue := os.Getenv("TEST_ENV_VAR")
	defer os.Setenv("TEST_ENV_VAR", oldValue)

	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable set",
			envValue:     "env-value",
			defaultValue: "default-value",
			expected:     "env-value",
		},
		{
			name:         "Environment variable empty",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_ENV_VAR", tt.envValue)
			} else {
				os.Unsetenv("TEST_ENV_VAR")
			}

			result := getEnvOrDefault("TEST_ENV_VAR", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[0:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(substr) < len(s) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
