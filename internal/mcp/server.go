package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/rafaribe/portabase-mcp/internal/core"
)

func NewServer(port core.PortabasePort) *server.MCPServer {
	s := server.NewMCPServer("portabase-mcp", "0.1.0", server.WithToolCapabilities(true))

	// Agents
	s.AddTool(mcp.NewTool("list_agents", mcp.WithDescription("List all Portabase agents")), ListAgentsHandler(port))
	s.AddTool(mcp.NewTool("create_agent", mcp.WithDescription("Create a new agent"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Agent name")),
		mcp.WithString("organization_id", mcp.Description("Organization ID (optional)")),
	), CreateAgentHandler(port))
	s.AddTool(mcp.NewTool("get_agent", mcp.WithDescription("Get agent by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Agent UUID")),
	), GetAgentHandler(port))
	s.AddTool(mcp.NewTool("delete_agent", mcp.WithDescription("Delete an agent"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Agent UUID")),
	), DeleteAgentHandler(port))
	s.AddTool(mcp.NewTool("get_agent_key", mcp.WithDescription("Get agent edge key for connecting the agent to the server"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Agent UUID")),
	), GetAgentKeyHandler(port))

	// Databases
	s.AddTool(mcp.NewTool("list_databases", mcp.WithDescription("List all databases")), ListDatabasesHandler(port))
	s.AddTool(mcp.NewTool("get_database", mcp.WithDescription("Get database by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Database UUID")),
	), GetDatabaseHandler(port))
	s.AddTool(mcp.NewTool("get_database_status", mcp.WithDescription("Get database status including latest backup and restoration info"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Database UUID")),
	), GetDatabaseStatusHandler(port))

	// Backups
	s.AddTool(mcp.NewTool("list_backups", mcp.WithDescription("List backups for a database"),
		mcp.WithString("database_id", mcp.Required(), mcp.Description("Database UUID")),
	), ListBackupsHandler(port))
	s.AddTool(mcp.NewTool("trigger_backup", mcp.WithDescription("Trigger a manual backup for a database"),
		mcp.WithString("database_id", mcp.Required(), mcp.Description("Database UUID")),
	), TriggerBackupHandler(port))
	s.AddTool(mcp.NewTool("get_backup", mcp.WithDescription("Get a specific backup with storage details"),
		mcp.WithString("database_id", mcp.Required(), mcp.Description("Database UUID")),
		mcp.WithString("backup_id", mcp.Required(), mcp.Description("Backup UUID")),
	), GetBackupHandler(port))

	// Restore
	s.AddTool(mcp.NewTool("restore_database", mcp.WithDescription("Restore a database from a backup"),
		mcp.WithString("database_id", mcp.Required(), mcp.Description("Database UUID")),
		mcp.WithString("backup_id", mcp.Required(), mcp.Description("Backup UUID")),
		mcp.WithString("backup_storage_id", mcp.Required(), mcp.Description("Backup storage UUID")),
	), RestoreDatabaseHandler(port))

	return s
}

func jsonResult(v any) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return mcp.NewToolResultText(string(b))
}

func errResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

func arg(req mcp.CallToolRequest, key string) string {
	v, _ := req.GetArguments()[key].(string)
	return v
}

func requireArg(req mcp.CallToolRequest, key string) (string, error) {
	v := arg(req, key)
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

// --- Agent Handlers ---

func ListAgentsHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agents, err := port.ListAgents(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(agents), nil
	}
}

func CreateAgentHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := requireArg(req, "name")
		if err != nil {
			return errResult(err), nil
		}
		input := core.CreateAgentInput{Name: name}
		if orgID := arg(req, "organization_id"); orgID != "" {
			input.OrganizationID = &orgID
		}
		agent, err := port.CreateAgent(ctx, input)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(agent), nil
	}
}

func GetAgentHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireArg(req, "id")
		if err != nil {
			return errResult(err), nil
		}
		agent, err := port.GetAgent(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(agent), nil
	}
}

func DeleteAgentHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireArg(req, "id")
		if err != nil {
			return errResult(err), nil
		}
		if err := port.DeleteAgent(ctx, id); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("agent deleted"), nil
	}
}

func GetAgentKeyHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireArg(req, "id")
		if err != nil {
			return errResult(err), nil
		}
		key, err := port.GetAgentKey(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(key), nil
	}
}

// --- Database Handlers ---

func ListDatabasesHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dbs, err := port.ListDatabases(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(dbs), nil
	}
}

func GetDatabaseHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireArg(req, "id")
		if err != nil {
			return errResult(err), nil
		}
		db, err := port.GetDatabase(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(db), nil
	}
}

func GetDatabaseStatusHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireArg(req, "id")
		if err != nil {
			return errResult(err), nil
		}
		status, err := port.GetDatabaseStatus(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(status), nil
	}
}

// --- Backup Handlers ---

func ListBackupsHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dbID, err := requireArg(req, "database_id")
		if err != nil {
			return errResult(err), nil
		}
		backups, err := port.ListBackups(ctx, dbID)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(backups), nil
	}
}

func TriggerBackupHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dbID, err := requireArg(req, "database_id")
		if err != nil {
			return errResult(err), nil
		}
		backup, err := port.TriggerBackup(ctx, dbID)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(backup), nil
	}
}

func GetBackupHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dbID, err := requireArg(req, "database_id")
		if err != nil {
			return errResult(err), nil
		}
		backupID, err := requireArg(req, "backup_id")
		if err != nil {
			return errResult(err), nil
		}
		backup, err := port.GetBackup(ctx, dbID, backupID)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(backup), nil
	}
}

// --- Restore Handler ---

func RestoreDatabaseHandler(port core.PortabasePort) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dbID, err := requireArg(req, "database_id")
		if err != nil {
			return errResult(err), nil
		}
		backupID, err := requireArg(req, "backup_id")
		if err != nil {
			return errResult(err), nil
		}
		storageID, err := requireArg(req, "backup_storage_id")
		if err != nil {
			return errResult(err), nil
		}
		result, err := port.RestoreDatabase(ctx, dbID, core.RestoreInput{
			BackupID:        backupID,
			BackupStorageID: storageID,
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result), nil
	}
}
