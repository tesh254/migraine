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

func (wm *WorkflowMapper) RunWorkflow(name string) error {
	return nil
}
