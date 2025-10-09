package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type WorkflowStore struct {
	dbService *DBService
}

func NewWorkflowStore(dbService *DBService) *WorkflowStore {
	return &WorkflowStore{dbService: dbService}
}

func (ws *WorkflowStore) CreateWorkflow(workflow Workflow) error {
	query := `
		INSERT INTO workflows (id, name, path, use_vault, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	metadataBytes, err := json.Marshal(workflow.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow metadata: %v", err)
	}

	_, err = ws.dbService.db.Exec(
		query,
		workflow.ID,
		workflow.Name,
		workflow.Path,
		workflow.UseVault,
		string(metadataBytes),
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %v", err)
	}

	return nil
}

func (ws *WorkflowStore) GetWorkflow(id string) (*Workflow, error) {
	query := `SELECT id, name, path, use_vault, metadata, created_at, updated_at FROM workflows WHERE id = ?`

	var workflow Workflow
	var metadataBytes string

	err := ws.dbService.db.QueryRow(query, id).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Path,
		&workflow.UseVault,
		&metadataBytes,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get workflow: %v", err)
	}

	// Unmarshal metadata JSON
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataBytes), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow metadata: %v", err)
	}
	workflow.Metadata = metadata

	return &workflow, nil
}

func (ws *WorkflowStore) UpdateWorkflow(workflow Workflow) error {
	query := `
		UPDATE workflows 
		SET name = ?, path = ?, use_vault = ?, metadata = ?, updated_at = ?
		WHERE id = ?
	`

	metadataBytes, err := json.Marshal(workflow.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow metadata: %v", err)
	}

	_, err = ws.dbService.db.Exec(
		query,
		workflow.Name,
		workflow.Path,
		workflow.UseVault,
		string(metadataBytes),
		time.Now(),
		workflow.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow: %v", err)
	}

	return nil
}

func (ws *WorkflowStore) DeleteWorkflow(id string) error {
	query := `DELETE FROM workflows WHERE id = ?`

	result, err := ws.dbService.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workflow with id %s not found", id)
	}

	return nil
}

func (ws *WorkflowStore) ListWorkflows() ([]Workflow, error) {
	query := `SELECT id, name, path, use_vault, metadata, created_at, updated_at FROM workflows ORDER BY name`

	rows, err := ws.dbService.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %v", err)
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var workflow Workflow
		var metadataBytes string

		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Path,
			&workflow.UseVault,
			&metadataBytes,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %v", err)
		}

		// Unmarshal metadata JSON
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataBytes), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow metadata: %v", err)
		}
		workflow.Metadata = metadata

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}
