package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tesh254/migraine/internal/storage/sqlite"
	"gopkg.in/yaml.v3"
)

// YAMLStep represents a step in a YAML workflow
type YAMLStep struct {
	Command     string  `yaml:"command" json:"command"`
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
}

// YAMLConfig represents configuration for a YAML workflow
type YAMLConfig struct {
	Variables      map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	StoreVariables bool                   `yaml:"store_variables,omitempty" json:"store_variables,omitempty"`
}

// ProjectConfig represents the structure of migraine.yml or migraine.json
type ProjectConfig struct {
	Name        string              `yaml:"name" json:"name"`
	Description *string             `yaml:"description,omitempty" json:"description,omitempty"`
	PreChecks   []YAMLStep          `yaml:"pre_checks,omitempty" json:"pre_checks,omitempty"`
	Steps       []YAMLStep          `yaml:"steps" json:"steps"`
	Actions     map[string]YAMLStep `yaml:"actions,omitempty" json:"actions,omitempty"`
	Config      YAMLConfig          `yaml:"config,omitempty" json:"config,omitempty"`
	UseVault    bool                `yaml:"use_vault,omitempty" json:"use_vault,omitempty"`
	EnvFile     string              `yaml:"env_file,omitempty" json:"env_file,omitempty"`
}

// LoadProjectWorkflow loads a workflow from migraine.yml or migraine.json in the current directory
func LoadProjectWorkflow() (*YAMLWorkflow, error) {
	// Look for migraine.yml or migraine.json in current directory
	possibleFiles := []string{"migraine.yml", "migraine.yaml", "migraine.json"}

	for _, filename := range possibleFiles {
		filePath := filepath.Join(".", filename)

		if _, err := os.Stat(filePath); err == nil {
			// File exists, load it
			if strings.HasSuffix(filename, ".json") {
				return loadProjectWorkflowFromJSON(filePath)
			} else {
				return loadProjectWorkflowFromYAML(filePath)
			}
		}
	}

	return nil, fmt.Errorf("no migraine.yml or migraine.json file found in current directory")
}

// loadProjectWorkflowFromYAML loads a workflow from a YAML file
func loadProjectWorkflowFromYAML(filePath string) (*YAMLWorkflow, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Convert ProjectConfig to YAMLWorkflow
	workflow := &YAMLWorkflow{
		Name:        config.Name,
		Description: config.Description,
		PreChecks:   config.PreChecks,
		Steps:       config.Steps,
		Actions:     config.Actions,
		Config:      config.Config,
		UseVault:    config.UseVault,
		Path:        filePath,
	}

	return workflow, nil
}

// loadProjectWorkflowFromJSON loads a workflow from a JSON file
func loadProjectWorkflowFromJSON(filePath string) (*YAMLWorkflow, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Convert ProjectConfig to YAMLWorkflow
	workflow := &YAMLWorkflow{
		Name:        config.Name,
		Description: config.Description,
		PreChecks:   config.PreChecks,
		Steps:       config.Steps,
		Actions:     config.Actions,
		Config:      config.Config,
		UseVault:    config.UseVault,
		Path:        filePath,
	}

	return workflow, nil
}

// UpsertProjectWorkflowToDB upserts the project workflow to the database
func UpsertProjectWorkflowToDB(yamlWf *YAMLWorkflow, storage *sqlite.StorageService) error {
	// Convert YAML workflow to internal format
	internalWf, err := ConvertYAMLToInternal(yamlWf)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to internal format: %v", err)
	}

	// Convert internal workflow to the format used by the DB
	metadata := map[string]interface{}{
		"pre_checks":  internalWf.PreChecks,
		"steps":       internalWf.Steps,
		"actions":     internalWf.Actions,
		"config":      internalWf.Config,
		"description": internalWf.Description,
	}

	// Create the workflow record
	newWorkflow := sqlite.Workflow{
		ID:       yamlWf.Name, // Use the name as the ID
		Name:     yamlWf.Name,
		Path:     yamlWf.Path,
		UseVault: yamlWf.UseVault,
		Metadata: metadata, // Store as the map for the DB
	}

	// Check if workflow already exists in DB
	_, err = storage.WorkflowStore().GetWorkflow(yamlWf.Name)
	if err != nil {
		// If not found, create new workflow
		err = storage.WorkflowStore().CreateWorkflow(newWorkflow)
		if err != nil {
			return err
		}
	} else {
		// If found, update existing workflow
		err = storage.WorkflowStore().UpdateWorkflow(newWorkflow)
		if err != nil {
			return err
		}
	}

	// Store the working directory as a vault entry
	if yamlWf.Path != "" {
		// Extract directory from the path and convert to absolute path
		dir := filepath.Dir(yamlWf.Path)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			// If we can't get absolute path, use the relative one
			absDir = dir
		}
		
		// Create or update the WORKING_DIR vault entry for this workflow
		workingDirEntry := sqlite.VaultEntry{
			Key:        "WORKING_DIR",
			Value:      absDir,
			Scope:      "workflow",
			WorkflowID: &yamlWf.Name,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		
		// Check if a WORKING_DIR entry already exists for this workflow
		existingEntry, err := storage.VaultStore().GetVariable("WORKING_DIR", "workflow", &yamlWf.Name)
		if err == nil && existingEntry != nil {
			// Entry exists, update it
			err = storage.VaultStore().UpdateVariable("WORKING_DIR", "workflow", &yamlWf.Name, absDir)
			if err != nil {
				// Log the error but don't fail the entire operation
				fmt.Printf("Warning: failed to update WORKING_DIR vault entry: %v\n", err)
			}
		} else {
			// Entry doesn't exist, create it
			err = storage.VaultStore().CreateVariable(workingDirEntry)
			if err != nil {
				// Log the error but don't fail the entire operation
				fmt.Printf("Warning: failed to create WORKING_DIR vault entry: %v\n", err)
			}
		}
	}

	return nil
}

// ValidateWorkflowName checks if the workflow name is a valid slug
func ValidateWorkflowName(name string) error {
	if name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}

	// Check if name is a valid slug (alphanumeric, hyphens, underscores)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("workflow name '%s' contains invalid characters, only alphanumeric, hyphens, and underscores are allowed", name)
		}
	}

	return nil
}
