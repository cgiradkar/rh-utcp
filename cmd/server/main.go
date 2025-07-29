package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rh-utcp/rh-utcp/internal/providers/jira"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting UTCP discovery server on port %s", port)
	log.Printf("Discovery endpoint: http://localhost:%s/utcp", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleUTCPDiscovery(c *gin.Context) {
	manual := utcp.NewManual()

	// Initialize providers based on environment configuration
	if jiraURL := os.Getenv("JIRA_BASE_URL"); jiraURL != "" {
		log.Println("Adding Jira provider")
		jiraProvider := jira.NewProvider(
			jiraURL,
			os.Getenv("JIRA_USERNAME"),
			os.Getenv("JIRA_PASSWORD"),
		)

		for _, tool := range jiraProvider.GetTools() {
			manual.AddTool(tool)
		}
	}

	// TODO: Add other providers (Wiki, GitLab, etc.) here as they are implemented
	// Example:
	// if wikiURL := os.Getenv("WIKI_BASE_URL"); wikiURL != "" {
	//     wikiProvider := wiki.NewProvider(wikiURL, os.Getenv("WIKI_API_KEY"))
	//     for _, tool := range wikiProvider.GetTools() {
	//         manual.AddTool(tool)
	//     }
	// }

	// Return the UTCP manual
	c.JSON(http.StatusOK, manual)
}
