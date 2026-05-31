# portabase-mcp

MCP (Model Context Protocol) server for the [Portabase](https://portabase.io) v1 API. Built in Go with hexagonal architecture.

Implements all endpoints from the [Portabase OpenAPI spec](https://github.com/Portabase/portabase/blob/main/src/lib/api-v1/openapi/).

## Architecture

```
internal/
  core/
    types.go    # Domain types matching Portabase DB schema
    ports.go    # Outbound port interface (PortabasePort)
  portabase/
    client.go   # HTTP client adapter (implements PortabasePort)
  mcp/
    server.go   # MCP server with tool handlers (inbound adapter)
main.go         # Composition root
```

## Tools Exposed

### Agents

| Tool | Description |
|------|-------------|
| `list_agents` | List all Portabase agents |
| `create_agent` | Create a new agent |
| `get_agent` | Get agent by ID |
| `delete_agent` | Delete an agent |
| `get_agent_key` | Get agent edge key for connecting the agent |

### Databases

| Tool | Description |
|------|-------------|
| `list_databases` | List all databases |
| `get_database` | Get database by ID |
| `get_database_status` | Get database status (last contact, latest backup/restoration) |

### Backups

| Tool | Description |
|------|-------------|
| `list_backups` | List backups for a database |
| `trigger_backup` | Trigger a manual backup |
| `get_backup` | Get a specific backup with storage details |

### Restore

| Tool | Description |
|------|-------------|
| `restore_database` | Restore a database from a backup |

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `PORTABASE_BASE_URL` | Yes | Portabase instance URL |
| `PORTABASE_API_TOKEN` | Yes | API key (from Portabase dashboard) |
| `PORTABASE_TRANSPORT` | No | `stdio` (default) or `sse` |
| `PORT` | No | HTTP port for SSE mode (default: 8080) |

## Usage

### Stdio (default)

```json
{
  "mcpServers": {
    "portabase": {
      "command": "portabase-mcp",
      "env": {
        "PORTABASE_BASE_URL": "https://your-portabase.example.com",
        "PORTABASE_API_TOKEN": "your-api-key"
      }
    }
  }
}
```

### SSE (HTTP)

```bash
PORTABASE_BASE_URL=https://your-portabase.example.com \
PORTABASE_API_TOKEN=your-api-key \
PORTABASE_TRANSPORT=sse \
PORT=8080 \
portabase-mcp
```

### Docker

```bash
docker run -e PORTABASE_BASE_URL=... -e PORTABASE_API_TOKEN=... -e PORTABASE_TRANSPORT=sse -p 8080:8080 ghcr.io/rafaribe/portabase-mcp:rolling
```

## Development

```bash
go test ./...                          # Unit tests
INTEGRATION=1 go test ./tests/... -v   # Integration tests (uses httptest fake server)
go build -o portabase-mcp .            # Build
```
