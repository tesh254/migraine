package workflow

import (
	"gopkg.in/yaml.v3"
	"os"
)

// YAMLWorkflow represents a workflow defined in YAML format
type YAMLWorkflow struct {
	Name        string                 `yaml:"name"`
	Description *string                `yaml:"description,omitempty"`
	PreChecks   []YAMLStep             `yaml:"pre_checks,omitempty"`
	Steps       []YAMLStep             `yaml:"steps"`
	Actions     map[string]YAMLStep    `yaml:"actions,omitempty"`
	Config      YAMLConfig             `yaml:"config,omitempty"`
	UseVault    bool                   `yaml:"use_vault,omitempty"`
	Path        string                 `json:"-"` // Not stored in the YAML, but used for file location
}

// LoadYAMLWorkflow loads a workflow from a YAML file
func LoadYAMLWorkflow(filePath string) (*YAMLWorkflow, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var workflow YAMLWorkflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}

	workflow.Path = filePath

	return &workflow, nil
}

// SaveYAMLWorkflow saves a workflow to a YAML file
func SaveYAMLWorkflow(workflow *YAMLWorkflow, filePath string) error {
	data, err := yaml.Marshal(&workflow)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ConvertYAMLToInternal converts a YAML workflow to the internal storage format
func ConvertYAMLToInternal(yamlWf *YAMLWorkflow) (*Workflow, error) {
	// Convert YAMLStep to internal Atom format
	preChecks := make([]Atom, len(yamlWf.PreChecks))
	for i, step := range yamlWf.PreChecks {
		preChecks[i] = Atom{
			Command:     step.Command,
			Description: step.Description,
		}
	}

	steps := make([]Atom, len(yamlWf.Steps))
	for i, step := range yamlWf.Steps {
		steps[i] = Atom{
			Command:     step.Command,
			Description: step.Description,
		}
	}

	actions := make(map[string]Atom)
	for name, action := range yamlWf.Actions {
		actions[name] = Atom{
			Command:     action.Command,
			Description: action.Description,
		}
	}

	// Convert YAMLConfig to internal Config
	config := Config{
		Variables:      yamlWf.Config.Variables,
		StoreVariables: yamlWf.Config.StoreVariables,
	}

	return &Workflow{
		Name:        yamlWf.Name,
		Description: yamlWf.Description,
		PreChecks:   preChecks,
		Steps:       steps,
		Actions:     actions,
		Config:      config,
	}, nil
}

// ConvertInternalToYAML converts an internal workflow to YAML format
func ConvertInternalToYAML(internalWf *Workflow, id string) *YAMLWorkflow {
	// Convert internal Atom to YAMLStep
	preChecks := make([]YAMLStep, len(internalWf.PreChecks))
	for i, step := range internalWf.PreChecks {
		preChecks[i] = YAMLStep{
			Command:     step.Command,
			Description: step.Description,
		}
	}

	steps := make([]YAMLStep, len(internalWf.Steps))
	for i, step := range internalWf.Steps {
		steps[i] = YAMLStep{
			Command:     step.Command,
			Description: step.Description,
		}
	}

	actions := make(map[string]YAMLStep)
	for name, action := range internalWf.Actions {
		actions[name] = YAMLStep{
			Command:     action.Command,
			Description: action.Description,
		}
	}

	// Convert internal Config to YAMLConfig
	config := YAMLConfig{
		Variables:      internalWf.Config.Variables,
		StoreVariables: internalWf.Config.StoreVariables,
	}

	return &YAMLWorkflow{
		Name:        internalWf.Name,
		Description: internalWf.Description,
		PreChecks:   preChecks,
		Steps:       steps,
		Actions:     actions,
		Config:      config,
	}
}