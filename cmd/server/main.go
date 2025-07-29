package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rh-utcp/rh-utcp/internal/config"
	"github.com/rh-utcp/rh-utcp/internal/providers"
	"github.com/rh-utcp/rh-utcp/internal/providers/gitlab"
	"github.com/rh-utcp/rh-utcp/internal/providers/jira"
	"github.com/rh-utcp/rh-utcp/internal/providers/wiki"
	"github.com/rh-utcp/rh-utcp/pkg/errors"
	"github.com/rh-utcp/rh-utcp/pkg/logger"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

var (
	cfg      *config.Config
	registry *providers.Registry
	log      logger.Logger
)

func main() {
	// Initialize logger
	log = logger.New(logger.Config{
		Level:    "info",
		UseColor: true,
	})

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Debug("No .env file found, using system environment variables")
	}

	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Update logger level from config
	log = logger.New(logger.Config{
		Level:    cfg.Server.LogLevel,
		UseColor: true,
	})
	logger.SetGlobal(log.(*logger.StructuredLogger))

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.WithError(err).Fatal("Invalid configuration")
	}

	// Initialize provider registry
	registry = providers.NewRegistry()

	// Register provider factories
	if err := registerProviderFactories(); err != nil {
		log.WithError(err).Fatal("Failed to register provider factories")
	}

	// Create providers from configuration
	if err := createProviders(); err != nil {
		log.WithError(err).Fatal("Failed to create providers")
	}

	// Initialize Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Add logging middleware
	r.Use(ginLogger())
	r.Use(gin.Recovery())

	// UTCP discovery endpoint
	r.GET("/utcp", handleUTCPDiscovery)

	// Health check endpoint
	r.GET("/health", handleHealth)

	// Start server
	log.WithFields(map[string]interface{}{
		"port":        cfg.Server.Port,
		"environment": cfg.Server.Environment,
		"providers":   len(cfg.Providers),
		"enabled":     len(registry.GetEnabledProviders()),
	}).Info("Starting UTCP discovery server")

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

func registerProviderFactories() error {
	// Register Jira provider factory
	if err := registry.RegisterFactory("jira", jira.NewProviderFromConfig); err != nil {
		return errors.Wrap(err, errors.ErrorTypeConfiguration, "failed to register jira factory")
	}

	// Register Wiki/Confluence provider factory
	if err := registry.RegisterFactory("wiki", wiki.NewProviderFromConfig); err != nil {
		return errors.Wrap(err, errors.ErrorTypeConfiguration, "failed to register wiki factory")
	}
	if err := registry.RegisterFactory("confluence", wiki.NewProviderFromConfig); err != nil {
		return errors.Wrap(err, errors.ErrorTypeConfiguration, "failed to register confluence factory")
	}

	// Register GitLab provider factory
	if err := registry.RegisterFactory("gitlab", gitlab.NewProviderFromConfig); err != nil {
		return errors.Wrap(err, errors.ErrorTypeConfiguration, "failed to register gitlab factory")
	}

	log.Debug("Registered provider factories: jira, wiki, confluence, gitlab")
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
			log.WithError(err).WithFields(map[string]interface{}{
				"provider": providerConfig.Name,
				"type":     providerConfig.Type,
			}).Error("Failed to create provider")
			// Continue with other providers
		} else {
			log.WithFields(map[string]interface{}{
				"provider": providerConfig.Name,
				"type":     providerConfig.Type,
				"enabled":  providerConfig.Enabled,
			}).Info("Created provider")
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

	log.WithFields(map[string]interface{}{
		"tools":     len(tools),
		"providers": len(registry.GetEnabledProviders()),
		"ip":        c.ClientIP(),
		"userAgent": c.GetHeader("User-Agent"),
	}).Info("Serving UTCP discovery")

	// Return the UTCP manual
	c.JSON(http.StatusOK, manual)
}

func handleHealth(c *gin.Context) {
	enabledProviders := registry.GetEnabledProviders()
	providerStatus := make(map[string]string)

	for _, provider := range enabledProviders {
		providerStatus[provider.GetName()] = "healthy"
	}

	health := gin.H{
		"status": "ok",
		"providers": gin.H{
			"total":   len(cfg.Providers),
			"enabled": len(enabledProviders),
			"status":  providerStatus,
		},
		"server": gin.H{
			"environment": cfg.Server.Environment,
			"version":     "0.1.0",
		},
	}

	c.JSON(http.StatusOK, health)
}

// ginLogger creates a Gin middleware for logging
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log the request
		fields := map[string]interface{}{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    c.Writer.Status(),
			"ip":        c.ClientIP(),
			"userAgent": c.GetHeader("User-Agent"),
			"size":      c.Writer.Size(),
		}

		if c.Writer.Status() >= 400 {
			log.WithFields(fields).Error("Request failed")
		} else {
			log.WithFields(fields).Info("Request completed")
		}
	}
}
