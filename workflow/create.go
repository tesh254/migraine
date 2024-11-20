package workflow

import (
	"fmt"
	"os"
	"path/filepath"
)

func NewWorkflowManager() (*WorkflowMapper, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf(":::workflow::: failed to get home directory: %v", err)
	}

	workflowDir := filepath.Join(homeDir, WORKFLOW_DIR)
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return nil, fmt.Errorf(":::workflow::: failed to create workflow directory: %v", err)
	}

	return &WorkflowMapper{WorkflowDir: workflowDir}, nil
}

// func (wm *WorkflowMapper) ListWorkflows() ([]string, error) {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return nil, fmt.Errorf(":::workflow::: failed to get home directory: %v", err)
// 	}

// 	workflowDir := filepath.Join(homeDir, WORKFLOW_DIR)
// 	files, err := os.ReadDir(workflowDir)
// 	if err != nil {
// 		return nil, fmt.Errorf(":::workflow::: failed to read workflow directory: %v", err)
// 	}
// 	var workflows []string
// 	for _, file := range files {
// 		if filepath.Ext(file.Name()) == ".json" {
// 			workflows = append(workflows, strings.TrimSuffix(file.Name(), ".json"))
// 		}
// 	}

// 	return workflows, nil
// }

// func (wm *WorkflowMapper) ExtractVariables(jsonStr string) ([]string, error) {
// 	var workflow Workflow
// 	err := json.Unmarshal([]byte(jsonStr), &workflow)
// 	if err != nil {
// 		return nil, err
// 	}

// 	variables := make(map[string]bool)

// 	for _, step := range workflow.Steps {
// 		for _, value := range step.Command {
// 			extractVarsFromString(value, variables)
// 		}
// 	}

// 	for _, action := range workflow.Actions {
// 		for _, value := range action.Command {
// 			extractVarsFromString(value, variables)
// 		}
// 	}

// 	result := make([]string, 0, len(variables))
// 	for variable := range variables {
// 		result = append(result, variable)
// 	}

// 	return result, nil
// }

// func extractVarsFromString(s string, variables map[string]bool) {
// 	parts := strings.Split(s, "{{")
// 	for _, part := range parts[1:] {
// 		if idx := strings.Index(part, "}}"); idx != -1 {
// 			variable := strings.TrimSpace(part[:idx])
// 			if utils.IsVariable(variable) {
// 				variables[variable] = true
// 			}
// 		}
// 	}
// }

// func (wm *WorkflowMapper) CreateWorkflow(name string, steps []Atom, actions map[string]Atom, description *string) error {
// 	slugifiedName := slug.Make(name)
// 	slugifiedName = strings.ToLower(slugifiedName)

// 	workflows, err := wm.ListWorkflows()
// 	if err != nil {
// 		return fmt.Errorf(":::workflow::: failed to list workflows: %v", err)
// 	}

// 	for _, exisitingName := range workflows {
// 		if strings.EqualFold(exisitingName, slugifiedName) {
// 			return fmt.Errorf(":::workflow::: workflow `%s` already exists", name)
// 		}
// 	}

// 	desc := ""
// 	if description != nil {
// 		desc = *description
// 	}

// 	workflow := Workflow{
// 		Name:        name,
// 		Steps:       steps,
// 		Description: &desc,
// 		Actions:     actions,
// 	}

// 	data, err := json.MarshalIndent(workflow, "", "  ")
// 	if err != nil {
// 		return fmt.Errorf(":::workflow::: failed to marshal workflow: %v", err)
// 	}

// 	filename := filepath.Join(wm.WorkflowDir, fmt.Sprintf("%s.json", slugifiedName))
// 	if err := os.WriteFile(filename, data, 0644); err != nil {
// 		return fmt.Errorf(":::workflow::: failed to write workflow file: %v", err)
// 	}

// 	utils.LogSuccess(fmt.Sprintf(":::workflow::: workflow `%s` created successfully (slugified name: %s)", name, slugifiedName))
// 	return nil
// }

func (wm *WorkflowMapper) RunWorkflow(name string) error {
	return nil
}
