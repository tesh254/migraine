package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/tesh254/migraine/internal/storage/sqlite"
	"github.com/tesh254/migraine/pkg/utils"
)

// VariableResolver handles variable resolution for workflows
type VariableResolver struct {
	storage *sqlite.StorageService
}

func NewVariableResolver(storage *sqlite.StorageService) *VariableResolver {
	return &VariableResolver{
		storage: storage,
	}
}

// ResolveVariables resolves all variables for a workflow based on its configuration
func (vr *VariableResolver) ResolveVariables(workflowID string, workflowUseVault bool, flags map[string]string) (map[string]string, error) {
	variables := make(map[string]string)

	// First, use any variables provided via command line flags
	for k, v := range flags {
		variables[k] = v
	}

	// If workflow is configured to use vault, get variables from there
	if workflowUseVault {
		vaultVars, err := vr.storage.VaultStore().GetAllVariablesForWorkflow(workflowID)
		if err != nil {
			return nil, fmt.Errorf("failed to get variables from vault: %v", err)
		}

		// Merge vault variables, but command-line flags take precedence
		for k, v := range vaultVars {
			if _, exists := variables[k]; !exists {
				variables[k] = v
			}
		}
	} else {
		// If not using vault, fall back to environment files or prompting
		envVars := vr.loadEnvFileVariables(workflowID)
		
		// Merge env variables, but command-line flags and vault take precedence
		for k, v := range envVars {
			if _, exists := variables[k]; !exists {
				variables[k] = v
			}
		}
	}

	return variables, nil
}

// loadEnvFileVariables loads variables from environment files
func (vr *VariableResolver) loadEnvFileVariables(workflowID string) map[string]string {
	variables := make(map[string]string)
	
	// Look for environment files in multiple locations
	possiblePaths := []string{
		fmt.Sprintf("./env/%s.env", workflowID),
		"./.env",
		"./env/.env",
		fmt.Sprintf("./%s.env", workflowID),
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, load it
			envVars := vr.loadEnvFile(path)
			for k, v := range envVars {
				variables[k] = v
			}
			// Stop at the first file found
			break
		}
	}
	
	return variables
}

// loadEnvFile loads variables from a single .env file
func (vr *VariableResolver) loadEnvFile(path string) map[string]string {
	variables := make(map[string]string)
	
	content, err := os.ReadFile(path)
	if err != nil {
		return variables
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			
			variables[key] = value
		}
	}
	
	return variables
}

// ValidateRequiredVariables checks that all required variables are present
func (vr *VariableResolver) ValidateRequiredVariables(content string, variables map[string]string) error {
	// Extract variables from the content
	requiredVars := utils.ExtractTemplateVars(content)
	
	var missingVars []string
	for _, v := range requiredVars {
		if _, exists := variables[v]; !exists {
			missingVars = append(missingVars, v)
		}
	}
	
	if len(missingVars) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missingVars, ", "))
	}
	
	return nil
}

// ApplyVariables applies variable values to content with proper escaping
func (vr *VariableResolver) ApplyVariables(content string, variables map[string]string) (string, error) {
	// Validate that required variables are present
	if err := vr.ValidateRequiredVariables(content, variables); err != nil {
		return "", err
	}
	
	// Apply variables to the content
	result := content
	for k, v := range variables {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	
	return result, nil
}