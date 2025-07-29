package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rh-utcp/rh-utcp/internal/config"
	"github.com/rh-utcp/rh-utcp/internal/providers"
	"github.com/rh-utcp/rh-utcp/internal/providers/gitlab"
	"github.com/rh-utcp/rh-utcp/internal/providers/jira"
	"github.com/rh-utcp/rh-utcp/internal/providers/wiki"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

var (
	cfg      *config.Config
	registry *providers.Registry
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	// Initialize provider registry
	registry = providers.NewRegistry()

	// Register provider factories
	if err := registerProviderFactories(); err != nil {
		log.Fatal("Failed to register provider factories:", err)
	}

	// Create providers from configuration
	if err := createProviders(); err != nil {
		log.Fatal("Failed to create providers:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// UTCP discovery endpoint
	r.GET("/utcp", handleUTCPDiscovery)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"providers": len(registry.GetEnabledProviders()),
		})
	})

	// Start server
	log.Printf("Starting UTCP discovery server on port %s", cfg.Server.Port)
	log.Printf("Discovery endpoint: http://localhost:%s/utcp", cfg.Server.Port)
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Log level: %s", cfg.Server.LogLevel)
	log.Printf("Configured providers: %d", len(cfg.Providers))
	log.Printf("Enabled providers: %d", len(registry.GetEnabledProviders()))

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func registerProviderFactories() error {
	// Register Jira provider factory
	if err := registry.RegisterFactory("jira", jira.NewProviderFromConfig); err != nil {
		return err
	}

	// Register Wiki/Confluence provider factory
	if err := registry.RegisterFactory("wiki", wiki.NewProviderFromConfig); err != nil {
		return err
	}
	if err := registry.RegisterFactory("confluence", wiki.NewProviderFromConfig); err != nil {
		return err
	}

	// Register GitLab provider factory
	if err := registry.RegisterFactory("gitlab", gitlab.NewProviderFromConfig); err != nil {
		return err
	}

	log.Println("Registered provider factories: jira, wiki, confluence, gitlab")
	return nil
}

func createProviders() error {
	for _, providerConfig := range cfg.Providers {
		// Convert config to map for factory
		configMap := map[string]interface{}{
			"name":     providerConfig.Name,
			"enabled":  providerConfig.Enabled,
			"base_url": providerConfig.BaseURL,
		}

		// Add auth configuration based on type
		switch providerConfig.Auth.Type {
		case "basic":
			configMap["username"] = providerConfig.Auth.Username
			configMap["password"] = providerConfig.Auth.Password
		case "api_key":
			configMap["api_key"] = providerConfig.Auth.APIKey
		case "personal_token":
			configMap["token"] = providerConfig.Auth.Token
		case "oauth2":
			configMap["client_id"] = providerConfig.Auth.ClientID
			configMap["client_secret"] = providerConfig.Auth.ClientSecret
			configMap["token_url"] = providerConfig.Auth.TokenURL
		}

		// Create provider
		if err := registry.CreateProvider(providerConfig.Name, providerConfig.Type, configMap); err != nil {
			log.Printf("Failed to create provider %s: %v", providerConfig.Name, err)
			// Continue with other providers
		} else {
			log.Printf("Created provider: %s (%s)", providerConfig.Name, providerConfig.Type)
		}
	}

	return nil
}

func handleUTCPDiscovery(c *gin.Context) {
	manual := utcp.NewManual()

	// Get all tools from enabled providers
	tools := registry.GetAllTools()
	for _, tool := range tools {
		manual.AddTool(tool)
	}

	log.Printf("Serving %d tools from %d enabled providers", len(tools), len(registry.GetEnabledProviders()))

	// Return the UTCP manual
	c.JSON(http.StatusOK, manual)
}
