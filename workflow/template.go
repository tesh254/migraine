package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tesh254/migraine/kv"
	"github.com/tesh254/migraine/utils"
)

func (wm *WorkflowMapper) CreateWorkflowTemplate(templatePath string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v", err)
	}

	envVars, err := utils.ExtractEnvVarsFromJSON(string(content))
	if err != nil {
		return fmt.Errorf("failed to extract template variables: %v", err)
	}

	if len(envVars) > 0 {
		utils.LogInfo("Potential environment variables found:")
		for _, v := range envVars {
			utils.LogInfo("- " + v)
		}
		utils.LogInfo("Please ensure these are set before executing the workflow.")
	}

	slug := filepath.Base(templatePath)
	slug = slug[:len(slug)-len(filepath.Ext(slug))]

	template := kv.TemplateItem{
		Slug:     slug,
		Workflow: string(content),
	}

	err = kv.CreateTemplateSafe(template)
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	utils.LogSuccess(fmt.Sprintf("Workflow template '%s' created successfully", slug))
	return nil
}
