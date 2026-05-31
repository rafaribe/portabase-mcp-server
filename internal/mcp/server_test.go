package mcp_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/rafaribe/portabase-mcp/internal/core"
	mcpserver "github.com/rafaribe/portabase-mcp/internal/mcp"
)

type mockPort struct {
	agents    []core.Agent
	databases []core.Database
	backups   []core.Backup
	err       error
}

func (m *mockPort) ListAgents(_ context.Context) ([]core.Agent, error) { return m.agents, m.err }
func (m *mockPort) CreateAgent(_ context.Context, input core.CreateAgentInput) (*core.Agent, error) {
	return &core.Agent{ID: "new-agent", Name: input.Name, CreatedAt: time.Now()}, m.err
}
func (m *mockPort) GetAgent(_ context.Context, id string) (*core.Agent, error) {
	for _, a := range m.agents {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, m.err
}
func (m *mockPort) DeleteAgent(_ context.Context, _ string) error { return m.err }
func (m *mockPort) GetAgentKey(_ context.Context, _ string) (string, error) {
	return "edge-key-123", m.err
}
func (m *mockPort) ListDatabases(_ context.Context) ([]core.Database, error) {
	return m.databases, m.err
}
func (m *mockPort) GetDatabase(_ context.Context, id string) (*core.Database, error) {
	for _, d := range m.databases {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, m.err
}
func (m *mockPort) GetDatabaseStatus(_ context.Context, _ string) (*core.DatabaseStatus, error) {
	return &core.DatabaseStatus{IsWaitingForBackup: false}, m.err
}
func (m *mockPort) ListBackups(_ context.Context, _ string) ([]core.Backup, error) {
	return m.backups, m.err
}
func (m *mockPort) TriggerBackup(_ context.Context, _ string) (*core.Backup, error) {
	return &core.Backup{ID: "bk-new", Status: "waiting"}, m.err
}
func (m *mockPort) GetBackup(_ context.Context, _, _ string) (*core.BackupWithStorages, error) {
	return &core.BackupWithStorages{Backup: core.Backup{ID: "bk-1", Status: "success"}, Storages: []core.BackupStorage{}}, m.err
}
func (m *mockPort) RestoreDatabase(_ context.Context, _ string, _ core.RestoreInput) (*core.Restoration, error) {
	return &core.Restoration{ID: "rest-1", Status: "waiting"}, m.err
}

func makeReq(args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Arguments = args
	return req
}

func TestListAgents(t *testing.T) {
	mock := &mockPort{agents: []core.Agent{{ID: "a1", Name: "prod", Slug: "prod", CreatedAt: time.Now()}}}
	handler := mcpserver.ListAgentsHandler(mock)
	result, err := handler(context.Background(), makeReq(nil))
	if err != nil || result.IsError {
		t.Fatalf("err=%v isError=%v", err, result.IsError)
	}
	var agents []core.Agent
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agents)
	if len(agents) != 1 || agents[0].ID != "a1" {
		t.Fatalf("unexpected: %v", agents)
	}
}

func TestCreateAgent(t *testing.T) {
	handler := mcpserver.CreateAgentHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"name": "new-agent"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	var agent core.Agent
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agent)
	if agent.Name != "new-agent" {
		t.Fatalf("unexpected: %v", agent)
	}
}

func TestGetAgentMissingID(t *testing.T) {
	handler := mcpserver.GetAgentHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{}))
	if !result.IsError {
		t.Fatal("expected error for missing id")
	}
}

func TestDeleteAgent(t *testing.T) {
	handler := mcpserver.DeleteAgentHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"id": "a1"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
}

func TestGetAgentKey(t *testing.T) {
	handler := mcpserver.GetAgentKeyHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"id": "a1"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if text != "edge-key-123" {
		t.Fatalf("unexpected key: %s", text)
	}
}

func TestListDatabases(t *testing.T) {
	mock := &mockPort{databases: []core.Database{{ID: "db-1", Name: "prod-pg", DBMS: "postgresql"}}}
	handler := mcpserver.ListDatabasesHandler(mock)
	result, _ := handler(context.Background(), makeReq(nil))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	var dbs []core.Database
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &dbs)
	if len(dbs) != 1 || dbs[0].DBMS != "postgresql" {
		t.Fatalf("unexpected: %v", dbs)
	}
}

func TestGetDatabaseStatus(t *testing.T) {
	handler := mcpserver.GetDatabaseStatusHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"id": "db-1"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	var status core.DatabaseStatus
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &status)
	if status.IsWaitingForBackup {
		t.Fatal("expected false")
	}
}

func TestTriggerBackup(t *testing.T) {
	handler := mcpserver.TriggerBackupHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"database_id": "db-1"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	var backup core.Backup
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &backup)
	if backup.Status != "waiting" {
		t.Fatalf("unexpected: %v", backup)
	}
}

func TestGetBackup(t *testing.T) {
	handler := mcpserver.GetBackupHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"database_id": "db-1", "backup_id": "bk-1"}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
}

func TestRestoreDatabase(t *testing.T) {
	handler := mcpserver.RestoreDatabaseHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{
		"database_id":       "db-1",
		"backup_id":         "bk-1",
		"backup_storage_id": "bs-1",
	}))
	if result.IsError {
		t.Fatal("unexpected error")
	}
	var rest core.Restoration
	json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &rest)
	if rest.Status != "waiting" {
		t.Fatalf("unexpected: %v", rest)
	}
}

func TestRestoreMissingArgs(t *testing.T) {
	handler := mcpserver.RestoreDatabaseHandler(&mockPort{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"database_id": "db-1"}))
	if !result.IsError {
		t.Fatal("expected error for missing backup_id")
	}
}
