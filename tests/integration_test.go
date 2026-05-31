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
	mux.HandleFunc("GET /api/databases", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]core.Database{
			{ID: "db-1", Name: "test-pg", Engine: "postgresql"},
		})
	})
	mux.HandleFunc("GET /api/databases/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(core.Database{ID: r.PathValue("id"), Name: "test-pg", Engine: "postgresql"})
	})
	mux.HandleFunc("GET /api/databases/{id}/backups", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]core.Backup{
			{ID: "bk-1", DatabaseID: r.PathValue("id"), Status: "completed"},
		})
	})
	mux.HandleFunc("POST /api/databases/{id}/backups", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(core.Backup{ID: "bk-new", DatabaseID: r.PathValue("id"), Status: "running"})
	})
	mux.HandleFunc("GET /api/agents", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]core.Agent{{ID: "agent-1", Name: "test-agent", Status: "online"}})
	})
	mux.HandleFunc("GET /api/destinations", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]core.Destination{{ID: "dest-1", Name: "s3-bucket", Type: "s3"}})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := portabase.NewClient(ts.URL, "test-token")

	makeReq := func(args map[string]any) mcp.CallToolRequest {
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		return req
	}

	t.Run("list_databases", func(t *testing.T) {
		handler := mcpserver.ListDatabasesHandler(client)
		result, err := handler(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		var dbs []core.Database
		if err := json.Unmarshal([]byte(text), &dbs); err != nil {
			t.Fatal(err)
		}
		if len(dbs) != 1 || dbs[0].Name != "test-pg" {
			t.Fatalf("unexpected: %v", dbs)
		}
	})

	t.Run("trigger_backup", func(t *testing.T) {
		handler := mcpserver.TriggerBackupHandler(client)
		result, err := handler(context.Background(), makeReq(map[string]any{"database_id": "db-1"}))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		var backup core.Backup
		if err := json.Unmarshal([]byte(text), &backup); err != nil {
			t.Fatal(err)
		}
		if backup.Status != "running" {
			t.Fatalf("unexpected status: %s", backup.Status)
		}
	})

	t.Run("list_agents", func(t *testing.T) {
		handler := mcpserver.ListAgentsHandler(client)
		result, err := handler(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		var agents []core.Agent
		if err := json.Unmarshal([]byte(text), &agents); err != nil {
			t.Fatal(err)
		}
		if len(agents) != 1 || agents[0].Status != "online" {
			t.Fatalf("unexpected: %v", agents)
		}
	})

	t.Run("list_destinations", func(t *testing.T) {
		handler := mcpserver.ListDestinationsHandler(client)
		result, err := handler(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		var dests []core.Destination
		if err := json.Unmarshal([]byte(text), &dests); err != nil {
			t.Fatal(err)
		}
		if len(dests) != 1 || dests[0].Type != "s3" {
			t.Fatalf("unexpected: %v", dests)
		}
	})
}
