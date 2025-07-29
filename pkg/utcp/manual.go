package utcp

import (
	"encoding/json"
)

// Manual represents a UTCP manual with version and tools
type Manual struct {
	Version string `json:"version"`
	Tools   []Tool `json:"tools"`
}

// Tool represents a single tool in the UTCP manual
type Tool struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Inputs              Schema                 `json:"inputs"`
	Outputs             Schema                 `json:"outputs"`
	Tags                []string               `json:"tags,omitempty"`
	AverageResponseSize int                    `json:"average_response_size,omitempty"`
	ToolProvider        map[string]interface{} `json:"tool_provider"`
}

// Schema represents input/output schema for a tool
type Schema struct {
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
	Description string              `json:"description,omitempty"`
	Title       string              `json:"title,omitempty"`
}

// Property represents a single property in a schema
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Enum        []string    `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// NewManual creates a new UTCP manual
func NewManual() *Manual {
	return &Manual{
		Version: "0.1.0",
		Tools:   []Tool{},
	}
}

// AddTool adds a tool to the manual
func (m *Manual) AddTool(tool Tool) {
	m.Tools = append(m.Tools, tool)
}

// ToJSON converts the manual to JSON
func (m *Manual) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HTTPProvider creates an HTTP provider configuration
func HTTPProvider(name, url, method string, auth map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"provider_type": "http",
		"provider_id":   name,
		"url":           url,
		"http_method":   method,
		"auth":          auth,
	}
}

// APIKeyAuth creates API key authentication configuration
func APIKeyAuth(envVar, varName string) map[string]interface{} {
	return map[string]interface{}{
		"auth_type": "api_key",
		"api_key":   "$" + envVar,
		"var_name":  varName,
	}
}

// BasicAuth creates basic authentication configuration
func BasicAuth(usernameEnv, passwordEnv string) map[string]interface{} {
	return map[string]interface{}{
		"auth_type": "basic",
		"username":  "$" + usernameEnv,
		"password":  "$" + passwordEnv,
	}
}

// OAuth2Auth creates OAuth2 authentication configuration
func OAuth2Auth(clientIDEnv, clientSecretEnv, tokenURLEnv string) map[string]interface{} {
	return map[string]interface{}{
		"auth_type":     "oauth2",
		"client_id":     "$" + clientIDEnv,
		"client_secret": "$" + clientSecretEnv,
		"token_url":     "$" + tokenURLEnv,
	}
}

// PersonalTokenAuth creates personal token authentication configuration
func PersonalTokenAuth(tokenEnv, headerName string) map[string]interface{} {
	return map[string]interface{}{
		"auth_type":   "personal_token",
		"token":       "$" + tokenEnv,
		"header_name": headerName,
	}
}
