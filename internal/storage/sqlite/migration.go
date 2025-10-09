package sqlite

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type MigrationService struct {
	sqliteDB *DBService
}

func NewMigrationService(sqliteDB *DBService) *MigrationService {
	return &MigrationService{
		sqliteDB: sqliteDB,
	}
}

// MigrateFromBadger migrates data from the old Badger storage to SQLite
func (ms *MigrationService) MigrateFromBadger() error {
	log.Println("Starting migration from Badger to SQLite...")

	// Migrate workflows
	if err := ms.migrateWorkflows(); err != nil {
		return fmt.Errorf("failed to migrate workflows: %v", err)
	}

	// Migrate templates to workflows (in the new format)
	if err := ms.migrateTemplates(); err != nil {
		return fmt.Errorf("failed to migrate templates: %v", err)
	}

	// Note: We don't migrate runs since those are execution logs
	// and wouldn't make sense to migrate

	log.Println("Migration completed successfully!")
	return nil
}

// Helper function to open Badger DB for migration
func (ms *MigrationService) openBadgerDB() (*badger.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Path to the old Badger database
	dbPath := filepath.Join(homeDir, ".migraine_db")

	opts := badger.DefaultOptions(dbPath)
	opts.ReadOnly = true // Open in read-only mode for migration

	return badger.Open(opts)
}

func (ms *MigrationService) migrateWorkflows() error {
	badgerDB, err := ms.openBadgerDB()
	if err != nil {
		// If Badger DB doesn't exist, that's fine - no migration needed
		log.Printf("No old Badger DB found to migrate from: %v", err)
		return nil
	}
	defer badgerDB.Close()

	// Define the old workflow structure
	type OldWorkflow struct {
		ID          string                 `json:"id"`
		Name        string                 `json:"name"`
		PreChecks   []interface{}          `json:"pre_checks"`
		Steps       []interface{}          `json:"steps"`
		Description *string                `json:"description"`
		Actions     map[string]interface{} `json:"actions"`
		Config      interface{}            `json:"config"`
		UsesSudo    bool                   `json:"uses_sudo"`
	}

	// List all workflow keys from Badger
	var workflowKeys []string
	err = badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("mg_workflows:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			workflowKeys = append(workflowKeys, key)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to list workflow keys: %v", err)
	}

	workflowStore := NewWorkflowStore(ms.sqliteDB)

	for _, key := range workflowKeys {
		var workflowData []byte

		// Get the workflow data from Badger
		err = badgerDB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}

			return item.Value(func(val []byte) error {
				workflowData = make([]byte, len(val))
				copy(workflowData, val)
				return nil
			})
		})

		if err != nil {
			log.Printf("Warning: failed to get workflow data for key %s: %v", key, err)
			continue
		}

		// Unmarshal the old workflow
		var oldWf OldWorkflow
		if err := json.Unmarshal(workflowData, &oldWf); err != nil {
			log.Printf("Warning: failed to unmarshal old workflow: %v", err)
			continue
		}

		// Convert old workflow to new format
		metadata := map[string]interface{}{
			"pre_checks":  oldWf.PreChecks,
			"steps":       oldWf.Steps,
			"actions":     oldWf.Actions,
			"config":      oldWf.Config,
			"uses_sudo":   oldWf.UsesSudo,
			"description": oldWf.Description,
		}

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			log.Printf("Warning: failed to marshal metadata for workflow %s: %v", oldWf.ID, err)
			continue
		}

		// Extract the actual ID from the key (remove prefix)
		actualID := key[len("mg_workflows:"):]

		// Parse the metadata string back to map for the struct
		var parsedMetadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &parsedMetadata); err != nil {
			log.Printf("Warning: failed to parse metadata for workflow %s: %v", oldWf.ID, err)
			continue
		}

		newWorkflow := Workflow{
			ID:        actualID,
			Name:      oldWf.Name,
			Path:      "",    // Not used in old format
			UseVault:  false, // Default to false, user can update later
			Metadata:  parsedMetadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := workflowStore.CreateWorkflow(newWorkflow); err != nil {
			log.Printf("Warning: failed to migrate workflow %s: %v", actualID, err)
			continue
		}
	}

	log.Printf("Migrated %d workflows", len(workflowKeys))
	return nil
}

func (ms *MigrationService) migrateTemplates() error {
	badgerDB, err := ms.openBadgerDB()
	if err != nil {
		// If Badger DB doesn't exist, that's fine - no migration needed
		log.Printf("No old Badger DB found to migrate templates from: %v", err)
		return nil
	}
	defer badgerDB.Close()

	// Define the old template structure
	type OldTemplateItem struct {
		Slug     string `json:"slug"`
		Workflow string `json:"workflow"`
	}

	// List all template keys from Badger
	var templateKeys []string
	err = badgerDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("mg_templates:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			templateKeys = append(templateKeys, key)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to list template keys: %v", err)
	}

	workflowStore := NewWorkflowStore(ms.sqliteDB)

	for _, key := range templateKeys {
		var templateData []byte

		// Get the template data from Badger
		err = badgerDB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}

			return item.Value(func(val []byte) error {
				templateData = make([]byte, len(val))
				copy(templateData, val)
				return nil
			})
		})

		if err != nil {
			log.Printf("Warning: failed to get template data for key %s: %v", key, err)
			continue
		}

		// Unmarshal the old template
		var oldTmpl OldTemplateItem
		if err := json.Unmarshal(templateData, &oldTmpl); err != nil {
			log.Printf("Warning: failed to unmarshal old template: %v", err)
			continue
		}

		// Convert template to a workflow-like structure
		metadata := map[string]interface{}{
			"template_content": oldTmpl.Workflow,
		}

		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			log.Printf("Warning: failed to marshal metadata for template %s: %v", oldTmpl.Slug, err)
			continue
		}

		// Extract the actual slug from the key (remove prefix)
		actualSlug := key[len("mg_templates:"):]

		// Parse the metadata string back to map for the struct
		var parsedMetadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &parsedMetadata); err != nil {
			log.Printf("Warning: failed to parse metadata for template %s: %v", oldTmpl.Slug, err)
			continue
		}

		// Create a workflow entry for the template
		newWorkflow := Workflow{
			ID:        "tmpl_" + actualSlug,
			Name:      actualSlug,
			Path:      "", // Templates were stored in DB, not as files
			UseVault:  false,
			Metadata:  parsedMetadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := workflowStore.CreateWorkflow(newWorkflow); err != nil {
			log.Printf("Warning: failed to migrate template %s: %v", actualSlug, err)
			continue
		}
	}

	log.Printf("Migrated %d templates", len(templateKeys))
	return nil
}

// Add a method to check if migration is needed
func (ms *MigrationService) IsMigrationNeeded() (bool, error) {
	// Check if there are any records in the SQLite database
	workflowStore := NewWorkflowStore(ms.sqliteDB)
	workflows, err := workflowStore.ListWorkflows()
	if err != nil {
		return false, err
	}

	// If there are no workflows in SQLite, we might need to migrate
	return len(workflows) == 0, nil
}
