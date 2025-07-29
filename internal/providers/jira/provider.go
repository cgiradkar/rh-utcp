package jira

import (
	"fmt"

	"github.com/rh-utcp/rh-utcp/internal/providers"
	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// Provider represents a Jira provider
type Provider struct {
	providers.BaseProvider
	Username string
	Password string
}

// NewProvider creates a new Jira provider
func NewProvider(baseURL, username, password string) *Provider {
	return &Provider{
		BaseProvider: providers.BaseProvider{
			Type:    "jira",
			Enabled: true,
			BaseURL: baseURL,
		},
		Username: username,
		Password: password,
	}
}

// NewProviderFromConfig creates a new Jira provider from configuration
func NewProviderFromConfig(config map[string]interface{}) (providers.Provider, error) {
	name, _ := config["name"].(string)
	baseURL, _ := config["base_url"].(string)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	enabled, _ := config["enabled"].(bool)

	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required for Jira provider")
	}

	provider := NewProvider(baseURL, username, password)
	provider.Name = name
	provider.Enabled = enabled

	return provider, nil
}

// GetTools returns all available Jira tools
func (p *Provider) GetTools() []utcp.Tool {
	tools := []utcp.Tool{}

	// Search issues tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_search_issues",
		Description: "Search for Jira issues using JQL (Jira Query Language)",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"jql": {
					Type:        "string",
					Description: "JQL query string (e.g., 'project = PROJ AND status = Open')",
				},
				"fields": {
					Type:        "array",
					Description: "Fields to return (e.g., ['summary', 'status', 'assignee'])",
				},
				"maxResults": {
					Type:        "integer",
					Description: "Maximum number of results to return",
					Default:     50,
				},
				"startAt": {
					Type:        "integer",
					Description: "Starting index for pagination",
					Default:     0,
				},
			},
			Required: []string{"jql"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Search results containing issues and metadata",
		},
		Tags: []string{"jira", "search", "issues"},
		ToolProvider: utcp.HTTPProvider(
			"jira_search",
			fmt.Sprintf("%s/rest/api/2/search", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Get issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_issue",
		Description: "Get detailed information about a specific Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "Issue key (e.g., 'PROJ-123')",
				},
				"fields": {
					Type:        "array",
					Description: "Specific fields to return",
				},
				"expand": {
					Type:        "array",
					Description: "Additional data to expand (e.g., ['changelog', 'renderedFields'])",
				},
			},
			Required: []string{"issueKey"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Complete issue details",
		},
		Tags: []string{"jira", "issue", "get"},
		ToolProvider: utcp.HTTPProvider(
			"jira_get_issue",
			fmt.Sprintf("%s/rest/api/2/issue/${issueKey}", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Create issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_create_issue",
		Description: "Create a new Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project": {
					Type:        "object",
					Description: "Project key or ID",
				},
				"summary": {
					Type:        "string",
					Description: "Issue summary/title",
				},
				"description": {
					Type:        "string",
					Description: "Issue description",
				},
				"issuetype": {
					Type:        "object",
					Description: "Issue type (e.g., {'name': 'Bug'})",
				},
				"priority": {
					Type:        "object",
					Description: "Priority (e.g., {'name': 'High'})",
				},
				"assignee": {
					Type:        "object",
					Description: "Assignee account ID or name",
				},
				"labels": {
					Type:        "array",
					Description: "Labels to add to the issue",
				},
			},
			Required: []string{"project", "summary", "issuetype"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Created issue details including key and ID",
		},
		Tags: []string{"jira", "issue", "create"},
		ToolProvider: utcp.HTTPProvider(
			"jira_create_issue",
			fmt.Sprintf("%s/rest/api/2/issue", p.BaseURL),
			"POST",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Update issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_update_issue",
		Description: "Update an existing Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "Issue key to update",
				},
				"fields": {
					Type:        "object",
					Description: "Fields to update",
				},
				"update": {
					Type:        "object",
					Description: "Update operations (add, set, remove)",
				},
			},
			Required: []string{"issueKey"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Update confirmation",
		},
		Tags: []string{"jira", "issue", "update"},
		ToolProvider: utcp.HTTPProvider(
			"jira_update_issue",
			fmt.Sprintf("%s/rest/api/2/issue/${issueKey}", p.BaseURL),
			"PUT",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Get projects tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_projects",
		Description: "Get list of all Jira projects",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"expand": {
					Type:        "array",
					Description: "Additional project data to retrieve",
				},
				"recent": {
					Type:        "integer",
					Description: "Return only recent projects (number)",
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of projects with details",
		},
		Tags: []string{"jira", "projects", "list"},
		ToolProvider: utcp.HTTPProvider(
			"jira_get_projects",
			fmt.Sprintf("%s/rest/api/2/project", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Add comment tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_add_comment",
		Description: "Add a comment to a Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "Issue key to comment on",
				},
				"body": {
					Type:        "string",
					Description: "Comment text (supports Jira wiki markup)",
				},
				"visibility": {
					Type:        "object",
					Description: "Comment visibility restrictions",
				},
			},
			Required: []string{"issueKey", "body"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Created comment details",
		},
		Tags: []string{"jira", "comment", "add"},
		ToolProvider: utcp.HTTPProvider(
			"jira_add_comment",
			fmt.Sprintf("%s/rest/api/2/issue/${issueKey}/comment", p.BaseURL),
			"POST",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Get user issues tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_user_issues",
		Description: "Get issues assigned to or reported by a specific user",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"username": {
					Type:        "string",
					Description: "Username or account ID (use 'currentUser()' for current user)",
				},
				"filter": {
					Type:        "string",
					Description: "Filter type: 'assignee', 'reporter', or 'both'",
					Default:     "assignee",
				},
				"status": {
					Type:        "array",
					Description: "Status filters (e.g., ['Open', 'In Progress'])",
				},
				"maxResults": {
					Type:        "integer",
					Description: "Maximum results to return",
					Default:     50,
				},
			},
			Required: []string{"username"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Issues assigned to or reported by the user",
		},
		Tags: []string{"jira", "user", "issues"},
		ToolProvider: utcp.HTTPProvider(
			"jira_user_issues",
			fmt.Sprintf("%s/rest/api/2/search", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	return tools
}
