package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rh-utcp/rh-utcp/internal/config"
	"github.com/rh-utcp/rh-utcp/internal/providers/jira"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

var cfg *config.Config

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

	// Initialize Gin router
	r := gin.Default()

	// UTCP discovery endpoint
	r.GET("/utcp", handleUTCPDiscovery)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	log.Printf("Starting UTCP discovery server on port %s", cfg.Server.Port)
	log.Printf("Discovery endpoint: http://localhost:%s/utcp", cfg.Server.Port)
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Log level: %s", cfg.Server.LogLevel)
	log.Printf("Configured providers: %d", len(cfg.Providers))

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleUTCPDiscovery(c *gin.Context) {
	manual := utcp.NewManual()

	// Initialize providers based on configuration
	for _, providerConfig := range cfg.Providers {
		if !providerConfig.Enabled {
			continue
		}

		log.Printf("Loading provider: %s (%s)", providerConfig.Name, providerConfig.Type)

		switch providerConfig.Type {
		case "jira":
			jiraProvider := jira.NewProvider(
				providerConfig.BaseURL,
				providerConfig.Auth.Username,
				providerConfig.Auth.Password,
			)

			for _, tool := range jiraProvider.GetTools() {
				manual.AddTool(tool)
			}

		// TODO: Add other provider types as they are implemented
		// case "confluence", "wiki":
		//     wikiProvider := wiki.NewProvider(providerConfig.BaseURL, providerConfig.Auth.APIKey)
		//     for _, tool := range wikiProvider.GetTools() {
		//         manual.AddTool(tool)
		//     }
		// case "gitlab":
		//     gitlabProvider := gitlab.NewProvider(providerConfig.BaseURL, providerConfig.Auth.Token)
		//     for _, tool := range gitlabProvider.GetTools() {
		//         manual.AddTool(tool)
		//     }

		default:
			log.Printf("Unknown provider type: %s", providerConfig.Type)
		}
	}

	log.Printf("Loaded %d tools", len(manual.Tools))

	// Return the UTCP manual
	c.JSON(http.StatusOK, manual)
}
