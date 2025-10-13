package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScaffoldYAMLWorkflow creates a new YAML workflow file with commented sections
func ScaffoldYAMLWorkflow(name string, description string) error {
	// Format the name to be a valid filename
	formattedName := formatWorkflowName(name)

	// Create the workflows directory if it doesn't exist
	if err := os.MkdirAll("./workflows", 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %v", err)
	}

	// Create the file path
	filePath := filepath.Join("./workflows", formattedName+".yaml")

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("workflow file %s already exists", filePath)
	}

	// Create the content with comments
	content := fmt.Sprintf(`# %s - %s
# 
# This is a YAML workflow definition for Migraine.
# 
# Fields:
# - name: The name of the workflow (required)
# - description: A brief description of what this workflow does (optional)
# - pre_checks: Commands to run before executing the main steps (optional)
# - steps: The main commands to execute in the workflow (required)
# - actions: Additional commands that can be run independently (optional)
# - config: Configuration options for the workflow (optional)
# - use_vault: Whether to use the vault for variable resolution (default: false)

name: %s
description: %s

# Pre-checks are commands that run before the main workflow steps.
# They are typically used for validation, checking if required tools are available, etc.
# If any pre-check fails, the workflow execution stops.
pre_checks:
  # - command: "which git"  # Check if git is available
  #   description: "Verify git is installed"
  # - command: "test -d {{project_dir}}"  # Example using variables
  #   description: "Verify project directory exists"

# Steps are the main commands that make up the workflow.
# They will be executed in the order they appear.
steps:
  # Example step
  - command: "echo 'Hello from workflow %s!'" 
    description: "Example step that prints a message"
  # Add more steps as needed
  # - command: "npm install"
  #   description: "Install dependencies"
  # - command: "npm run build"
  #   description: "Build the project"

# Actions are optional commands that can be run independently of the main workflow.
# They are typically used for cleanup, deployment, or other operations.
actions:
  # Example action
  # cleanup:
  #   command: "rm -rf ./temp"
  #   description: "Clean up temporary files"
  # deploy:
  #   command: "rsync -av ./dist/ user@server:/path/to/deploy/"
  #   description: "Deploy built files to server"

# Configuration options for the workflow
config:
  # Variables that can be used in the workflow
  # These will be prompted for when running the workflow if not provided
  variables:
    # project_name:
    #   - "slugify"  # Apply slugify transformation to the variable value
  # Whether to store variables in the workflow (true) or prompt for them each time (false)
  store_variables: false

# Whether to use the vault for variable resolution
# If true, variables will be resolved from the vault
# If false, variables will be resolved from environment files or prompted
use_vault: false
`, name, time.Now().Format("2006-01-02 15:04:05"), name, description, name)

	// Write the content to the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %v", err)
	}

	fmt.Printf("Workflow '%s' scaffolded successfully at %s\n", name, filePath)
	return nil
}

// ScaffoldProjectConfig creates a project configuration file (migraine.yml or migraine.json)
func ScaffoldProjectConfig(format string, description string) error {
	var content string
	var fileName string

	if format == "json" {
		fileName = "migraine.json"

		// Create the example JSON content
		descriptionPtr := &description
		config := ProjectConfig{
			Name:        "project-name",
			Description: descriptionPtr,
			PreChecks: []YAMLStep{
				{
					Command:     "which npm",
					Description: stringPtr("Verify npm is available"),
				},
				{
					Command:     "test -f package.json",
					Description: stringPtr("Verify package.json exists"),
				},
			},
			Steps: []YAMLStep{
				{
					Command:     "npm install",
					Description: stringPtr("Install dependencies"),
				},
				{
					Command:     "npm run build",
					Description: stringPtr("Build the project"),
				},
				{
					Command:     "npm run test",
					Description: stringPtr("Run tests"),
				},
			},
			Actions: map[string]YAMLStep{
				"deploy": {
					Command:     "npm run deploy",
					Description: stringPtr("Deploy the application"),
				},
				"cleanup": {
					Command:     "rm -rf ./dist",
					Description: stringPtr("Clean up build artifacts"),
				},
			},
			Config: YAMLConfig{
				Variables:      map[string]interface{}{},
				StoreVariables: false,
			},
			UseVault: true,
		}

		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON configuration: %v", err)
		}

		content = string(jsonBytes)
	} else {
		// Default to YAML
		fileName = "migraine.yaml"

		// Create the example YAML content
		content = fmt.Sprintf(`# %s - Project-level workflow configuration
# 
# This is a project configuration file that can be used with 'migraine run'
# when no workflow name is specified. The file should be named migraine.yml
# or migraine.json and placed in your project root directory.
# 
# This workflow will be automatically upserted to the database when run with 'migraine run'
# without a name, making it available globally thereafter.

name: project-name
description: %s

# Pre-checks run before the main workflow steps
# They are used for validation and dependency checks
pre_checks:
  # Verify Node.js and npm are available
  - command: "which npm"
    description: "Verify npm is available"
  # Check if package.json exists
  - command: "test -f package.json"
    description: "Verify package.json exists"
  # Check if git repository is clean
  - command: "git diff --quiet || (echo 'Warning: uncommitted changes' && false)"
    description: "Ensure git repository is clean"

# Steps are the main commands that make up the workflow
steps:
  # Install dependencies
  - command: "npm install"
    description: "Install project dependencies"
  # Build the project
  - command: "npm run build"
    description: "Build the application"
  # Run tests
  - command: "npm run test"
    description: "Run project tests"
  # Run linter
  - command: "npm run lint"
    description: "Lint the code"

# Actions are optional commands that can be run independently
actions:
  deploy:
    command: "npm run deploy"
    description: "Deploy the application to production"
  cleanup:
    command: "rm -rf ./dist ./build"
    description: "Clean up build artifacts"
  setup-dev:
    command: "npm install && npm run setup"
    description: "Set up development environment"

# Configuration for the workflow
config:
  # Variables that can be used in the workflow
  variables:
    environment:
      - "required"  # This variable is required
  # Whether to store variables in the workflow or prompt for them
  store_variables: false

# Whether to use the vault for variable resolution
# When true, variables are resolved from the vault system
# When false, variables are loaded from environment files or prompted
use_vault: true
`, time.Now().Format("2006-01-02 15:04:05"), description)
	}

	// Check if file already exists
	if _, err := os.Stat(fileName); err == nil {
		return fmt.Errorf("project configuration file %s already exists", fileName)
	}

	// Write the content to the file
	if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write project configuration file: %v", err)
	}

	fmt.Printf("Project configuration file '%s' created successfully\n", fileName)
	return nil
}

// stringPtr helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// formatWorkflowName formats a workflow name to be a valid filename
func formatWorkflowName(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove invalid characters
	var validName strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			validName.WriteRune(r)
		}
	}

	return validName.String()
}
