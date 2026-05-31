package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/rafaribe/portabase-mcp/internal/core"
	mcpserver "github.com/rafaribe/portabase-mcp/internal/mcp"
	"github.com/rafaribe/portabase-mcp/internal/portabase"
)

// TestE2E runs against a real Portabase instance.
// Set E2E=1 PORTABASE_BASE_URL and PORTABASE_API_TOKEN to run.
func TestE2E(t *testing.T) {
	if os.Getenv("E2E") == "" {
		t.Skip("set E2E=1 to run e2e tests")
	}

	baseURL := os.Getenv("PORTABASE_BASE_URL")
	apiKey := os.Getenv("PORTABASE_API_TOKEN")
	if baseURL == "" || apiKey == "" {
		t.Fatal("PORTABASE_BASE_URL and PORTABASE_API_TOKEN are required for e2e tests")
	}

	client := portabase.NewClient(baseURL, apiKey)
	ctx := context.Background()

	makeReq := func(args map[string]any) mcp.CallToolRequest {
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		return req
	}

	// --- Agents ---
	var agentID string

	t.Run("create_agent", func(t *testing.T) {
		handler := mcpserver.CreateAgentHandler(client)
		result, err := handler(ctx, makeReq(map[string]any{"name": "e2e-test-agent"}))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
		var agent core.Agent
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agent)
		if agent.ID == "" {
			t.Fatal("expected agent ID")
		}
		agentID = agent.ID
		t.Logf("Created agent: %s (%s)", agent.Name, agent.ID)
	})

	t.Run("list_agents", func(t *testing.T) {
		handler := mcpserver.ListAgentsHandler(client)
		result, err := handler(ctx, makeReq(nil))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
		var agents []core.Agent
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &agents)
		if len(agents) == 0 {
			t.Fatal("expected at least one agent")
		}
		t.Logf("Found %d agents", len(agents))
	})

	t.Run("get_agent", func(t *testing.T) {
		if agentID == "" {
			t.Skip("no agent created")
		}
		handler := mcpserver.GetAgentHandler(client)
		result, err := handler(ctx, makeReq(map[string]any{"id": agentID}))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
	})

	t.Run("get_agent_key", func(t *testing.T) {
		if agentID == "" {
			t.Skip("no agent created")
		}
		handler := mcpserver.GetAgentKeyHandler(client)
		result, err := handler(ctx, makeReq(map[string]any{"id": agentID}))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
		key := result.Content[0].(mcp.TextContent).Text
		if key == "" {
			t.Fatal("expected non-empty edge key")
		}
		t.Logf("Got edge key: %s...", key[:min(20, len(key))])
	})

	t.Run("list_databases", func(t *testing.T) {
		handler := mcpserver.ListDatabasesHandler(client)
		result, err := handler(ctx, makeReq(nil))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
		var dbs []core.Database
		json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &dbs)
		t.Logf("Found %d databases", len(dbs))
	})

	t.Run("delete_agent", func(t *testing.T) {
		if agentID == "" {
			t.Skip("no agent created")
		}
		handler := mcpserver.DeleteAgentHandler(client)
		result, err := handler(ctx, makeReq(map[string]any{"id": agentID}))
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("error: %s", result.Content[0].(mcp.TextContent).Text)
		}
		t.Log("Agent deleted")
	})
}

// setupAPIKey logs in to Portabase and creates an API key.
// This is a helper for the test script.
func SetupAPIKey(baseURL, email, password string) (string, error) {
	// Sign in via better-auth
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	resp, err := http.Post(baseURL+"/api/auth/sign-in/email", "application/json", strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("sign-in request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("sign-in failed: %d", resp.StatusCode)
	}

	// Extract session cookie
	var sessionToken string
	for _, c := range resp.Cookies() {
		if c.Name == "better-auth.session_token" || c.Name == "session_token" {
			sessionToken = c.Value
		}
	}
	if sessionToken == "" {
		// Try to get token from response body
		var signInResp map[string]any
		json.NewDecoder(resp.Body).Decode(&signInResp)
		if token, ok := signInResp["token"].(string); ok {
			sessionToken = token
		}
	}

	if sessionToken == "" {
		return "", fmt.Errorf("no session token in sign-in response")
	}

	// Create API key
	req, _ := http.NewRequest("POST", baseURL+"/api/auth/api-key/create", strings.NewReader(`{"name":"e2e-test","configId":"standard"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: sessionToken})

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create api-key request failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		return "", fmt.Errorf("create api-key failed: %d", resp2.StatusCode)
	}

	var keyResp map[string]any
	json.NewDecoder(resp2.Body).Decode(&keyResp)

	if key, ok := keyResp["key"].(string); ok {
		return key, nil
	}

	return "", fmt.Errorf("no key in api-key response: %v", keyResp)
}
