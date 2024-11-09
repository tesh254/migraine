package kv

import (
	"encoding/json"
	"fmt"
	"strings"
)

const WorkflowPrefix = "mg_workflows:"

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
}

type Config struct {
	Variables      map[string]interface{} `json:variables`
	StoreVariables bool                   `json:"store_variables"`
}

type Workflow struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Steps       []Atom          `json:"steps"`
	Description *string         `json:"description"`
	Actions     map[string]Atom `json:"actions"`
	Config      Config          `json:"config"`
	UsesSudo    bool            `json:"uses_sudo"`
}

type WorkflowStore struct {
	store *Store
}

func workflowStringConcat(workflowString string) string {
	return fmt.Sprintf("mg_workflows:%s", workflowString)
}

func NewWorkflowStore(store *Store) *WorkflowStore {
	return &WorkflowStore{store: store}
}

func (ws *WorkflowStore) CreateWorkflow(id string, workflow Workflow) error {
	workflow.ID = id
	key := workflowStringConcat(id)
	return ws.store.Set(key, workflow)
}

func (ws *WorkflowStore) GetWorkflow(id string) (*Workflow, error) {
	var workflow Workflow
	key := workflowStringConcat(id)
	err := ws.store.Get(key, &workflow)
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (ws *WorkflowStore) UpdateWorkflow(id string, workflow Workflow) error {
	return ws.CreateWorkflow(id, workflow)
}

func (ws *WorkflowStore) DeleteWorkflow(id string) error {
	key := workflowStringConcat(id)
	return ws.store.Delete(key)
}

func (ws *WorkflowStore) ListWorkflows() ([]Workflow, error) {
	keys, err := ws.store.List(WorkflowPrefix)
	if err != nil {
		return nil, err
	}

	workflows := make([]Workflow, 0, len(keys))
	for _, key := range keys {
		var workflow Workflow
		err := ws.store.Get(key, &workflow)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func (ws *WorkflowStore) SearchWorkflows(query string) ([]Workflow, error) {
	allWorkflows, err := ws.ListWorkflows()
	if err != nil {
		return nil, err
	}

	var results []Workflow
	for _, workflow := range allWorkflows {
		if strings.Contains(strings.ToLower(workflow.Name), strings.ToLower(query)) ||
			(workflow.Description != nil && strings.Contains(strings.ToLower(*workflow.Description), strings.ToLower(query))) {
			results = append(results, workflow)
		}
	}

	return results, nil
}

func (ws *WorkflowStore) ExportWorkflow(id string) (string, error) {
	workflow, err := ws.GetWorkflow(id)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (ws *WorkflowStore) ImportWorkflow(id string, jsonData string) error {
	var workflow Workflow
	err := json.Unmarshal([]byte(jsonData), &workflow)
	if err != nil {
		return err
	}

	return ws.CreateWorkflow(id, workflow)
}

func (ws *WorkflowStore) AddStep(id string, step Atom) error {
	workflow, err := ws.GetWorkflow(id)
	if err != nil {
		return err
	}

	workflow.Steps = append(workflow.Steps, step)
	return ws.UpdateWorkflow(id, *workflow)
}

func (ws *WorkflowStore) RemoveStep(id string, index int) error {
	workflow, err := ws.GetWorkflow(id)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(workflow.Steps) {
		return fmt.Errorf("invalid step index")
	}

	workflow.Steps = append(workflow.Steps[:index], workflow.Steps[index+1:]...)
	return ws.UpdateWorkflow(id, *workflow)
}

func (ws *WorkflowStore) UpdateConfig(id string, configMap map[string]interface{}) error {
	workflow, err := ws.GetWorkflow(id)
	if err != nil {
		return err
	}

	// Create a new Config struct with proper type conversion
	newConfig := Config{}

	// Convert and assign Variables
	if vars, ok := configMap["variables"]; ok {
		if varsMap, ok := vars.(map[string]interface{}); ok {
			newConfig.Variables = varsMap
		} else {
			return fmt.Errorf("invalid variables format in config")
		}
	}

	// Convert and assign StoreVariables
	if storeVars, ok := configMap["store_variables"]; ok {
		if boolValue, ok := storeVars.(bool); ok {
			newConfig.StoreVariables = boolValue
		} else {
			return fmt.Errorf("invalid store_variables format in config")
		}
	}

	// Update the workflow with the new config
	workflow.Config = newConfig
	return ws.UpdateWorkflow(id, *workflow)
}

// You might also want to add a helper method to validate config
func validateConfig(config map[string]interface{}) error {
	// Check for required fields
	if _, hasVars := config["variables"]; !hasVars {
		return fmt.Errorf("config must contain 'variables' field")
	}

	if _, hasStoreVars := config["store_variables"]; !hasStoreVars {
		return fmt.Errorf("config must contain 'store_variables' field")
	}

	// Validate types
	if _, ok := config["variables"].(map[string]interface{}); !ok {
		return fmt.Errorf("'variables' must be a map")
	}

	if _, ok := config["store_variables"].(bool); !ok {
		return fmt.Errorf("'store_variables' must be a boolean")
	}

	return nil
}
