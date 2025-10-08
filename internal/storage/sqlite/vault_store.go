package sqlite

import (
	"database/sql"
	"fmt"
	"time"
)

type VaultStore struct {
	dbService *DBService
}

func NewVaultStore(dbService *DBService) *VaultStore {
	return &VaultStore{dbService: dbService}
}

func (vs *VaultStore) CreateVariable(entry VaultEntry) error {
	query := `
		INSERT INTO vault (key, value, scope, workflow_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	var workflowID *string
	if entry.WorkflowID != nil {
		workflowID = entry.WorkflowID
	}
	
	_, err := vs.dbService.db.Exec(
		query,
		entry.Key,
		entry.Value,
		entry.Scope,
		workflowID,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create vault entry: %v", err)
	}

	return nil
}

func (vs *VaultStore) GetVariable(key, scope string, workflowID *string) (*VaultEntry, error) {
	var query string
	var rows *sql.Rows
	var err error
	
	if workflowID != nil {
		query = `SELECT id, key, value, scope, workflow_id, created_at, updated_at FROM vault WHERE key = ? AND scope = ? AND workflow_id = ?`
		rows, err = vs.dbService.db.Query(query, key, scope, *workflowID)
	} else {
		query = `SELECT id, key, value, scope, workflow_id, created_at, updated_at FROM vault WHERE key = ? AND scope = ? AND workflow_id IS NULL`
		rows, err = vs.dbService.db.Query(query, key, scope)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get vault entry: %v", err)
	}
	defer rows.Close()

	var entry VaultEntry
	var workflowIdPtr *string
	
	if rows.Next() {
		err := rows.Scan(
			&entry.ID,
			&entry.Key,
			&entry.Value,
			&entry.Scope,
			&workflowIdPtr,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vault entry: %v", err)
		}
		
		entry.WorkflowID = workflowIdPtr
		return &entry, nil
	}
	
	return nil, fmt.Errorf("variable with key '%s' and scope '%s' not found", key, scope)
}

// GetVariableWithFallback implements the fallback logic: workflow -> project -> global
func (vs *VaultStore) GetVariableWithFallback(key, workflowID string) (*VaultEntry, error) {
	// Try workflow scope first
	entry, err := vs.GetVariable(key, "workflow", &workflowID)
	if err == nil {
		return entry, nil
	}
	
	// Try project scope
	entry, err = vs.GetVariable(key, "project", nil)
	if err == nil {
		return entry, nil
	}
	
	// Try global scope
	return vs.GetVariable(key, "global", nil)
}

func (vs *VaultStore) UpdateVariable(key, scope string, workflowID *string, value string) error {
	var query string
	
	if workflowID != nil {
		query = `UPDATE vault SET value = ?, updated_at = ? WHERE key = ? AND scope = ? AND workflow_id = ?`
	} else {
		query = `UPDATE vault SET value = ?, updated_at = ? WHERE key = ? AND scope = ? AND workflow_id IS NULL`
	}
	
	result, err := vs.dbService.db.Exec(
		query,
		value,
		time.Now(),
		key,
		scope,
		workflowID,
	)
	if err != nil {
		return fmt.Errorf("failed to update vault entry: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variable with key '%s' and scope '%s' not found", key, scope)
	}

	return nil
}

func (vs *VaultStore) DeleteVariable(key, scope string, workflowID *string) error {
	var query string
	
	if workflowID != nil {
		query = `DELETE FROM vault WHERE key = ? AND scope = ? AND workflow_id = ?`
	} else {
		query = `DELETE FROM vault WHERE key = ? AND scope = ? AND workflow_id IS NULL`
	}
	
	result, err := vs.dbService.db.Exec(query, key, scope, workflowID)
	if err != nil {
		return fmt.Errorf("failed to delete vault entry: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variable with key '%s' and scope '%s' not found", key, scope)
	}

	return nil
}

func (vs *VaultStore) ListVariables(scope string, workflowID *string) ([]VaultEntry, error) {
	var query string
	
	if workflowID != nil {
		query = `SELECT id, key, value, scope, workflow_id, created_at, updated_at FROM vault WHERE scope = ? AND workflow_id = ? ORDER BY key`
	} else if scope != "" {
		query = `SELECT id, key, value, scope, workflow_id, created_at, updated_at FROM vault WHERE scope = ? AND workflow_id IS NULL ORDER BY key`
	} else {
		query = `SELECT id, key, value, scope, workflow_id, created_at, updated_at FROM vault ORDER BY scope, key`
	}
	
	var rows *sql.Rows
	var err error
	
	if workflowID != nil {
		rows, err = vs.dbService.db.Query(query, scope, *workflowID)
	} else if scope != "" {
		rows, err = vs.dbService.db.Query(query, scope)
	} else {
		rows, err = vs.dbService.db.Query(query)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list vault entries: %v", err)
	}
	defer rows.Close()

	var entries []VaultEntry
	for rows.Next() {
		var entry VaultEntry
		var workflowIdPtr *string
		
		err := rows.Scan(
			&entry.ID,
			&entry.Key,
			&entry.Value,
			&entry.Scope,
			&workflowIdPtr,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vault entry: %v", err)
		}
		
		entry.WorkflowID = workflowIdPtr
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAllVariablesForWorkflow returns all variables that apply to a specific workflow
// including workflow-specific, project, and global variables
func (vs *VaultStore) GetAllVariablesForWorkflow(workflowID string) (map[string]string, error) {
	variables := make(map[string]string)
	
	// Get workflow-specific variables
	workflowVars, err := vs.ListVariables("workflow", &workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow variables: %v", err)
	}
	
	for _, entry := range workflowVars {
		variables[entry.Key] = entry.Value
	}
	
	// Get project variables
	projectVars, err := vs.ListVariables("project", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project variables: %v", err)
	}
	
	for _, entry := range projectVars {
		// Only add if key doesn't already exist (workflow variables have priority)
		if _, exists := variables[entry.Key]; !exists {
			variables[entry.Key] = entry.Value
		}
	}
	
	// Get global variables
	globalVars, err := vs.ListVariables("global", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get global variables: %v", err)
	}
	
	for _, entry := range globalVars {
		// Only add if key doesn't already exist
		if _, exists := variables[entry.Key]; !exists {
			variables[entry.Key] = entry.Value
		}
	}
	
	return variables, nil
}