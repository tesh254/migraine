package sqlite

import (
	"github.com/tesh254/migraine/internal/storage/kv"
)

// RunInitialMigration checks if migration is needed and runs it if necessary
func RunInitialMigration() error {
	migrationService := NewMigrationService(GetStorageService().GetDB())
	
	needsMigration, err := migrationService.IsMigrationNeeded()
	if err != nil {
		// If we can't check if migration is needed, try to migrate anyway
		// This handles the case where old Badger db might not exist
		needsMigration = true
	}
	
	if needsMigration {
		// Check if old Badger store exists by trying to list workflows
		_, err := kv.ListWorkflowsSafe()
		if err == nil {
			// Old store exists, run migration
			if err := migrationService.MigrateFromBadger(); err != nil {
				return err
			}
		}
	}
	
	return nil
}