package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rh-utcp/rh-utcp/internal/config"
	"github.com/rh-utcp/rh-utcp/internal/providers"
	"github.com/rh-utcp/rh-utcp/internal/providers/jira"
	"github.com/rh-utcp/rh-utcp/pkg/logger"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	// Initialize test dependencies
	if log == nil {
		log = logger.New(logger.Config{
			Level:    "error",
			UseColor: false,
		})
	}

	if registry == nil {
		registry = providers.NewRegistry()
	}

	if cfg == nil {
		cfg = &config.Config{
			Server: config.ServerConfig{
				Port:        "8080",
				Environment: "test",
				LogLevel:    "error",
			},
			Providers: []config.ProviderConfig{},
		}
	}

	r := gin.New()
	r.GET("/utcp", handleUTCPDiscovery)
	r.GET("/health", handleHealth)

	return r
}

func TestHealthEndpoint(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}
}

func TestUTCPDiscoveryWithoutProviders(t *testing.T) {
	r := setupTestRouter()

	// Clear any existing providers
	registry.Clear()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/utcp", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var manual map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check structure
	if _, exists := manual["version"]; !exists {
		t.Error("Missing 'version' field in manual")
	}

	tools, ok := manual["tools"].([]interface{})
	if !ok {
		t.Error("'tools' field is not an array")
	}

	if len(tools) != 0 {
		t.Errorf("Expected 0 tools without providers, got %d", len(tools))
	}
}

func TestUTCPDiscoveryWithJiraProvider(t *testing.T) {
	r := setupTestRouter()

	// Clear and add a Jira provider
	registry.Clear()
	registry.RegisterFactory("jira", jira.NewProviderFromConfig)

	err := registry.CreateProvider("test-jira", "jira", map[string]interface{}{
		"name":     "test-jira",
		"enabled":  true,
		"base_url": "https://jira.example.com",
		"username": "testuser",
		"password": "testpass",
	})

	if err != nil {
		t.Fatalf("Failed to create Jira provider: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/utcp", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var manual map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	tools, ok := manual["tools"].([]interface{})
	if !ok {
		t.Error("'tools' field is not an array")
	}

	// Jira provider should provide 7 tools
	if len(tools) != 7 {
		t.Errorf("Expected 7 tools from Jira provider, got %d", len(tools))
	}

	// Check first tool structure
	if len(tools) > 0 {
		firstTool, ok := tools[0].(map[string]interface{})
		if !ok {
			t.Error("Tool is not a map")
		} else {
			// Check required fields
			requiredFields := []string{"name", "description", "inputs", "outputs", "tags", "tool_provider"}
			for _, field := range requiredFields {
				if _, exists := firstTool[field]; !exists {
					t.Errorf("Missing required field '%s' in tool", field)
				}
			}
		}
	}
}

func TestUTCPDiscoveryResponseStructure(t *testing.T) {
	r := setupTestRouter()

	// Clear providers
	registry.Clear()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/utcp", nil)
	r.ServeHTTP(w, req)

	var manual map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &manual); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check version
	version, ok := manual["version"].(string)
	if !ok {
		t.Error("'version' is not a string")
	} else if version == "" {
		t.Error("'version' is empty")
	}

	// Check tools array
	tools, ok := manual["tools"].([]interface{})
	if !ok {
		t.Error("'tools' is not an array")
	} else if tools == nil {
		t.Error("'tools' is nil")
	}

	// Check no extra fields
	expectedFields := map[string]bool{
		"version": true,
		"tools":   true,
	}

	for key := range manual {
		if !expectedFields[key] {
			t.Errorf("Unexpected field '%s' in manual", key)
		}
	}
}

func TestUTCPDiscoveryContentType(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/utcp", nil)
	r.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/json; charset=utf-8"

	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}
}

// TestMain validates that the main function can be called without errors
func TestMain(t *testing.T) {
	// This test simply ensures the code compiles and basic structure is valid
	// The actual main() function would start a server, so we don't call it in tests

	// Instead, we test that our handler functions exist
	if ginLogger() == nil {
		t.Error("ginLogger function should not return nil")
	}
}
