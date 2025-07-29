package jira

import (
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

	// Verify tool properties
	if searchTool.Description != "Search Jira issues using JQL (Jira Query Language)" {
		t.Errorf("Unexpected description: %s", searchTool.Description)
	}

	// Check inputs
	if searchTool.Inputs.Type != "object" {
		t.Errorf("Expected inputs type 'object', got %s", searchTool.Inputs.Type)
	}

	// Check required fields
	if len(searchTool.Inputs.Required) != 1 || searchTool.Inputs.Required[0] != "jql" {
		t.Errorf("Expected 'jql' as required field, got %v", searchTool.Inputs.Required)
	}

	// Check properties
	if _, exists := searchTool.Inputs.Properties["jql"]; !exists {
		t.Error("Missing 'jql' property in inputs")
	}

	if _, exists := searchTool.Inputs.Properties["maxResults"]; !exists {
		t.Error("Missing 'maxResults' property in inputs")
	}

	// Check provider configuration
	toolProvider := searchTool.ToolProvider
	if toolProvider["provider_type"] != "http" {
		t.Errorf("Expected provider_type 'http', got %v", toolProvider["provider_type"])
	}

	expectedURL := "https://jira.example.com/rest/api/2/search"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}

	if toolProvider["http_method"] != "POST" {
		t.Errorf("Expected http_method 'POST', got %v", toolProvider["http_method"])
	}

	// Check authentication
	auth, ok := toolProvider["auth"].(map[string]interface{})
	if !ok {
		t.Fatal("Auth is not a map")
	}

	if auth["auth_type"] != "basic" {
		t.Errorf("Expected auth_type 'basic', got %v", auth["auth_type"])
	}

	if auth["username"] != "$JIRA_USERNAME" {
		t.Errorf("Expected username '$JIRA_USERNAME', got %v", auth["username"])
	}

	if auth["password"] != "$JIRA_PASSWORD" {
		t.Errorf("Expected password '$JIRA_PASSWORD', got %v", auth["password"])
	}
}

func TestJiraGetIssueTool(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

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

	// Check URL has parameter placeholder
	toolProvider := getTool.ToolProvider
	expectedURL := "https://jira.example.com/rest/api/2/issue/${issueKey}"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}

	// Check required fields
	if len(getTool.Inputs.Required) != 1 || getTool.Inputs.Required[0] != "issueKey" {
		t.Errorf("Expected 'issueKey' as required field, got %v", getTool.Inputs.Required)
	}

	// Check default values
	fieldsProperty := getTool.Inputs.Properties["fields"]
	if fieldsProperty.Default != "*all" {
		t.Errorf("Expected default fields '*all', got %v", fieldsProperty.Default)
	}
}

func TestJiraCreateIssueTool(t *testing.T) {
	provider := NewProvider("https://jira.example.com", "user", "pass")
	tools := provider.GetTools()

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

	// Check required fields
	expectedRequired := []string{"project", "summary", "issuetype"}
	if len(createTool.Inputs.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(createTool.Inputs.Required))
	}

	// Verify all required fields
	requiredMap := make(map[string]bool)
	for _, field := range createTool.Inputs.Required {
		requiredMap[field] = true
	}

	for _, expected := range expectedRequired {
		if !requiredMap[expected] {
			t.Errorf("Missing required field: %s", expected)
		}
	}

	// Check priority default
	priorityProperty := createTool.Inputs.Properties["priority"]
	if priorityProperty.Default != "Medium" {
		t.Errorf("Expected default priority 'Medium', got %v", priorityProperty.Default)
	}

	// Check labels property type
	labelsProperty := createTool.Inputs.Properties["labels"]
	if labelsProperty.Type != "array" {
		t.Errorf("Expected labels type 'array', got %s", labelsProperty.Type)
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
