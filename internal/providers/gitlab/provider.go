package gitlab

import (
	"fmt"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// Provider represents a GitLab provider
type Provider struct {
	BaseURL string
	Token   string
}

// NewProvider creates a new GitLab provider
func NewProvider(baseURL, token string) *Provider {
	return &Provider{
		BaseURL: baseURL,
		Token:   token,
	}
}

// GetTools returns all available GitLab tools
func (p *Provider) GetTools() []utcp.Tool {
	tools := []utcp.Tool{}

	// Search projects tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_search_projects",
		Description: "Search for GitLab projects by name or description",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"search": {
					Type:        "string",
					Description: "Search query for project name or description",
				},
				"visibility": {
					Type:        "string",
					Description: "Filter by visibility level",
					Enum:        []string{"public", "internal", "private"},
				},
				"owned": {
					Type:        "boolean",
					Description: "Limit to projects owned by current user",
					Default:     false,
				},
				"membership": {
					Type:        "boolean",
					Description: "Limit to projects where current user is a member",
					Default:     false,
				},
				"per_page": {
					Type:        "integer",
					Description: "Number of results per page (max 100)",
					Default:     20,
				},
				"page": {
					Type:        "integer",
					Description: "Page number for pagination",
					Default:     1,
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of projects matching the search criteria",
		},
		Tags: []string{"gitlab", "projects", "search"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_search_projects",
			fmt.Sprintf("%s/api/v4/projects", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Get project tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_get_project",
		Description: "Get detailed information about a GitLab project",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path (namespace/project)",
				},
				"statistics": {
					Type:        "boolean",
					Description: "Include project statistics",
					Default:     false,
				},
			},
			Required: []string{"id"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Project details including settings, permissions, and metadata",
		},
		Tags: []string{"gitlab", "project", "info"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_get_project",
			fmt.Sprintf("%s/api/v4/projects/${id}", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// List merge requests tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_list_merge_requests",
		Description: "List merge requests for a project",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"state": {
					Type:        "string",
					Description: "Filter by state",
					Enum:        []string{"opened", "closed", "locked", "merged", "all"},
					Default:     "opened",
				},
				"scope": {
					Type:        "string",
					Description: "Filter by scope",
					Enum:        []string{"created_by_me", "assigned_to_me", "all"},
					Default:     "all",
				},
				"author_id": {
					Type:        "integer",
					Description: "Filter by author user ID",
				},
				"assignee_id": {
					Type:        "integer",
					Description: "Filter by assignee user ID",
				},
				"labels": {
					Type:        "string",
					Description: "Comma-separated list of label names",
				},
				"milestone": {
					Type:        "string",
					Description: "Milestone title",
				},
				"per_page": {
					Type:        "integer",
					Description: "Results per page",
					Default:     20,
				},
			},
			Required: []string{"project_id"},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of merge requests with details",
		},
		Tags: []string{"gitlab", "merge_requests", "list"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_list_mrs",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/merge_requests", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Get merge request tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_get_merge_request",
		Description: "Get detailed information about a specific merge request",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"merge_request_iid": {
					Type:        "integer",
					Description: "Internal ID of the merge request",
				},
				"include_diverged_commits_count": {
					Type:        "boolean",
					Description: "Include diverged commits count",
					Default:     false,
				},
				"include_rebase_in_progress": {
					Type:        "boolean",
					Description: "Include rebase in progress flag",
					Default:     false,
				},
			},
			Required: []string{"project_id", "merge_request_iid"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Merge request details including diff stats, participants, and status",
		},
		Tags: []string{"gitlab", "merge_request", "details"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_get_mr",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/merge_requests/${merge_request_iid}", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// List issues tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_list_issues",
		Description: "List issues for a project or group",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path (optional, for project issues)",
				},
				"group_id": {
					Type:        "string",
					Description: "Group ID or URL-encoded path (optional, for group issues)",
				},
				"state": {
					Type:        "string",
					Description: "Filter by state",
					Enum:        []string{"opened", "closed", "all"},
					Default:     "opened",
				},
				"labels": {
					Type:        "string",
					Description: "Comma-separated list of label names",
				},
				"milestone": {
					Type:        "string",
					Description: "Milestone title",
				},
				"assignee_id": {
					Type:        "integer",
					Description: "Filter by assignee user ID",
				},
				"author_id": {
					Type:        "integer",
					Description: "Filter by author user ID",
				},
				"search": {
					Type:        "string",
					Description: "Search issues for text present in title or description",
				},
				"per_page": {
					Type:        "integer",
					Description: "Results per page",
					Default:     20,
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of issues with details",
		},
		Tags: []string{"gitlab", "issues", "list"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_list_issues",
			fmt.Sprintf("%s/api/v4/issues", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Get file contents tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_get_file",
		Description: "Get contents of a file from a GitLab repository",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"file_path": {
					Type:        "string",
					Description: "URL-encoded file path",
				},
				"ref": {
					Type:        "string",
					Description: "Branch, tag, or commit SHA",
					Default:     "main",
				},
			},
			Required: []string{"project_id", "file_path"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "File metadata and content (base64 encoded)",
		},
		Tags: []string{"gitlab", "repository", "file"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_get_file",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/repository/files/${file_path}", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// List repository tree tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_list_repository_tree",
		Description: "Get repository tree structure",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"path": {
					Type:        "string",
					Description: "Path inside repository (optional)",
				},
				"ref": {
					Type:        "string",
					Description: "Branch, tag, or commit SHA",
					Default:     "main",
				},
				"recursive": {
					Type:        "boolean",
					Description: "Get tree recursively",
					Default:     false,
				},
				"per_page": {
					Type:        "integer",
					Description: "Results per page",
					Default:     20,
				},
			},
			Required: []string{"project_id"},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of repository items (files and directories)",
		},
		Tags: []string{"gitlab", "repository", "tree"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_list_tree",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/repository/tree", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Get pipelines tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_list_pipelines",
		Description: "List CI/CD pipelines for a project",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"status": {
					Type:        "string",
					Description: "Filter by status",
					Enum:        []string{"created", "waiting_for_resource", "preparing", "pending", "running", "success", "failed", "canceled", "skipped", "manual", "scheduled"},
				},
				"ref": {
					Type:        "string",
					Description: "Filter by ref (branch or tag)",
				},
				"sha": {
					Type:        "string",
					Description: "Filter by commit SHA",
				},
				"username": {
					Type:        "string",
					Description: "Filter by username of pipeline triggerer",
				},
				"per_page": {
					Type:        "integer",
					Description: "Results per page",
					Default:     20,
				},
			},
			Required: []string{"project_id"},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "List of pipelines with status and metadata",
		},
		Tags: []string{"gitlab", "ci/cd", "pipelines"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_list_pipelines",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/pipelines", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Get pipeline details tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_get_pipeline",
		Description: "Get detailed information about a specific pipeline",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID or URL-encoded path",
				},
				"pipeline_id": {
					Type:        "integer",
					Description: "Pipeline ID",
				},
			},
			Required: []string{"project_id", "pipeline_id"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Pipeline details including jobs and status",
		},
		Tags: []string{"gitlab", "ci/cd", "pipeline"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_get_pipeline",
			fmt.Sprintf("%s/api/v4/projects/${project_id}/pipelines/${pipeline_id}", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	// Search code tool
	tools = append(tools, utcp.Tool{
		Name:        "gitlab_search_code",
		Description: "Search for code across all accessible projects",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"search": {
					Type:        "string",
					Description: "Search query",
				},
				"scope": {
					Type:        "string",
					Description: "Search scope",
					Default:     "blobs",
					Enum:        []string{"blobs", "commits"},
				},
				"per_page": {
					Type:        "integer",
					Description: "Results per page",
					Default:     20,
				},
			},
			Required: []string{"search"},
		},
		Outputs: utcp.Schema{
			Type:        "array",
			Description: "Search results with file paths and matching content",
		},
		Tags: []string{"gitlab", "search", "code"},
		ToolProvider: utcp.HTTPProvider(
			"gitlab_search_code",
			fmt.Sprintf("%s/api/v4/search", p.BaseURL),
			"GET",
			utcp.PersonalTokenAuth("GITLAB_TOKEN", "PRIVATE-TOKEN"),
		),
	})

	return tools
}
