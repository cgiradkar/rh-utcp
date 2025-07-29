package jira

import (
	"strings"
	"testing"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

func TestNewProvider(t *testing.T) {
	baseURL := "https://jira.example.com"
	username := "testuser"
	password := "testpass"

	provider := NewProvider(baseURL, username, password)

	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	if provider.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, provider.BaseURL)
	}

	if provider.Username != username {
		t.Errorf("Expected Username %s, got %s", username, provider.Username)
	}

	if provider.Password != password {
		t.Errorf("Expected Password %s, got %s", password, provider.Password)
	}
}

func TestGetTools(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	if len(tools) == 0 {
		t.Fatal("GetTools returned empty slice")
	}

	// Expected tools
	expectedTools := map[string]bool{
		"jira_search_issues":   false,
		"jira_get_issue":       false,
		"jira_create_issue":    false,
		"jira_update_issue":    false,
		"jira_get_projects":    false,
		"jira_add_comment":     false,
		"jira_get_user_issues": false,
	}

	// Check all expected tools are present
	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		} else {
			t.Errorf("Unexpected tool: %s", tool.Name)
		}
	}

	// Verify all expected tools were found
	for toolName, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool not found: %s", toolName)
		}
	}
}

func TestJiraSearchIssuesTool(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	// Find the search issues tool
	var searchTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "jira_search_issues" {
			searchTool = &tool
			break
		}
	}

	if searchTool == nil {
		t.Fatal("jira_search_issues tool not found")
	}

	// Test basic properties
	if searchTool.Description != "Search for Jira issues using JQL (Jira Query Language)" {
		t.Errorf("Unexpected description: %s", searchTool.Description)
	}

	// Test inputs
	if searchTool.Inputs.Type != "object" {
		t.Errorf("Expected inputs type 'object', got %s", searchTool.Inputs.Type)
	}

	// Check required fields
	if len(searchTool.Inputs.Required) != 1 || searchTool.Inputs.Required[0] != "jql" {
		t.Error("Expected 'jql' to be the only required field")
	}

	// Check fields property
	fieldsProperty, exists := searchTool.Inputs.Properties["fields"]
	if !exists {
		t.Error("'fields' property missing")
	} else {
		if fieldsProperty.Type != "array" {
			t.Errorf("Expected 'fields' to be array type, got %s", fieldsProperty.Type)
		}
	}

	// Check defaults
	maxResultsProperty := searchTool.Inputs.Properties["maxResults"]
	if maxResultsProperty.Default != 50 {
		t.Errorf("Expected maxResults default 50, got %v", maxResultsProperty.Default)
	}

	// Test provider configuration
	providerConfig := searchTool.ToolProvider

	if providerConfig["http_method"] != "GET" {
		t.Errorf("Expected http_method 'GET', got %v", providerConfig["http_method"])
	}
}

func TestJiraGetIssueTool(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	// Find the get issue tool
	var getTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "jira_get_issue" {
			getTool = &tool
			break
		}
	}

	if getTool == nil {
		t.Fatal("jira_get_issue tool not found")
	}

	// Test basic properties
	if getTool.Description != "Get detailed information about a specific Jira issue" {
		t.Errorf("Unexpected description: %s", getTool.Description)
	}

	// Check required fields
	if len(getTool.Inputs.Required) != 1 || getTool.Inputs.Required[0] != "issueKey" {
		t.Error("Expected 'issueKey' to be the only required field")
	}

	// Check fields property - it should be an array type now, not have a default
	fieldsProperty, exists := getTool.Inputs.Properties["fields"]
	if !exists {
		t.Error("'fields' property missing")
	} else {
		if fieldsProperty.Type != "array" {
			t.Errorf("Expected 'fields' to be array type, got %s", fieldsProperty.Type)
		}
	}

	// Test URL includes parameter placeholder
	providerConfig := getTool.ToolProvider

	url, ok := providerConfig["url"].(string)
	if !ok {
		t.Fatal("URL is not a string")
	}

	if !strings.Contains(url, "${issueKey}") {
		t.Error("URL should contain ${issueKey} placeholder")
	}
}

func TestJiraCreateIssueTool(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	// Find the create issue tool
	var createTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "jira_create_issue" {
			createTool = &tool
			break
		}
	}

	if createTool == nil {
		t.Fatal("jira_create_issue tool not found")
	}

	// Test required fields
	required := createTool.Inputs.Required
	if len(required) != 3 {
		t.Errorf("Expected 3 required fields, got %d", len(required))
	}

	expectedRequired := []string{"project", "summary", "issuetype"}
	for _, exp := range expectedRequired {
		found := false
		for _, req := range required {
			if req == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field '%s' not found", exp)
		}
	}

	// Check priority property - should not have a default anymore
	priorityProperty := createTool.Inputs.Properties["priority"]
	if priorityProperty.Type != "object" {
		t.Errorf("Expected 'priority' to be object type, got %s", priorityProperty.Type)
	}
	if priorityProperty.Default != nil {
		t.Errorf("Expected no default for priority, got %v", priorityProperty.Default)
	}

	// Test HTTP method
	providerConfig := createTool.ToolProvider

	if providerConfig["http_method"] != "POST" {
		t.Errorf("Expected http_method 'POST', got %v", providerConfig["http_method"])
	}
}

func TestAllToolsHaveValidProviders(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	for _, tool := range tools {
		// Check tool has required fields
		if tool.Name == "" {
			t.Error("Tool has empty name")
		}

		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Name)
		}

		if tool.Inputs.Type == "" {
			t.Errorf("Tool %s has empty inputs type", tool.Name)
		}

		if tool.Outputs.Type == "" {
			t.Errorf("Tool %s has empty outputs type", tool.Name)
		}

		// Check provider configuration
		if tool.ToolProvider == nil {
			t.Errorf("Tool %s has nil ToolProvider", tool.Name)
		}

		providerType, ok := tool.ToolProvider["provider_type"].(string)
		if !ok || providerType != "http" {
			t.Errorf("Tool %s has invalid provider_type", tool.Name)
		}

		url, ok := tool.ToolProvider["url"].(string)
		if !ok || url == "" {
			t.Errorf("Tool %s has invalid URL", tool.Name)
		}

		method, ok := tool.ToolProvider["http_method"].(string)
		if !ok || (method != "GET" && method != "POST" && method != "PUT") {
			t.Errorf("Tool %s has invalid HTTP method: %s", tool.Name, method)
		}

		// Check authentication exists
		auth, ok := tool.ToolProvider["auth"].(map[string]interface{})
		if !ok {
			t.Errorf("Tool %s has invalid auth configuration", tool.Name)
		}

		authType, ok := auth["auth_type"].(string)
		if !ok || authType != "basic" {
			t.Errorf("Tool %s has invalid auth_type", tool.Name)
		}
	}
}

func TestToolTags(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

	for _, tool := range tools {
		if len(tool.Tags) == 0 {
			t.Errorf("Tool %s has no tags", tool.Name)
		}

		// All tools should have "jira" tag
		hasJiraTag := false
		for _, tag := range tool.Tags {
			if tag == "jira" {
				hasJiraTag = true
				break
			}
		}

		if !hasJiraTag {
			t.Errorf("Tool %s missing 'jira' tag", tool.Name)
		}
	}
}
