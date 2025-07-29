package gitlab

import (
	"testing"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

func TestNewProvider(t *testing.T) {
	baseURL := "https://gitlab.example.com"
	token := "test-token-123"

	provider := NewProvider(baseURL, token)

	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	if provider.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, provider.BaseURL)
	}

	if provider.Token != token {
		t.Errorf("Expected Token %s, got %s", token, provider.Token)
	}
}

func TestGetTools(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	if len(tools) == 0 {
		t.Fatal("GetTools returned empty slice")
	}

	// Expected tools
	expectedTools := map[string]bool{
		"gitlab_search_projects":      false,
		"gitlab_get_project":          false,
		"gitlab_list_merge_requests":  false,
		"gitlab_get_merge_request":    false,
		"gitlab_list_issues":          false,
		"gitlab_get_file":             false,
		"gitlab_list_repository_tree": false,
		"gitlab_list_pipelines":       false,
		"gitlab_get_pipeline":         false,
		"gitlab_search_code":          false,
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

func TestGitLabSearchProjectsTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var searchTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_search_projects" {
			searchTool = &tool
			break
		}
	}

	if searchTool == nil {
		t.Fatal("gitlab_search_projects tool not found")
	}

	// Verify tool properties
	if searchTool.Description != "Search for GitLab projects by name or description" {
		t.Errorf("Unexpected description: %s", searchTool.Description)
	}

	// Check no required fields
	if len(searchTool.Inputs.Required) != 0 {
		t.Errorf("Expected no required fields, got %v", searchTool.Inputs.Required)
	}

	// Check properties
	props := searchTool.Inputs.Properties
	if _, exists := props["search"]; !exists {
		t.Error("Missing 'search' property")
	}

	// Check visibility enum
	visProp := props["visibility"]
	if len(visProp.Enum) != 3 {
		t.Errorf("Expected 3 visibility options, got %d", len(visProp.Enum))
	}

	// Check defaults
	if props["owned"].Default != false {
		t.Errorf("Expected default owned=false, got %v", props["owned"].Default)
	}

	if props["per_page"].Default != 20 {
		t.Errorf("Expected default per_page=20, got %v", props["per_page"].Default)
	}

	// Check provider configuration
	toolProvider := searchTool.ToolProvider
	if toolProvider["provider_type"] != "http" {
		t.Errorf("Expected provider_type 'http', got %v", toolProvider["provider_type"])
	}

	expectedURL := "https://gitlab.example.com/api/v4/projects"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}

	// Check authentication
	auth, ok := toolProvider["auth"].(map[string]interface{})
	if !ok {
		t.Fatal("Auth is not a map")
	}

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

func TestGitLabGetProjectTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var getTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_get_project" {
			getTool = &tool
			break
		}
	}

	if getTool == nil {
		t.Fatal("gitlab_get_project tool not found")
	}

	// Check required fields
	if len(getTool.Inputs.Required) != 1 || getTool.Inputs.Required[0] != "id" {
		t.Errorf("Expected 'id' as required field, got %v", getTool.Inputs.Required)
	}

	// Check URL has parameter placeholder
	toolProvider := getTool.ToolProvider
	expectedURL := "https://gitlab.example.com/api/v4/projects/${id}"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}
}

func TestGitLabListMergeRequestsTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var mrTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_list_merge_requests" {
			mrTool = &tool
			break
		}
	}

	if mrTool == nil {
		t.Fatal("gitlab_list_merge_requests tool not found")
	}

	// Check required fields
	if len(mrTool.Inputs.Required) != 1 || mrTool.Inputs.Required[0] != "project_id" {
		t.Errorf("Expected 'project_id' as required field, got %v", mrTool.Inputs.Required)
	}

	// Check state enum
	stateProp := mrTool.Inputs.Properties["state"]
	expectedStates := []string{"opened", "closed", "locked", "merged", "all"}
	if len(stateProp.Enum) != len(expectedStates) {
		t.Errorf("Expected %d state options, got %d", len(expectedStates), len(stateProp.Enum))
	}

	// Check default state
	if stateProp.Default != "opened" {
		t.Errorf("Expected default state 'opened', got %v", stateProp.Default)
	}
}

func TestGitLabListIssuesTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var issuesTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_list_issues" {
			issuesTool = &tool
			break
		}
	}

	if issuesTool == nil {
		t.Fatal("gitlab_list_issues tool not found")
	}

	// Check no required fields (project_id and group_id are optional)
	if len(issuesTool.Inputs.Required) != 0 {
		t.Errorf("Expected no required fields, got %v", issuesTool.Inputs.Required)
	}

	// Check properties exist
	props := issuesTool.Inputs.Properties
	if _, exists := props["project_id"]; !exists {
		t.Error("Missing 'project_id' property")
	}

	if _, exists := props["group_id"]; !exists {
		t.Error("Missing 'group_id' property")
	}

	if _, exists := props["search"]; !exists {
		t.Error("Missing 'search' property")
	}
}

func TestGitLabGetFileTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var fileTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_get_file" {
			fileTool = &tool
			break
		}
	}

	if fileTool == nil {
		t.Fatal("gitlab_get_file tool not found")
	}

	// Check required fields
	expectedRequired := []string{"project_id", "file_path"}
	if len(fileTool.Inputs.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(fileTool.Inputs.Required))
	}

	// Check default ref
	refProp := fileTool.Inputs.Properties["ref"]
	if refProp.Default != "main" {
		t.Errorf("Expected default ref 'main', got %v", refProp.Default)
	}

	// Check URL pattern
	toolProvider := fileTool.ToolProvider
	expectedURL := "https://gitlab.example.com/api/v4/projects/${project_id}/repository/files/${file_path}"
	if toolProvider["url"] != expectedURL {
		t.Errorf("Expected URL %s, got %v", expectedURL, toolProvider["url"])
	}
}

func TestGitLabListPipelinesTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var pipelineTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_list_pipelines" {
			pipelineTool = &tool
			break
		}
	}

	if pipelineTool == nil {
		t.Fatal("gitlab_list_pipelines tool not found")
	}

	// Check status enum
	statusProp := pipelineTool.Inputs.Properties["status"]
	if len(statusProp.Enum) != 11 {
		t.Errorf("Expected 11 status options, got %d", len(statusProp.Enum))
	}

	// Check tags include ci/cd
	hasCITag := false
	for _, tag := range pipelineTool.Tags {
		if tag == "ci/cd" {
			hasCITag = true
			break
		}
	}
	if !hasCITag {
		t.Error("Missing 'ci/cd' tag")
	}
}

func TestGitLabSearchCodeTool(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	var searchTool *utcp.Tool
	for _, tool := range tools {
		if tool.Name == "gitlab_search_code" {
			searchTool = &tool
			break
		}
	}

	if searchTool == nil {
		t.Fatal("gitlab_search_code tool not found")
	}

	// Check required fields
	if len(searchTool.Inputs.Required) != 1 || searchTool.Inputs.Required[0] != "search" {
		t.Errorf("Expected 'search' as required field, got %v", searchTool.Inputs.Required)
	}

	// Check scope enum and default
	scopeProp := searchTool.Inputs.Properties["scope"]
	if scopeProp.Default != "blobs" {
		t.Errorf("Expected default scope 'blobs', got %v", scopeProp.Default)
	}

	if len(scopeProp.Enum) != 2 {
		t.Errorf("Expected 2 scope options, got %d", len(scopeProp.Enum))
	}
}

func TestAllToolsHaveValidProviders(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
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
		if !ok || method != "GET" {
			t.Errorf("Tool %s has invalid HTTP method: %s", tool.Name, method)
		}

		// Check authentication exists
		auth, ok := tool.ToolProvider["auth"].(map[string]interface{})
		if !ok {
			t.Errorf("Tool %s has invalid auth configuration", tool.Name)
		}

		authType, ok := auth["auth_type"].(string)
		if !ok || authType != "personal_token" {
			t.Errorf("Tool %s has invalid auth_type", tool.Name)
		}
	}
}

func TestToolTags(t *testing.T) {
	provider := NewProvider("https://gitlab.example.com", "test-token")
	tools := provider.GetTools()

	for _, tool := range tools {
		if len(tool.Tags) == 0 {
			t.Errorf("Tool %s has no tags", tool.Name)
		}

		// All tools should have "gitlab" tag
		hasGitLabTag := false
		for _, tag := range tool.Tags {
			if tag == "gitlab" {
				hasGitLabTag = true
				break
			}
		}

		if !hasGitLabTag {
			t.Errorf("Tool %s missing 'gitlab' tag", tool.Name)
		}
	}
}
