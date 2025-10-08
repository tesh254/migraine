package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DBService struct {
	db   *sql.DB
	path string
}

func NewDBService(appName string) (*DBService, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create the application directory
	dbPath := filepath.Join(homeDir, "."+appName+"_db")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	// Full path to the SQLite file
	dbFilePath := filepath.Join(dbPath, "migraine.db")

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	service := &DBService{
		db:   db,
		path: dbFilePath,
	}

	// Run migrations
	if err := service.migrate(); err != nil {
		service.Close()
		return nil, err
	}

	return service, nil
}

func (s *DBService) DB() *sql.DB {
	return s.db
}

func (s *DBService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *DBService) migrate() error {
	// Create workflows table
	workflowTableSQL := `
	CREATE TABLE IF NOT EXISTS workflows (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		path TEXT,
		use_vault BOOLEAN DEFAULT 0,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.Exec(workflowTableSQL); err != nil {
		return fmt.Errorf("failed to create workflows table: %v", err)
	}

	// Create vault table
	vaultTableSQL := `
	CREATE TABLE IF NOT EXISTS vault (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		scope TEXT NOT NULL DEFAULT 'global',
		workflow_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workflow_id) REFERENCES workflows (id)
	);`

	if _, err := s.db.Exec(vaultTableSQL); err != nil {
		return fmt.Errorf("failed to create vault table: %v", err)
	}

	// Create unique index for vault
	vaultIndexSQL := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_key_scope ON vault(key, scope, workflow_id);`

	if _, err := s.db.Exec(vaultIndexSQL); err != nil {
		return fmt.Errorf("failed to create vault index: %v", err)
	}

	// Create runs table
	runsTableSQL := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workflow_id TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		logs TEXT,
		FOREIGN KEY (workflow_id) REFERENCES workflows (id)
	);`

	if _, err := s.db.Exec(runsTableSQL); err != nil {
		return fmt.Errorf("failed to create runs table: %v", err)
	}

	return nil
}