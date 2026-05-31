package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/rafaribe/portabase-mcp/internal/core"
	mcpserver "github.com/rafaribe/portabase-mcp/internal/mcp"
	"github.com/rafaribe/portabase-mcp/internal/portabase"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run integration tests")
	}

	mux := http.NewServeMux()

	// Agents
	mux.HandleFunc("GET /api/v1/agents", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []core.Agent{
			{ID: "a1", Slug: "prod", Name: "prod-agent", Description: "Production"},
		}})
	})
	mux.HandleFunc("POST /api/v1/agents", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"data": core.Agent{ID: "a2", Slug: "new", Name: "new-agent", Description: "New"}})
	})
	mux.HandleFunc("GET /api/v1/agents/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": core.Agent{ID: r.PathValue("id"), Slug: "prod", Name: "prod-agent", Description: "Prod"}})
	})
	mux.HandleFunc("DELETE /api/v1/agents/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /api/v1/agents/{id}/key", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": "edge-key-abc"})
	})

	// Databases
	mux.HandleFunc("GET /api/v1/databases", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []core.Database{
			{ID: "db-1", Name: "prod-pg", DBMS: "postgresql", AgentID: "a1", AgentDatabaseID: "adb-1"},
		}})
	})
	mux.HandleFunc("GET /api/v1/databases/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": core.Database{ID: r.PathValue("id"), Name: "prod-pg", DBMS: "postgresql", AgentID: "a1", AgentDatabaseID: "adb-1"}})
	})
	mux.HandleFunc("GET /api/v1/databases/{id}/status", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": core.DatabaseStatus{IsWaitingForBackup: false}})
	})
	mux.HandleFunc("GET /api/v1/databases/{id}/backup", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []core.Backup{
			{ID: "bk-1", Status: "success", DatabaseID: "db-1"},
		}})
	})
	mux.HandleFunc("POST /api/v1/databases/{id}/backup", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"data": core.Backup{ID: "bk-new", Status: "waiting", DatabaseID: "db-1"}})
	})
	mux.HandleFunc("GET /api/v1/databases/{id}/backup/{backupId}", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": core.BackupWithStorages{
			Backup:   core.Backup{ID: "bk-1", Status: "success", DatabaseID: "db-1"},
			Storages: []core.BackupStorage{{ID: "bs-1", BackupID: "bk-1", StorageChannelID: "sc-1", Status: "success"}},
		}})
	})
	mux.HandleFunc("POST /api/v1/databases/{id}/restore", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{"data": core.Restoration{ID: "rest-1", Status: "waiting", BackupID: "bk-1"}})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := portabase.NewClient(ts.URL, "test-key")

	makeReq := func(args map[string]any) mcp.CallToolRequest {
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		return req
	}

	t.Run("list_agents", func(t *testing.T) {
		result, _ := mcpserver.ListAgentsHandler(client)(context.Background(), makeReq(nil))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var agents []core.Agent
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agents)
		if len(agents) != 1 || agents[0].Name != "prod-agent" {
			t.Fatalf("unexpected: %v", agents)
		}
	})

	t.Run("create_agent", func(t *testing.T) {
		result, _ := mcpserver.CreateAgentHandler(client)(context.Background(), makeReq(map[string]any{"name": "new-agent"}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var agent core.Agent
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agent)
		if agent.Name != "new-agent" {
			t.Fatalf("unexpected: %v", agent)
		}
	})

	t.Run("get_agent_key", func(t *testing.T) {
		result, _ := mcpserver.GetAgentKeyHandler(client)(context.Background(), makeReq(map[string]any{"id": "a1"}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if text != "edge-key-abc" {
			t.Fatalf("unexpected key: %s", text)
		}
	})

	t.Run("list_databases", func(t *testing.T) {
		result, _ := mcpserver.ListDatabasesHandler(client)(context.Background(), makeReq(nil))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var dbs []core.Database
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &dbs)
		if len(dbs) != 1 || dbs[0].DBMS != "postgresql" {
			t.Fatalf("unexpected: %v", dbs)
		}
	})

	t.Run("get_database_status", func(t *testing.T) {
		result, _ := mcpserver.GetDatabaseStatusHandler(client)(context.Background(), makeReq(map[string]any{"id": "db-1"}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
	})

	t.Run("trigger_backup", func(t *testing.T) {
		result, _ := mcpserver.TriggerBackupHandler(client)(context.Background(), makeReq(map[string]any{"database_id": "db-1"}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var backup core.Backup
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &backup)
		if backup.Status != "waiting" {
			t.Fatalf("unexpected: %v", backup)
		}
	})

	t.Run("get_backup_with_storages", func(t *testing.T) {
		result, _ := mcpserver.GetBackupHandler(client)(context.Background(), makeReq(map[string]any{"database_id": "db-1", "backup_id": "bk-1"}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var bws core.BackupWithStorages
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &bws)
		if len(bws.Storages) != 1 {
			t.Fatalf("unexpected storages: %v", bws)
		}
	})

	t.Run("restore_database", func(t *testing.T) {
		result, _ := mcpserver.RestoreDatabaseHandler(client)(context.Background(), makeReq(map[string]any{
			"database_id": "db-1", "backup_id": "bk-1", "backup_storage_id": "bs-1",
		}))
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		var rest core.Restoration
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &rest)
		if rest.Status != "waiting" {
			t.Fatalf("unexpected: %v", rest)
		}
	})
}
