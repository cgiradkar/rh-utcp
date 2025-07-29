# RH-UTCP

Universal Tool Calling Protocol (UTCP) implementation for exposing corporate tools (Jira, Wiki, GitLab) to AI agents.

## What is RH-UTCP?

RH-UTCP is a discovery server that allows AI agents to:
- Discover available corporate tools via a single `/utcp` endpoint
- Get standardized tool definitions with authentication details
- Make direct API calls to your existing enterprise systems

No middleware, no proxies - just tool discovery and direct API access.

## Key Features

- 🚀 **Direct API Access** - AI agents call your tools directly
- 🔐 **Flexible Authentication** - API keys, OAuth2, basic auth, tokens
- 🔧 **Easy Integration** - Point to existing APIs, no wrappers needed
- 📦 **Extensible** - Simple provider interface for new tools
- 🌐 **Protocol Agnostic** - HTTP, gRPC, GraphQL, WebSocket support

## Quick Start

```bash
# 1. Clone and setup
git clone <your-repo-url>
cd rh-utcp
make setup

# 2. Configure environment
cp env.example .env
# Edit .env with your credentials

# 3. Run the server
make run

# 4. Test discovery
curl http://localhost:8080/utcp
```

## Configuration

Set these environment variables in your `.env` file:

```bash
# Jira
JIRA_BASE_URL=https://jira.company.com
JIRA_USERNAME=your-username
JIRA_PASSWORD=your-password

# Wiki (Confluence)
WIKI_BASE_URL=https://wiki.company.com
WIKI_API_KEY=your-api-key

# GitLab
GITLAB_BASE_URL=https://gitlab.company.com
GITLAB_TOKEN=your-personal-token
```

## Available Tools

### Current
- **Jira**: Search issues, create/update tickets, manage projects
- **Wiki** (planned): Search pages, CRUD operations, attachments
- **GitLab** (planned): Projects, merge requests, code search

## Usage Example

```go
// AI agents discover tools
GET http://localhost:8080/utcp

// Response includes tool definitions
{
  "version": "1.0",
  "tools": [{
    "name": "jira_search_issues",
    "description": "Search Jira issues using JQL",
    "tool_provider": {
      "url": "https://jira.company.com/rest/api/2/search",
      "auth": {"auth_type": "basic", ...}
    }
  }]
}

// AI agents call tools directly
POST https://jira.company.com/rest/api/2/search
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make docker   # Build container
make help     # See all commands
```

## Documentation

- [Architecture](./docs/ARCHITECTURE.md) - System design and implementation details
- [UTCP Specification](https://github.com/universal-tool-calling-protocol/utcp-specification) - Protocol documentation

## License

Apache 2.0

## Acknowledgments

Built on the [Universal Tool Calling Protocol (UTCP)](https://github.com/universal-tool-calling-protocol/utcp-specification). 