package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/universal-tool-calling-protocol/go-utcp"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/http"
)

func main() {
	// Create UTCP client configuration
	config := &utcp.UtcpClientConfig{
		Variables: map[string]string{
			// These will be used to replace $JIRA_USERNAME and $JIRA_PASSWORD in the tool definitions
			"JIRA_USERNAME": os.Getenv("JIRA_USERNAME"),
			"JIRA_PASSWORD": os.Getenv("JIRA_PASSWORD"),
		},
	}

	// Create UTCP client
	client := utcp.NewUtcpClient(config)

	// Create HTTP provider for our discovery server
	provider := &http.HttpProvider{
		BaseProvider: utcp.BaseProvider{
			Name:         "rh-utcp",
			ProviderType: "http",
		},
		HTTPMethod: "GET",
		URL:        "http://localhost:8080/utcp",
	}

	// Register the provider to discover available tools
	ctx := context.Background()
	tools, err := client.RegisterProvider(ctx, provider)
	if err != nil {
		log.Fatal("Failed to register provider:", err)
	}

	fmt.Printf("Discovered %d tools:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}

	// Example: Search for Jira issues
	fmt.Println("\n--- Searching Jira Issues ---")

	params := map[string]interface{}{
		"jql":        "project = RHEL AND status = Open",
		"maxResults": 10,
	}

	result, err := client.CallTool(ctx, "jira_search_issues", params)
	if err != nil {
		log.Printf("Error calling jira_search_issues: %v", err)
	} else {
		fmt.Printf("Search result: %v\n", result)
	}

	// Example: Get specific issue details
	fmt.Println("\n--- Getting Issue Details ---")

	issueParams := map[string]interface{}{
		"issueKey": "RHEL-12345", // Replace with actual issue key
	}

	issueResult, err := client.CallTool(ctx, "jira_get_issue", issueParams)
	if err != nil {
		log.Printf("Error calling jira_get_issue: %v", err)
	} else {
		fmt.Printf("Issue details: %v\n", issueResult)
	}

	// Example: Create a new issue
	fmt.Println("\n--- Creating New Issue ---")

	createParams := map[string]interface{}{
		"project":     "RHEL",
		"summary":     "Test issue from UTCP client",
		"description": "This is a test issue created via UTCP",
		"issuetype":   "Task",
		"priority":    "Medium",
	}

	createResult, err := client.CallTool(ctx, "jira_create_issue", createParams)
	if err != nil {
		log.Printf("Error calling jira_create_issue: %v", err)
	} else {
		fmt.Printf("Created issue: %v\n", createResult)
	}
}
