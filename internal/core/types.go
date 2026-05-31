package core

import "time"

type Agent struct {
	ID               string     `json:"id"`
	Slug             string     `json:"slug"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Version          *string    `json:"version"`
	IsArchived       bool       `json:"isArchived"`
	HealthErrorCount *int       `json:"healthErrorCount"`
	LastContact      *time.Time `json:"lastContact"`
	OrganizationID   *string    `json:"organizationId"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt"`
}

type CreateAgentInput struct {
	Name           string  `json:"name"`
	OrganizationID *string `json:"organizationId,omitempty"`
}

type Database struct {
	ID                 string     `json:"id"`
	AgentDatabaseID    string     `json:"agentDatabaseId"`
	Name               string     `json:"name"`
	DBMS               string     `json:"dbms"` // postgresql, mysql, mariadb, mongodb, sqlite, redis, valkey, firebird, mssql
	Description        *string    `json:"description"`
	BackupPolicy       *string    `json:"backupPolicy"`
	IsWaitingForBackup bool       `json:"isWaitingForBackup"`
	BackupToRestore    *string    `json:"backupToRestore"`
	HealthErrorCount   *int       `json:"healthErrorCount"`
	AgentID            string     `json:"agentId"`
	LastContact        *time.Time `json:"lastContact"`
	ProjectID          *string    `json:"projectId"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt"`
}

type Backup struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"` // waiting, ongoing, failed, success
	File       *string    `json:"file"`
	FileSize   *int64     `json:"fileSize"`
	DatabaseID string     `json:"databaseId"`
	Imported   bool       `json:"imported"`
	Migrated   bool       `json:"migrated"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
}

type BackupStorage struct {
	ID               string     `json:"id"`
	BackupID         string     `json:"backupId"`
	StorageChannelID string     `json:"storageChannelId"`
	Status           string     `json:"status"` // pending, success, failed
	Path             *string    `json:"path"`
	Size             *int64     `json:"size"`
	Checksum         *string    `json:"checksum"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt"`
}

type BackupWithStorages struct {
	Backup
	Storages []BackupStorage `json:"storages"`
}

type Restoration struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"` // waiting, ongoing, failed, success
	BackupStorageID *string    `json:"backupStorageId"`
	BackupID        string     `json:"backupId"`
	DatabaseID      *string    `json:"databaseId"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

type DatabaseStatus struct {
	IsWaitingForBackup bool         `json:"isWaitingForBackup"`
	LastContact        *time.Time   `json:"lastContact"`
	LatestBackup       *Backup      `json:"latestBackup"`
	LatestRestoration  *Restoration `json:"latestRestoration"`
}

type RestoreInput struct {
	BackupID        string `json:"backupId"`
	BackupStorageID string `json:"backupStorageId"`
}
