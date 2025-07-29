# Integrating RH-UTCP with Cursor Editor

## Prerequisites
1. RH-UTCP server running on `http://localhost:8080`
2. Cursor editor with Composer feature enabled

## Setup Instructions

### Method 1: Using Cursor's Tool Configuration

1. **Open Cursor Settings**
   - Press `Cmd/Ctrl + ,` to open settings
   - Search for "tools" or "composer tools"

2. **Add UTCP Discovery Endpoint**
   - Add your UTCP server URL: `http://localhost:8080/utcp`
   - Cursor will automatically discover all 25 tools

3. **Test the Integration**
   - Open Composer (`Cmd/Ctrl + K`)
   - Type commands like:
     - "Search for open Jira issues in project XYZ"
     - "Get the latest merge requests from GitLab"
     - "Find wiki pages about authentication"

### Method 2: Using .cursorrules (Project-Specific)

Create a `.cursorrules` file in your project root:

```yaml
tools:
  discovery:
    - url: http://localhost:8080/utcp
      name: "Corporate Tools"
      description: "Access to Jira, Wiki, and GitLab"
```

### Method 3: Environment Variables

Set the UTCP discovery URL in your environment:

```bash
export CURSOR_UTCP_DISCOVERY_URL="http://localhost:8080/utcp"
```

## Available Commands in Cursor

Once integrated, you can use natural language commands:

### Jira Examples:
- "Find all my open Jira tickets"
- "Create a bug report in project ABC"
- "Add a comment to PROJ-123"
- "Show me high priority issues"

### Wiki/Confluence Examples:
- "Search for documentation about API authentication"
- "Get the deployment guide from wiki"
- "List all spaces I have access to"

### GitLab Examples:
- "Show my open merge requests"
- "Search for code containing 'validateUser'"
- "Get the README from project xyz"
- "Check pipeline status for my latest commit"

## Troubleshooting

### Server Not Accessible
- Ensure RH-UTCP server is running: `curl http://localhost:8080/health`
- Check firewall settings
- Verify port 8080 is not blocked

### Tools Not Appearing
- Restart Cursor after adding the configuration
- Check Cursor's console for any errors (View → Developer Tools)
- Verify the discovery endpoint: `curl http://localhost:8080/utcp`

### Authentication Issues
- Update your `.env` file with correct credentials
- Restart the RH-UTCP server after changes
- Check server logs for authentication errors

## Security Considerations

1. **Local Development**: The server runs on localhost by default
2. **Credentials**: Stored in `.env` file (never commit this)
3. **HTTPS**: For production, use HTTPS and proper certificates
4. **Access Control**: Consider adding authentication to the UTCP server

## Advanced Configuration

### Custom Port
If using a different port, update the discovery URL:
```
http://localhost:YOUR_PORT/utcp
```

### Multiple Environments
Create different `.env` files:
- `.env.development`
- `.env.staging`
- `.env.production`

### Filtering Tools
You can disable specific providers in the configuration if needed. 