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

	if manual.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", manual.Version)
	}

	if manual.Tools == nil {
		t.Error("Tools slice is nil")
	}

	if len(manual.Tools) != 0 {
		t.Errorf("Expected empty tools slice, got %d tools", len(manual.Tools))
	}
}

func TestAddTool(t *testing.T) {
	manual := NewManual()

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Inputs: Schema{
			Type: "object",
			Properties: map[string]Property{
				"param1": {
					Type:        "string",
					Description: "Test parameter",
				},
			},
			Required: []string{"param1"},
		},
		Outputs: Schema{
			Type:        "object",
			Description: "Test output",
		},
		Tags: []string{"test"},
		ToolProvider: map[string]interface{}{
			"provider_type": "http",
			"url":           "https://example.com/api",
		},
	}

	manual.AddTool(tool)

	if len(manual.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(manual.Tools))
	}

	if manual.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %s", manual.Tools[0].Name)
	}
}

func TestManualToJSON(t *testing.T) {
	manual := NewManual()
	manual.AddTool(Tool{
		Name:        "json_test",
		Description: "Test JSON serialization",
		Inputs: Schema{
			Type: "object",
		},
		Outputs: Schema{
			Type: "object",
		},
		ToolProvider: HTTPProvider("test", "https://example.com", "GET", nil),
	})

	jsonData, err := manual.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded Manual
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if decoded.Version != "1.0" {
		t.Errorf("Expected version 1.0 in JSON, got %s", decoded.Version)
	}

	if len(decoded.Tools) != 1 {
		t.Errorf("Expected 1 tool in JSON, got %d", len(decoded.Tools))
	}
}

func TestHTTPProvider(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		method   string
		auth     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:   "Basic HTTP provider",
			url:    "https://api.example.com",
			method: "GET",
			auth:   nil,
			expected: map[string]interface{}{
				"name":          "Basic HTTP provider",
				"provider_type": "http",
				"url":           "https://api.example.com",
				"http_method":   "GET",
				"content_type":  "application/json",
			},
		},
		{
			name:   "HTTP provider with auth",
			url:    "https://api.example.com",
			method: "POST",
			auth: map[string]interface{}{
				"auth_type": "api_key",
				"api_key":   "$API_KEY",
			},
			expected: map[string]interface{}{
				"name":          "HTTP provider with auth",
				"provider_type": "http",
				"url":           "https://api.example.com",
				"http_method":   "POST",
				"content_type":  "application/json",
				"auth": map[string]interface{}{
					"auth_type": "api_key",
					"api_key":   "$API_KEY",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := HTTPProvider(tt.name, tt.url, tt.method, tt.auth)

			if provider["name"] != tt.expected["name"] {
				t.Errorf("Expected name %s, got %s", tt.expected["name"], provider["name"])
			}

			if provider["provider_type"] != tt.expected["provider_type"] {
				t.Errorf("Expected provider_type %s, got %s", tt.expected["provider_type"], provider["provider_type"])
			}

			if provider["url"] != tt.expected["url"] {
				t.Errorf("Expected url %s, got %s", tt.expected["url"], provider["url"])
			}

			if provider["http_method"] != tt.expected["http_method"] {
				t.Errorf("Expected http_method %s, got %s", tt.expected["http_method"], provider["http_method"])
			}

			if tt.auth != nil && provider["auth"] == nil {
				t.Error("Expected auth to be set, but it's nil")
			}
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	auth := APIKeyAuth("MY_API_KEY", "X-Custom-Key")

	if auth["auth_type"] != "api_key" {
		t.Errorf("Expected auth_type 'api_key', got %s", auth["auth_type"])
	}

	if auth["api_key"] != "$MY_API_KEY" {
		t.Errorf("Expected api_key '$MY_API_KEY', got %s", auth["api_key"])
	}

	if auth["var_name"] != "X-Custom-Key" {
		t.Errorf("Expected var_name 'X-Custom-Key', got %s", auth["var_name"])
	}
}

func TestBasicAuth(t *testing.T) {
	auth := BasicAuth("USER_VAR", "PASS_VAR")

	if auth["auth_type"] != "basic" {
		t.Errorf("Expected auth_type 'basic', got %s", auth["auth_type"])
	}

	if auth["username"] != "$USER_VAR" {
		t.Errorf("Expected username '$USER_VAR', got %s", auth["username"])
	}

	if auth["password"] != "$PASS_VAR" {
		t.Errorf("Expected password '$PASS_VAR', got %s", auth["password"])
	}
}

func TestOAuth2Auth(t *testing.T) {
	auth := OAuth2Auth("CLIENT_ID", "CLIENT_SECRET", "https://auth.example.com/token")

	if auth["auth_type"] != "oauth2" {
		t.Errorf("Expected auth_type 'oauth2', got %s", auth["auth_type"])
	}

	if auth["client_id"] != "$CLIENT_ID" {
		t.Errorf("Expected client_id '$CLIENT_ID', got %s", auth["client_id"])
	}

	if auth["client_secret"] != "$CLIENT_SECRET" {
		t.Errorf("Expected client_secret '$CLIENT_SECRET', got %s", auth["client_secret"])
	}

	if auth["token_url"] != "https://auth.example.com/token" {
		t.Errorf("Expected token_url 'https://auth.example.com/token', got %s", auth["token_url"])
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
