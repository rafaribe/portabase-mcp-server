# portabase-mcp

MCP (Model Context Protocol) server for the [Portabase](https://portabase.io) database backup & restore API. Built in Go with hexagonal architecture.

## Architecture

```
internal/
  core/
    types.go    # Domain types (Database, Backup, Agent, etc.)
    ports.go    # Outbound port interface (PortabasePort)
  portabase/
    client.go   # HTTP client adapter (implements PortabasePort)
  mcp/
    server.go   # MCP server with tool handlers (inbound adapter)
main.go         # Composition root
```

## Tools Exposed

| Tool | Description |
|------|-------------|
| `list_databases` | List all databases managed by Portabase |
| `get_database` | Get details of a specific database |
| `list_backups` | List backups for a database |
| `get_backup` | Get details of a specific backup |
| `trigger_backup` | Trigger a manual backup |
| `restore_backup` | Restore a backup to a database |
| `list_agents` | List all Portabase agents |
| `get_agent` | Get details of a specific agent |
| `list_destinations` | List all storage destinations |
| `list_schedules` | List backup schedules for a database |

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `PORTABASE_BASE_URL` | Yes | Portabase instance URL |
| `PORTABASE_API_TOKEN` | Yes | API token for authentication |
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
        "PORTABASE_API_TOKEN": "your-token"
      }
    }
  }
}
```

### SSE (HTTP)

```bash
PORTABASE_BASE_URL=https://your-portabase.example.com \
PORTABASE_API_TOKEN=your-token \
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
INTEGRATION=1 go test ./tests/... -v   # Integration tests
go build -o portabase-mcp .            # Build
```
