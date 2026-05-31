package core

import "context"

// PortabasePort defines the outbound port matching the Portabase v1 API.
type PortabasePort interface {
	// Agents
	ListAgents(ctx context.Context) ([]Agent, error)
	CreateAgent(ctx context.Context, input CreateAgentInput) (*Agent, error)
	GetAgent(ctx context.Context, id string) (*Agent, error)
	DeleteAgent(ctx context.Context, id string) error
	GetAgentKey(ctx context.Context, id string) (string, error)

	// Databases
	ListDatabases(ctx context.Context) ([]Database, error)
	GetDatabase(ctx context.Context, id string) (*Database, error)
	GetDatabaseStatus(ctx context.Context, id string) (*DatabaseStatus, error)

	// Backups
	ListBackups(ctx context.Context, databaseID string) ([]Backup, error)
	TriggerBackup(ctx context.Context, databaseID string) (*Backup, error)
	GetBackup(ctx context.Context, databaseID, backupID string) (*BackupWithStorages, error)

	// Restore
	RestoreDatabase(ctx context.Context, databaseID string, input RestoreInput) (*Restoration, error)
}
