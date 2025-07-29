package wiki

import (
	"testing"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

func TestNewProvider(t *testing.T) {
	baseURL := "https://wiki.example.com"
	apiKey := "test-api-key"

	provider := NewProvider(baseURL, apiKey)

	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	if provider.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, provider.BaseURL)
	}

	if provider.APIKey != apiKey {
		t.Errorf("Expected APIKey %s, got %s", apiKey, provider.APIKey)
	}
}

func TestGetTools(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	if len(tools) == 0 {
		t.Fatal("GetTools returned empty slice")
	}

	// Expected tools
	expectedTools := map[string]bool{
		"wiki_search_pages":     false,
		"wiki_get_page":         false,
		"wiki_create_page":      false,
		"wiki_update_page":      false,
		"wiki_list_spaces":      false,
		"wiki_get_attachments":  false,
		"wiki_export_page":      false,
		"wiki_get_page_history": false,
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

func TestWikiSearchPagesTool(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	var searchTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "wiki_search_pages" {
			searchTool = &tool
			break
		}
	}

	if searchTool == nil {
		t.Fatal("wiki_search_pages tool not found")
	}

	// Verify tool properties
	if searchTool.Description != "Search for wiki pages by keyword or content" {
		t.Errorf("Unexpected description: %s", searchTool.Description)
	}

	// Check required fields
	if len(searchTool.Inputs.Required) != 1 || searchTool.Inputs.Required[0] != "query" {
		t.Errorf("Expected 'query' as required field, got %v", searchTool.Inputs.Required)
	}

	// Check properties
	if _, exists := searchTool.Inputs.Properties["query"]; !exists {
		t.Error("Missing 'query' property in inputs")
	}

	if _, exists := searchTool.Inputs.Properties["space"]; !exists {
		t.Error("Missing 'space' property in inputs")
	}

	// Check defaults
	limitProp := searchTool.Inputs.Properties["limit"]
	if limitProp.Default != 25 {
		t.Errorf("Expected default limit 25, got %v", limitProp.Default)
	}

	// Check provider configuration
	toolProvider := searchTool.ToolProvider
	if toolProvider["provider_type"] != "http" {
		t.Errorf("Expected provider_type 'http', got %v", toolProvider["provider_type"])
	}

	expectedURL := "https://wiki.example.com/rest/api/content/search"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}

	if toolProvider["http_method"] != "GET" {
		t.Errorf("Expected http_method 'GET', got %v", toolProvider["http_method"])
	}

	// Check authentication
	auth, ok := toolProvider["auth"].(map[string]interface{})
	if !ok {
		t.Fatal("Auth is not a map")
	}

	if auth["auth_type"] != "api_key" {
		t.Errorf("Expected auth_type 'api_key', got %v", auth["auth_type"])
	}

	if auth["api_key"] != "$WIKI_API_KEY" {
		t.Errorf("Expected api_key '$WIKI_API_KEY', got %v", auth["api_key"])
	}

	if auth["var_name"] != "Authorization" {
		t.Errorf("Expected var_name 'Authorization', got %v", auth["var_name"])
	}
}

func TestWikiGetPageTool(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	var getTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "wiki_get_page" {
			getTool = &tool
			break
		}
	}

	if getTool == nil {
		t.Fatal("wiki_get_page tool not found")
	}

	// Check URL has parameter placeholder
	toolProvider := getTool.ToolProvider
	expectedURL := "https://wiki.example.com/rest/api/content/${pageId}"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}

	// Check no required fields (pageId or title+spaceKey)
	if len(getTool.Inputs.Required) != 0 {
		t.Errorf("Expected no required fields, got %v", getTool.Inputs.Required)
	}

	// Check expand default
	expandProp := getTool.Inputs.Properties["expand"]
	if expandProp.Default != "body.storage,version,space" {
		t.Errorf("Expected default expand, got %v", expandProp.Default)
	}

	// Check average response size
	if getTool.AverageResponseSize != 1000 {
		t.Errorf("Expected average response size 1000, got %d", getTool.AverageResponseSize)
	}
}

func TestWikiCreatePageTool(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	var createTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "wiki_create_page" {
			createTool = &tool
			break
		}
	}

	if createTool == nil {
		t.Fatal("wiki_create_page tool not found")
	}

	// Check required fields
	expectedRequired := []string{"title", "spaceKey", "content"}
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

	// Check HTTP method
	toolProvider := createTool.ToolProvider
	if toolProvider["http_method"] != "POST" {
		t.Errorf("Expected http_method 'POST', got %v", toolProvider["http_method"])
	}
}

func TestWikiListSpacesTool(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	var spacesTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "wiki_list_spaces" {
			spacesTool = &tool
			break
		}
	}

	if spacesTool == nil {
		t.Fatal("wiki_list_spaces tool not found")
	}

	// Check enum values for type property
	typeProp := spacesTool.Inputs.Properties["type"]
	if len(typeProp.Enum) != 3 {
		t.Errorf("Expected 3 enum values for type, got %d", len(typeProp.Enum))
	}

	expectedEnums := map[string]bool{"global": false, "personal": false, "all": false}
	for _, enum := range typeProp.Enum {
		if _, exists := expectedEnums[enum]; exists {
			expectedEnums[enum] = true
		}
	}

	for enum, found := range expectedEnums {
		if !found {
			t.Errorf("Missing enum value: %s", enum)
		}
	}

	// Check default value
	if typeProp.Default != "all" {
		t.Errorf("Expected default type 'all', got %v", typeProp.Default)
	}
}

func TestWikiExportPageTool(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	var exportTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "wiki_export_page" {
			exportTool = &tool
			break
		}
	}

	if exportTool == nil {
		t.Fatal("wiki_export_page tool not found")
	}

	// Check format enum values
	formatProp := exportTool.Inputs.Properties["format"]
	expectedFormats := []string{"pdf", "word", "html", "xml"}

	if len(formatProp.Enum) != len(expectedFormats) {
		t.Errorf("Expected %d format options, got %d", len(expectedFormats), len(formatProp.Enum))
	}

	// Check default format
	if formatProp.Default != "pdf" {
		t.Errorf("Expected default format 'pdf', got %v", formatProp.Default)
	}

	// Check URL pattern
	toolProvider := exportTool.ToolProvider
	expectedURL := "https://wiki.example.com/rest/api/content/${pageId}/export/${format}"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}
}

func TestAllToolsHaveValidProviders(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
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
		if !ok || authType != "api_key" {
			t.Errorf("Tool %s has invalid auth_type", tool.Name)
		}
	}
}

func TestToolTags(t *testing.T) {
	provider := NewProvider("https://wiki.example.com", "test-key")
	tools := provider.GetTools()

	for _, tool := range tools {
		if len(tool.Tags) == 0 {
			t.Errorf("Tool %s has no tags", tool.Name)
		}

		// All tools should have "wiki" tag
		hasWikiTag := false
		for _, tag := range tool.Tags {
			if tag == "wiki" {
				hasWikiTag = true
				break
			}
		}

		if !hasWikiTag {
			t.Errorf("Tool %s missing 'wiki' tag", tool.Name)
		}
	}
}
