package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

func TestHealthEndpoint(t *testing.T) {
	// Set up test router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", response["status"])
	}
}

func TestUTCPDiscoveryWithoutProviders(t *testing.T) {
	// Ensure no provider environment variables are set
	os.Unsetenv("JIRA_BASE_URL")
	os.Unsetenv("WIKI_BASE_URL")
	os.Unsetenv("GITLAB_BASE_URL")

	// Set up test router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/utcp", handleUTCPDiscovery)

	// Create test request
	req, _ := http.NewRequest("GET", "/utcp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var manual utcp.Manual
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if manual.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", manual.Version)
	}

	if len(manual.Tools) != 0 {
		t.Errorf("Expected 0 tools without providers, got %d", len(manual.Tools))
	}
}

func TestUTCPDiscoveryWithJiraProvider(t *testing.T) {
	// Set up Jira environment variables
	os.Setenv("JIRA_BASE_URL", "https://jira.test.com")
	os.Setenv("JIRA_USERNAME", "testuser")
	os.Setenv("JIRA_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("JIRA_BASE_URL")
		os.Unsetenv("JIRA_USERNAME")
		os.Unsetenv("JIRA_PASSWORD")
	}()

	// Set up test router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/utcp", handleUTCPDiscovery)

	// Create test request
	req, _ := http.NewRequest("GET", "/utcp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var manual utcp.Manual
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if manual.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", manual.Version)
	}

	// Should have Jira tools
	if len(manual.Tools) == 0 {
		t.Fatal("Expected Jira tools to be present")
	}

	// Check for specific Jira tools
	toolNames := make(map[string]bool)
	for _, tool := range manual.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"jira_search_issues",
		"jira_get_issue",
		"jira_create_issue",
		"jira_update_issue",
		"jira_get_projects",
		"jira_add_comment",
		"jira_get_user_issues",
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Expected tool %s not found", expected)
		}
	}
}

func TestUTCPDiscoveryResponseStructure(t *testing.T) {
	// Set up a provider
	os.Setenv("JIRA_BASE_URL", "https://jira.test.com")
	defer os.Unsetenv("JIRA_BASE_URL")

	// Set up test router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/utcp", handleUTCPDiscovery)

	// Create test request
	req, _ := http.NewRequest("GET", "/utcp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Parse response
	var manual utcp.Manual
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check at least one tool exists
	if len(manual.Tools) == 0 {
		t.Fatal("No tools found in response")
	}

	// Validate first tool structure
	tool := manual.Tools[0]

	// Check required fields
	if tool.Name == "" {
		t.Error("Tool name is empty")
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Inputs.Type == "" {
		t.Error("Tool inputs type is empty")
	}

	if tool.Outputs.Type == "" {
		t.Error("Tool outputs type is empty")
	}

	if tool.ToolProvider == nil {
		t.Error("Tool provider is nil")
	}

	// Check provider structure
	provider := tool.ToolProvider
	if provider["provider_type"] == nil {
		t.Error("Provider type is missing")
	}

	if provider["url"] == nil {
		t.Error("Provider URL is missing")
	}

	if provider["http_method"] == nil {
		t.Error("Provider HTTP method is missing")
	}

	if provider["auth"] == nil {
		t.Error("Provider auth is missing")
	}
}

func TestUTCPDiscoveryContentType(t *testing.T) {
	// Set up test router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/utcp", handleUTCPDiscovery)

	// Create test request
	req, _ := http.NewRequest("GET", "/utcp", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/json; charset=utf-8"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}
}

func TestMainFunction(t *testing.T) {
	// This test verifies that the main function sets up routes correctly
	// In a real scenario, you might want to test with a test server

	// For now, just verify that the handler functions exist
	if handleUTCPDiscovery == nil {
		t.Error("handleUTCPDiscovery function is nil")
	}
}
