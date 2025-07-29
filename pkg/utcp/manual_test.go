package utcp

import (
	"encoding/json"
	"testing"
)

func TestNewManual(t *testing.T) {
	manual := NewManual()

	if manual == nil {
		t.Fatal("NewManual returned nil")
	}

	if manual.Version != "0.1.0" {
		t.Errorf("Expected version 0.1.0, got %s", manual.Version)
	}

	if len(manual.Tools) != 0 {
		t.Errorf("Expected empty tools slice, got %d tools", len(manual.Tools))
	}
}

func TestAddTool(t *testing.T) {
	manual := NewManual()

	tool := Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Inputs:      Schema{Type: "object"},
		Outputs:     Schema{Type: "object"},
	}

	manual.AddTool(tool)

	if len(manual.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(manual.Tools))
	}

	if manual.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %s", manual.Tools[0].Name)
	}
}

func TestToJSON(t *testing.T) {
	manual := NewManual()
	manual.AddTool(Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Inputs:      Schema{Type: "object"},
		Outputs:     Schema{Type: "object"},
	})

	jsonStr, err := manual.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse JSON to verify it's valid
	var parsed Manual
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.Version != manual.Version {
		t.Errorf("Version mismatch: expected %s, got %s", manual.Version, parsed.Version)
	}

	if len(parsed.Tools) != len(manual.Tools) {
		t.Errorf("Tools count mismatch: expected %d, got %d", len(manual.Tools), len(parsed.Tools))
	}
}

func TestHTTPProvider(t *testing.T) {
	auth := map[string]interface{}{
		"auth_type": "basic",
		"username":  "$USER",
		"password":  "$PASS",
	}

	provider := HTTPProvider("test_provider", "https://api.example.com", "POST", auth)

	if provider["provider_type"] != "http" {
		t.Errorf("Expected provider_type 'http', got %v", provider["provider_type"])
	}

	if provider["provider_id"] != "test_provider" {
		t.Errorf("Expected provider_id 'test_provider', got %v", provider["provider_id"])
	}

	if provider["url"] != "https://api.example.com" {
		t.Errorf("Expected url 'https://api.example.com', got %v", provider["url"])
	}

	if provider["http_method"] != "POST" {
		t.Errorf("Expected http_method 'POST', got %v", provider["http_method"])
	}

	authMap, ok := provider["auth"].(map[string]interface{})
	if !ok {
		t.Fatal("Auth is not a map")
	}

	if authMap["auth_type"] != "basic" {
		t.Errorf("Expected auth_type 'basic', got %v", authMap["auth_type"])
	}
}

func TestAPIKeyAuth(t *testing.T) {
	auth := APIKeyAuth("API_KEY", "X-API-Key")

	if auth["auth_type"] != "api_key" {
		t.Errorf("Expected auth_type 'api_key', got %v", auth["auth_type"])
	}

	if auth["api_key"] != "$API_KEY" {
		t.Errorf("Expected api_key '$API_KEY', got %v", auth["api_key"])
	}

	if auth["var_name"] != "X-API-Key" {
		t.Errorf("Expected var_name 'X-API-Key', got %v", auth["var_name"])
	}
}

func TestBasicAuth(t *testing.T) {
	auth := BasicAuth("USERNAME", "PASSWORD")

	if auth["auth_type"] != "basic" {
		t.Errorf("Expected auth_type 'basic', got %v", auth["auth_type"])
	}

	if auth["username"] != "$USERNAME" {
		t.Errorf("Expected username '$USERNAME', got %v", auth["username"])
	}

	if auth["password"] != "$PASSWORD" {
		t.Errorf("Expected password '$PASSWORD', got %v", auth["password"])
	}
}

func TestOAuth2Auth(t *testing.T) {
	auth := OAuth2Auth("CLIENT_ID", "CLIENT_SECRET", "TOKEN_URL")

	if auth["auth_type"] != "oauth2" {
		t.Errorf("Expected auth_type 'oauth2', got %v", auth["auth_type"])
	}

	if auth["client_id"] != "$CLIENT_ID" {
		t.Errorf("Expected client_id '$CLIENT_ID', got %v", auth["client_id"])
	}

	if auth["client_secret"] != "$CLIENT_SECRET" {
		t.Errorf("Expected client_secret '$CLIENT_SECRET', got %v", auth["client_secret"])
	}

	if auth["token_url"] != "$TOKEN_URL" {
		t.Errorf("Expected token_url '$TOKEN_URL', got %v", auth["token_url"])
	}
}

func TestPersonalTokenAuth(t *testing.T) {
	auth := PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN")

	if auth["auth_type"] != "personal_token" {
		t.Errorf("Expected auth_type 'personal_token', got %v", auth["auth_type"])
	}

	if auth["token"] != "$GITLAB_TOKEN" {
		t.Errorf("Expected token '$GITLAB_TOKEN', got %v", auth["token"])
	}

	if auth["header_name"] != "PRIVATE-TOKEN" {
		t.Errorf("Expected header_name 'PRIVATE-TOKEN', got %v", auth["header_name"])
	}
}

func TestSchemaValidation(t *testing.T) {
	schema := Schema{
		Type: "object",
		Properties: map[string]Property{
			"name": {
				Type:        "string",
				Description: "User name",
			},
			"age": {
				Type:        "integer",
				Description: "User age",
			},
			"status": {
				Type:        "string",
				Description: "User status",
				Enum:        []string{"active", "inactive"},
				Default:     "active",
			},
		},
		Required:    []string{"name"},
		Description: "User object",
		Title:       "User",
	}

	// Test that all fields are set correctly
	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if len(schema.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(schema.Properties))
	}

	if schema.Properties["status"].Default != "active" {
		t.Errorf("Expected default 'active', got %v", schema.Properties["status"].Default)
	}

	if len(schema.Properties["status"].Enum) != 2 {
		t.Errorf("Expected 2 enum values, got %d", len(schema.Properties["status"].Enum))
	}
}
