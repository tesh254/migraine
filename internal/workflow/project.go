package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tesh254/migraine/internal/storage/sqlite"
	"gopkg.in/yaml.v3"
)

// YAMLStep represents a step in a YAML workflow
type YAMLStep struct {
	Command     string  `yaml:"command" json:"command"`
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`
	OnFail      string  `yaml:"on_fail,omitempty" json:"on_fail,omitempty"`
	OnSuccess   string  `yaml:"on_success,omitempty" json:"on_success,omitempty"`
}

// YAMLConfig represents configuration for a YAML workflow
type YAMLConfig struct {
	Variables      map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	StoreVariables bool                   `yaml:"store_variables,omitempty" json:"store_variables,omitempty"`
	StoreLogs      bool                   `yaml:"store_logs,omitempty" json:"store_logs,omitempty"`
	Background     bool                   `yaml:"background,omitempty" json:"background,omitempty"`
	Global         bool                   `yaml:"global,omitempty" json:"global,omitempty"`
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
	// Look for Migraine file first
	if _, err := os.Stat("Migraine"); err == nil {
		return loadProjectWorkflowFromMigraine("Migraine")
	}

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

// loadProjectWorkflowFromMigraine loads a workflow from a Migraine file
func loadProjectWorkflowFromMigraine(filePath string) (*YAMLWorkflow, error) {
	parser, err := NewMigraineParser(filePath)
	if err != nil {
		return nil, err
	}

	wf, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	yamlWf := ConvertInternalToYAML(wf, "")
	yamlWf.Path = filePath

	return yamlWf, nil
}

// ValidateWorkflowName checks if the workflow name is valid
func ValidateWorkflowName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}
	// Add more validation if needed
	return nil
}

// UpsertProjectWorkflowToDB upserts the project workflow to the database
func UpsertProjectWorkflowToDB(wf *YAMLWorkflow, storage *sqlite.StorageService) error {
	// Reconstruct ProjectConfig to use as metadata
	config := ProjectConfig{
		Name:        wf.Name,
		Description: wf.Description,
		PreChecks:   wf.PreChecks,
		Steps:       wf.Steps,
		Actions:     wf.Actions,
		Config:      wf.Config,
		UseVault:    wf.UseVault,
	}

	// Convert to map for metadata
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %v", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(configBytes, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal project config: %v", err)
	}

	// Create DB workflow structure
	// We use the workflow name as the ID for project workflows to ensure uniqueness
	// and to make it easy to look up.
	dbWf := sqlite.Workflow{
		ID:       wf.Name,
		Name:     wf.Name,
		Path:     wf.Path,
		UseVault: wf.UseVault,
		Metadata: metadata,
	}

	// Check if workflow exists in DB
	existing, err := storage.WorkflowStore().GetWorkflow(wf.Name)
	if err == nil && existing != nil {
		// Update existing workflow
		dbWf.ID = existing.ID // Keep existing ID if different (though we set it to name above)
		return storage.WorkflowStore().UpdateWorkflow(dbWf)
	}

	// Create new workflow
	return storage.WorkflowStore().CreateWorkflow(dbWf)
}
