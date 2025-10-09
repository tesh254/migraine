package sqlite

import (
	"database/sql"
	"fmt"
	"time"
)

type RunStore struct {
	dbService *DBService
}

func NewRunStore(dbService *DBService) *RunStore {
	return &RunStore{dbService: dbService}
}

func (rs *RunStore) CreateRun(run Run) error {
	query := `
		INSERT INTO runs (workflow_id, status, started_at, completed_at, logs)
		VALUES (?, ?, ?, ?, ?)
	`

	var completedAt *time.Time
	if run.CompletedAt != nil {
		completedAt = run.CompletedAt
	}

	var logs *string
	if run.Logs != nil {
		logs = run.Logs
	}

	_, err := rs.dbService.db.Exec(
		query,
		run.WorkflowID,
		run.Status,
		run.StartedAt,
		completedAt,
		logs,
	)
	if err != nil {
		return fmt.Errorf("failed to create run: %v", err)
	}

	return nil
}

func (rs *RunStore) GetRun(id int64) (*Run, error) {
	query := `SELECT id, workflow_id, status, started_at, completed_at, logs FROM runs WHERE id = ?`

	var run Run
	var completedAt *time.Time
	var logs *string

	err := rs.dbService.db.QueryRow(query, id).Scan(
		&run.ID,
		&run.WorkflowID,
		&run.Status,
		&run.StartedAt,
		&completedAt,
		&logs,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("run with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get run: %v", err)
	}

	run.CompletedAt = completedAt
	run.Logs = logs

	return &run, nil
}

func (rs *RunStore) UpdateRun(run Run) error {
	query := `
		UPDATE runs 
		SET status = ?, completed_at = ?, logs = ?
		WHERE id = ?
	`

	var completedAt *time.Time
	if run.CompletedAt != nil {
		completedAt = run.CompletedAt
	}

	var logs *string
	if run.Logs != nil {
		logs = run.Logs
	}

	_, err := rs.dbService.db.Exec(
		query,
		run.Status,
		completedAt,
		logs,
		run.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update run: %v", err)
	}

	return nil
}

func (rs *RunStore) ListRuns(workflowID string) ([]Run, error) {
	query := `SELECT id, workflow_id, status, started_at, completed_at, logs FROM runs WHERE workflow_id = ? ORDER BY started_at DESC`

	rows, err := rs.dbService.db.Query(query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %v", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var run Run
		var completedAt *time.Time
		var logs *string

		err := rows.Scan(
			&run.ID,
			&run.WorkflowID,
			&run.Status,
			&run.StartedAt,
			&completedAt,
			&logs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %v", err)
		}

		run.CompletedAt = completedAt
		run.Logs = logs
		runs = append(runs, run)
	}

	return runs, nil
}

func (rs *RunStore) ListRecentRuns(limit int) ([]Run, error) {
	query := `SELECT id, workflow_id, status, started_at, completed_at, logs FROM runs ORDER BY started_at DESC LIMIT ?`

	rows, err := rs.dbService.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent runs: %v", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var run Run
		var completedAt *time.Time
		var logs *string

		err := rows.Scan(
			&run.ID,
			&run.WorkflowID,
			&run.Status,
			&run.StartedAt,
			&completedAt,
			&logs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %v", err)
		}

		run.CompletedAt = completedAt
		run.Logs = logs
		runs = append(runs, run)
	}

	return runs, nil
}

func (rs *RunStore) DeleteRun(id int64) error {
	query := `DELETE FROM runs WHERE id = ?`

	result, err := rs.dbService.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete run: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("run with id %d not found", id)
	}

	return nil
}
