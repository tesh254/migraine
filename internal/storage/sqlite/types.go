package sqlite

import (
	"time"
)

// Workflow represents a workflow in the system
type Workflow struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Path      string                 `json:"path" db:"path"`
	UseVault  bool                   `json:"use_vault" db:"use_vault"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"` // This will be stored as JSON string in DB
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// VaultEntry represents an encrypted variable in the vault
type VaultEntry struct {
	ID         int64     `json:"id" db:"id"`
	Key        string    `json:"key" db:"key"`
	Value      string    `json:"value" db:"value"`
	Scope      string    `json:"scope" db:"scope"` // global, project, workflow
	WorkflowID *string   `json:"workflow_id" db:"workflow_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Run represents an execution run of a workflow
type Run struct {
	ID          int64     `json:"id" db:"id"`
	WorkflowID  string    `json:"workflow_id" db:"workflow_id"`
	Status      string    `json:"status" db:"status"`
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	Logs        *string   `json:"logs" db:"logs"`
}