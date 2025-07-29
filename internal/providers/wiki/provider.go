package wiki

import (
	"fmt"

	"github.com/rh-utcp/rh-utcp/pkg/utcp"
)

// Provider represents a Wiki/Confluence provider
type Provider struct {
	BaseURL string
	APIKey  string
}

// NewProvider creates a new Wiki provider
func NewProvider(baseURL, apiKey string) *Provider {
	return &Provider{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

// GetTools returns all available Wiki tools
func (p *Provider) GetTools() []utcp.Tool {
	tools := []utcp.Tool{}

	// Search pages tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_search_pages",
		Description: "Search for wiki pages by keyword or content",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"query": {
					Type:        "string",
					Description: "Search query (keywords, phrases, or CQL for advanced search)",
				},
				"space": {
					Type:        "string",
					Description: "Space key to limit search (optional)",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of results (default: 25)",
					Default:     25,
				},
				"start": {
					Type:        "integer",
					Description: "Starting index for pagination (default: 0)",
					Default:     0,
				},
			},
			Required: []string{"query"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Search results with pages and metadata",
		},
		Tags:                []string{"wiki", "search", "confluence"},
		AverageResponseSize: 500,
		ToolProvider: utcp.HTTPProvider(
			"wiki_search",
			fmt.Sprintf("%s/rest/api/content/search", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Get page tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_get_page",
		Description: "Get wiki page content by ID or title",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"pageId": {
					Type:        "string",
					Description: "Page ID (numeric string)",
				},
				"title": {
					Type:        "string",
					Description: "Page title (alternative to pageId)",
				},
				"spaceKey": {
					Type:        "string",
					Description: "Space key (required when using title)",
				},
				"expand": {
					Type:        "string",
					Description: "Comma-separated list of expansions (e.g., 'body.storage,version,ancestors')",
					Default:     "body.storage,version,space",
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Page content and metadata",
		},
		Tags:                []string{"wiki", "page", "content"},
		AverageResponseSize: 1000,
		ToolProvider: utcp.HTTPProvider(
			"wiki_get_page",
			fmt.Sprintf("%s/rest/api/content/${pageId}", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Create page tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_create_page",
		Description: "Create a new wiki page",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"title": {
					Type:        "string",
					Description: "Page title",
				},
				"spaceKey": {
					Type:        "string",
					Description: "Space key where the page will be created",
				},
				"content": {
					Type:        "string",
					Description: "Page content in storage format (HTML)",
				},
				"parentId": {
					Type:        "string",
					Description: "Parent page ID (optional)",
				},
			},
			Required: []string{"title", "spaceKey", "content"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Created page details including ID",
		},
		Tags: []string{"wiki", "create", "page"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_create_page",
			fmt.Sprintf("%s/rest/api/content", p.BaseURL),
			"POST",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Update page tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_update_page",
		Description: "Update an existing wiki page",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"pageId": {
					Type:        "string",
					Description: "Page ID to update",
				},
				"title": {
					Type:        "string",
					Description: "New page title",
				},
				"content": {
					Type:        "string",
					Description: "New page content in storage format (HTML)",
				},
				"version": {
					Type:        "integer",
					Description: "Current version number (for conflict detection)",
				},
				"message": {
					Type:        "string",
					Description: "Version message/comment",
				},
			},
			Required: []string{"pageId", "title", "content", "version"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Updated page details",
		},
		Tags: []string{"wiki", "update", "page"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_update_page",
			fmt.Sprintf("%s/rest/api/content/${pageId}", p.BaseURL),
			"PUT",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// List spaces tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_list_spaces",
		Description: "List all accessible wiki spaces",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"type": {
					Type:        "string",
					Description: "Space type filter (e.g., 'global', 'personal')",
					Enum:        []string{"global", "personal", "all"},
					Default:     "all",
				},
				"status": {
					Type:        "string",
					Description: "Space status filter",
					Enum:        []string{"current", "archived", "all"},
					Default:     "current",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of results",
					Default:     100,
				},
			},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "List of spaces with metadata",
		},
		Tags: []string{"wiki", "spaces", "list"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_list_spaces",
			fmt.Sprintf("%s/rest/api/space", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Get attachments tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_get_attachments",
		Description: "Get attachments for a wiki page",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"pageId": {
					Type:        "string",
					Description: "Page ID",
				},
				"filename": {
					Type:        "string",
					Description: "Filter by filename (optional)",
				},
				"mediaType": {
					Type:        "string",
					Description: "Filter by media type (optional)",
				},
			},
			Required: []string{"pageId"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "List of attachments with download links",
		},
		Tags: []string{"wiki", "attachments", "files"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_get_attachments",
			fmt.Sprintf("%s/rest/api/content/${pageId}/child/attachment", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Export page tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_export_page",
		Description: "Export wiki page in various formats",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"pageId": {
					Type:        "string",
					Description: "Page ID to export",
				},
				"format": {
					Type:        "string",
					Description: "Export format",
					Enum:        []string{"pdf", "word", "html", "xml"},
					Default:     "pdf",
				},
			},
			Required: []string{"pageId"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "Export URL or binary content",
		},
		Tags: []string{"wiki", "export", "download"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_export_page",
			fmt.Sprintf("%s/rest/api/content/${pageId}/export/${format}", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	// Get page history tool
	tools = append(tools, utcp.Tool{
		Name:        "wiki_get_page_history",
		Description: "Get version history of a wiki page",
		Inputs: utcp.Schema{
			Type: "object",
			Properties: map[string]utcp.Property{
				"pageId": {
					Type:        "string",
					Description: "Page ID",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of versions to return",
					Default:     20,
				},
			},
			Required: []string{"pageId"},
		},
		Outputs: utcp.Schema{
			Type:        "object",
			Description: "List of page versions with metadata",
		},
		Tags: []string{"wiki", "history", "versions"},
		ToolProvider: utcp.HTTPProvider(
			"wiki_get_history",
			fmt.Sprintf("%s/rest/api/content/${pageId}/version", p.BaseURL),
			"GET",
			utcp.APIKeyAuth("WIKI_API_KEY", "Authorization"),
		),
	})

	return tools
}
