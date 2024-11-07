package kv

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
}

type Config struct {
	Args           map[string][]string `json:"args"`
	StoreVariables bool                `json:"store_variables"`
}

type Workflow struct {
	Name        string                 `json:"name"`
	Steps       []Atom                 `json:"steps"`
	Description *string                `json:"description"`
	Actions     map[string]Atom        `json:"actions"`
	Config      map[string]interface{} `json:"config"`
	UsesSudo    bool                   `json:"uses_sudo"`
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
	keys, err := ws.store.List("workflows:")
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

func (ws *WorkflowStore) UpdateConfig(id string, config map[string]interface{}) error {
	workflow, err := ws.GetWorkflow(id)
	if err != nil {
		return err
	}

	workflow.Config = config
	return ws.UpdateWorkflow(id, *workflow)
}
