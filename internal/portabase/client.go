package portabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rafaribe/portabase-mcp/internal/core"
)

// Client implements core.PortabasePort via the Portabase v1 REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, httpClient: &http.Client{}}
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api/v1"+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("portabase %s %s: %d %s", method, path, resp.StatusCode, string(data))
	}
	return data, nil
}

type dataWrapper[T any] struct {
	Data T `json:"data"`
}

func get[T any](c *Client, ctx context.Context, path string) (T, error) {
	var zero T
	raw, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return zero, err
	}
	var w dataWrapper[T]
	if err := json.Unmarshal(raw, &w); err != nil {
		return zero, err
	}
	return w.Data, nil
}

func (c *Client) ListAgents(ctx context.Context) ([]core.Agent, error) {
	return get[[]core.Agent](c, ctx, "/agents")
}

func (c *Client) CreateAgent(ctx context.Context, input core.CreateAgentInput) (*core.Agent, error) {
	raw, err := c.do(ctx, http.MethodPost, "/agents", input)
	if err != nil {
		return nil, err
	}
	var w dataWrapper[core.Agent]
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return &w.Data, nil
}

func (c *Client) GetAgent(ctx context.Context, id string) (*core.Agent, error) {
	a, err := get[core.Agent](c, ctx, "/agents/"+id)
	return &a, err
}

func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, "/agents/"+id, nil)
	return err
}

func (c *Client) GetAgentKey(ctx context.Context, id string) (string, error) {
	return get[string](c, ctx, "/agents/"+id+"/key")
}

func (c *Client) ListDatabases(ctx context.Context) ([]core.Database, error) {
	return get[[]core.Database](c, ctx, "/databases")
}

func (c *Client) GetDatabase(ctx context.Context, id string) (*core.Database, error) {
	d, err := get[core.Database](c, ctx, "/databases/"+id)
	return &d, err
}

func (c *Client) GetDatabaseStatus(ctx context.Context, id string) (*core.DatabaseStatus, error) {
	s, err := get[core.DatabaseStatus](c, ctx, "/databases/"+id+"/status")
	return &s, err
}

func (c *Client) ListBackups(ctx context.Context, databaseID string) ([]core.Backup, error) {
	return get[[]core.Backup](c, ctx, "/databases/"+databaseID+"/backup")
}

func (c *Client) TriggerBackup(ctx context.Context, databaseID string) (*core.Backup, error) {
	raw, err := c.do(ctx, http.MethodPost, "/databases/"+databaseID+"/backup", nil)
	if err != nil {
		return nil, err
	}
	var w dataWrapper[core.Backup]
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return &w.Data, nil
}

func (c *Client) GetBackup(ctx context.Context, databaseID, backupID string) (*core.BackupWithStorages, error) {
	b, err := get[core.BackupWithStorages](c, ctx, "/databases/"+databaseID+"/backup/"+backupID)
	return &b, err
}

func (c *Client) RestoreDatabase(ctx context.Context, databaseID string, input core.RestoreInput) (*core.Restoration, error) {
	raw, err := c.do(ctx, http.MethodPost, "/databases/"+databaseID+"/restore", input)
	if err != nil {
		return nil, err
	}
	var w dataWrapper[core.Restoration]
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return &w.Data, nil
}
