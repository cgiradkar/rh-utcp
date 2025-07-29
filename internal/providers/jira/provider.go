package jira

import (
	"fmt"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// Provider represents a Jira provider
type Provider struct {
	BaseURL  string
	Username string
	Password string
}

// NewProvider creates a new Jira provider
func NewProvider(baseURL, username, password string) *Provider {
	return &Provider{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
	}
}

// GetTools returns all available Jira tools
func (p *Provider) GetTools() []utcp.Tool {
	tools := []utcp.Tool{}

	// Search Issues tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_search_issues",
		Description: "Search Jira issues using JQL (Jira Query Language)",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"jql": {
					Type:        "string",
					Description: "JQL query string (e.g., 'project = PROJ AND status = Open')",
				},
				"fields": {
					Type:        "string",
					Description: "Comma-separated fields to return (default: key,summary,status,assignee)",
					Default:     "key,summary,status,assignee,priority,created,updated",
				},
				"maxResults": {
					Type:        "integer",
					Description: "Maximum number of results (default: 50)",
					Default:     50,
				},
				"startAt": {
					Type:        "integer",
					Description: "Starting index for pagination (default: 0)",
					Default:     0,
				},
			},
			Required: []string{"jql"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Search results with issues and pagination info",
		},
		Tags: []string{"jira", "search", "issues"},
		ToolProvider: utcp.HTTPProvider(
			"jira_search",
			fmt.Sprintf("%s/rest/api/2/search", p.BaseURL),
			"POST",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Get Issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_issue",
		Description: "Get detailed information about a specific Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123')",
				},
				"fields": {
					Type:        "string",
					Description: "Comma-separated fields to return (default: all fields)",
					Default:     "*all",
				},
				"expand": {
					Type:        "string",
					Description: "Comma-separated list of fields to expand",
					Default:     "renderedFields,names,schema,transitions,operations,editmeta,changelog",
				},
			},
			Required: []string{"issueKey"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Complete issue details",
		},
		Tags: []string{"jira", "issue", "details"},
		ToolProvider: utcp.HTTPProvider(
			"jira_get_issue",
			fmt.Sprintf("%s/rest/api/2/issue/${issueKey}", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Create Issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_create_issue",
		Description: "Create a new Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project": {
					Type:        "string",
					Description: "Project key where the issue will be created",
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
					Type:        "string",
					Description: "Issue type (e.g., 'Bug', 'Task', 'Story')",
				},
				"priority": {
					Type:        "string",
					Description: "Priority (e.g., 'High', 'Medium', 'Low')",
					Default:     "Medium",
				},
				"assignee": {
					Type:        "string",
					Description: "Username of the assignee",
				},
				"labels": {
					Type:        "array",
					Description: "Array of labels to add to the issue",
				},
			},
			Required: []string{"project", "summary", "issuetype"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Created issue details including key",
		},
		Tags: []string{"jira", "create", "issue"},
		ToolProvider: utcp.HTTPProvider(
			"jira_create_issue",
			fmt.Sprintf("%s/rest/api/2/issue", p.BaseURL),
			"POST",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Update Issue tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_update_issue",
		Description: "Update an existing Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "The issue key to update (e.g., 'PROJ-123')",
				},
				"fields": {
					Type:        "object",
					Description: "Fields to update (e.g., {\"summary\": \"New summary\", \"description\": \"New description\"})",
				},
			},
			Required: []string{"issueKey", "fields"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Update confirmation",
		},
		Tags: []string{"jira", "update", "issue"},
		ToolProvider: utcp.HTTPProvider(
			"jira_update_issue",
			fmt.Sprintf("%s/rest/api/2/issue/${issueKey}", p.BaseURL),
			"PUT",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Get Projects tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_projects",
		Description: "Get list of all Jira projects",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"expand": {
					Type:        "string",
					Description: "Comma-separated list of fields to expand",
					Default:     "description,lead,url,projectKeys",
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of projects",
		},
		Tags: []string{"jira", "projects", "list"},
		ToolProvider: utcp.HTTPProvider(
			"jira_get_projects",
			fmt.Sprintf("%s/rest/api/2/project", p.BaseURL),
			"GET",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	// Add Comment tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_add_comment",
		Description: "Add a comment to a Jira issue",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"issueKey": {
					Type:        "string",
					Description: "The issue key (e.g., 'PROJ-123')",
				},
				"body": {
					Type:        "string",
					Description: "Comment text",
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

	// Get User Issues tool
	tools = append(tools, utcp.Tool{
		Name:        "jira_get_user_issues",
		Description: "Get issues assigned to a specific user",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"username": {
					Type:        "string",
					Description: "Username to search for (use 'currentUser()' for current user)",
					Default:     "currentUser()",
				},
				"status": {
					Type:        "string",
					Description: "Filter by status (e.g., 'Open', 'In Progress', 'Done')",
				},
				"maxResults": {
					Type:        "integer",
					Description: "Maximum number of results",
					Default:     50,
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Search results with user's issues",
		},
		Tags: []string{"jira", "user", "issues"},
		ToolProvider: utcp.HTTPProvider(
			"jira_user_issues",
			fmt.Sprintf("%s/rest/api/2/search", p.BaseURL),
			"POST",
			utcp.BasicAuth("JIRA_USERNAME", "JIRA_PASSWORD"),
		),
	})

	return tools
}
