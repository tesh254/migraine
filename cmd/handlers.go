package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	execution "github.com/tesh254/migraine/internal/execution"
	"github.com/tesh254/migraine/internal/storage/sqlite"
	"github.com/tesh254/migraine/internal/ui"
	"github.com/tesh254/migraine/internal/workflow"
	"github.com/tesh254/migraine/pkg/utils"
)

func handleListWorkflows() {
	// List workflows from database
	storage := sqlite.GetStorageService()
	dbWorkflows, err := storage.WorkflowStore().ListWorkflows()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to list database workflows: %v", err))
		return
	}

	// List workflows from current directory
	fsWorkflows, err := workflow.DiscoverWorkflowsFromCWD()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to discover file-based workflows: %v", err))
		// Continue with just DB workflows
	}

	fmt.Printf("\n%sAvailable Workflows:%s\n\n", utils.BOLD, utils.RESET)

	// Print database workflows
	if len(dbWorkflows) > 0 {
		fmt.Printf("%sDatabase Workflows:%s\n", utils.BOLD, utils.RESET)
		for i, wf := range dbWorkflows {
			fmt.Printf("%d. (%s%s%s) %s\n", i+1, utils.BOLD, wf.ID, utils.RESET, wf.Name)

			// Parse the metadata to get more details
			// For now, we'll just show basic info
			fmt.Printf("   Source: Database\n")
			fmt.Printf("   Use Vault: %t\n", wf.UseVault)
			fmt.Println()
		}
	}

	// Print file system workflows
	if len(fsWorkflows) > 0 {
		if len(dbWorkflows) > 0 {
			fmt.Printf("\n%sFile-based Workflows:%s\n", utils.BOLD, utils.RESET)
		}

		for i, wf := range fsWorkflows {
			fmt.Printf("%d. %s\n", i+1+len(dbWorkflows), wf.Name)
			if wf.Description != nil {
				fmt.Printf("   Description: %s\n", *wf.Description)
			}
			fmt.Printf("   Source: %s\n", wf.Path)
			fmt.Printf("   Use Vault: %t\n", wf.UseVault)

			// Show required variables
			allContent := ""
			for _, step := range wf.Steps {
				allContent += step.Command + "\n"
			}
			for _, check := range wf.PreChecks {
				allContent += check.Command + "\n"
			}
			for name, action := range wf.Actions {
				allContent += fmt.Sprintf("%s: %s\n", name, action.Command)
			}

			variables := utils.ExtractTemplateVars(allContent)
			if len(variables) > 0 {
				fmt.Printf("   Required Variables: %s\n", strings.Join(variables, ", "))
			}

			fmt.Println()
		}
	}

	if len(dbWorkflows) == 0 && len(fsWorkflows) == 0 {
		fmt.Println("No workflows found in database or current directory")
	}
}

func handleRunWorkflowV2(workflowName string, cmd *cobra.Command) {
	// First, try to find the workflow in the database
	storage := sqlite.GetStorageService()
	dbWf, dbErr := storage.WorkflowStore().GetWorkflow(workflowName)

	// Also try to find the workflow in current directory
	fsWf, fsErr := workflow.FindWorkflowByName(workflowName)

	// If neither is found, error out
	if dbErr != nil && fsErr != nil {
		utils.LogError(fmt.Sprintf("Workflow '%s' not found in database or current directory", workflowName))
		os.Exit(1)
	}

	// If both are found, prefer the database one for now
	var useVault bool
	var workflowContent string

	if dbErr == nil {
		// Use database workflow
		useVault = dbWf.UseVault
		// For the content, we'll convert the metadata to a string representation
		metadataBytes, _ := json.Marshal(dbWf.Metadata)
		workflowContent = string(metadataBytes)
	} else {
		// Use file-based workflow
		useVault = fsWf.UseVault
		// For the content, we'll build a representation of all commands
		workflowContent = ""
		for _, step := range fsWf.Steps {
			workflowContent += step.Command + "\n"
		}
		for _, check := range fsWf.PreChecks {
			workflowContent += check.Command + "\n"
		}
		for name, action := range fsWf.Actions {
			workflowContent += fmt.Sprintf("%s: %s\n", name, action.Command)
		}
	}

	// Process variables from flags
	flagVars, err := cmd.Flags().GetStringArray("var")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get variables: %v", err))
		os.Exit(1)
	}

	variables := make(map[string]string)
	for _, v := range flagVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			utils.LogError(fmt.Sprintf("Invalid variable format: %s. Use KEY=VALUE format", v))
			os.Exit(1)
		}
		variables[parts[0]] = parts[1]
	}

	// Create variable resolver
	varResolver := workflow.NewVariableResolver(storage)

	// Determine workflow ID based on which workflow type we're using
	var workflowID string
	if dbErr == nil {
		workflowID = dbWf.ID
	} else {
		workflowID = workflowName
	}

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(workflowID, useVault, variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// If there are still missing variables, prompt for them if not using vault
	if !useVault {
		reader := bufio.NewReader(os.Stdin)

		// Get all required variables from the workflow content
		requiredVars := utils.ExtractTemplateVars(workflowContent)

		for _, v := range requiredVars {
			if _, exists := resolvedVars[v]; !exists {
				fmt.Printf("%s: ", v)
				value, err := reader.ReadString('\n')
				if err != nil {
					utils.LogError(fmt.Sprintf("Failed to read input: %v", err))
					os.Exit(1)
				}
				resolvedVars[v] = strings.TrimSpace(value)
			}
		}
	}

	// Execute the workflow based on its source
	if dbErr == nil {
		executeDBWorkflow(dbWf, resolvedVars)
	} else {
		executeYAMLWorkflow(fsWf, resolvedVars)
	}
}

func executeDBWorkflow(dbWf *sqlite.Workflow, variables map[string]string) {
	// Display workflow header
	ui.WorkflowHeader(dbWf.Name, "run")
	startTime := time.Now()

	// Parse metadata to get the workflow content
	metadataBytes, err := json.Marshal(dbWf.Metadata)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to marshal workflow metadata: %v", err))
		os.Exit(1)
	}

	// Convert the metadata back to a ProjectConfig (or similar structure)
	var config workflow.ProjectConfig
	if err := json.Unmarshal(metadataBytes, &config); err != nil {
		utils.LogError(fmt.Sprintf("Failed to unmarshal workflow metadata: %v", err))
		os.Exit(1)
	}

	// Create variable resolver for applying variables
	varResolver := workflow.NewVariableResolver(sqlite.GetStorageService())

	// Track precheck statistics
	precheckCount := 0
	prechecksPassed := 0
	prechecksFailed := 0
	prechecksWarn := 0

	// Run pre-checks section
	ui.SectionHeader("PRECHECKS")

	for i, check := range config.PreChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		// Execute the command using the execution package
		precheckCount++
		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")
		}
	}

	// Run steps section
	scriptCount := len(config.Steps)
	ui.SectionHeader("SCRIPTS")

	for i, step := range config.Steps {
		stepStartTime := time.Now()
		command, err := varResolver.ApplyVariables(step.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to step %d: %v", i+1, err))
			os.Exit(1)
		}

		// Display progress with elapsed time
		ui.ScriptProgress(i+1, scriptCount, *step.Description, time.Since(stepStartTime))

		// Execute the command using the execution package
		err = execution.ExecuteCommand(command)

		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Step %d failed: %v", i+1, err))
			os.Exit(1)
		}
		ui.LogInfoBordered("Step completed successfully")
	}

	// Display summary
	totalDuration := time.Since(startTime)
	ui.Summary("SUCCESS", totalDuration, prechecksPassed, prechecksFailed, prechecksWarn, scriptCount, scriptCount, "", "")

	ui.LogSuccessBordered(fmt.Sprintf("Database workflow '%s' completed successfully", dbWf.Name))
}

func executeYAMLWorkflow(yamlWf *workflow.YAMLWorkflow, variables map[string]string) {
	// Display workflow header
	ui.WorkflowHeader(yamlWf.Name, "run")
	startTime := time.Now()

	// Create variable resolver for applying variables
	varResolver := workflow.NewVariableResolver(sqlite.GetStorageService())

	// Track precheck statistics
	precheckCount := 0
	prechecksPassed := 0
	prechecksFailed := 0
	prechecksWarn := 0

	// Run pre-checks section
	ui.SectionHeader("PRECHECKS")

	for i, check := range yamlWf.PreChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		// Execute the command using the execution package
		precheckCount++
		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")
		}
	}

	// For YAML workflows without specific actions, run the main steps
	// (Note: this function doesn't currently handle specific actions like the project workflow does)
	scriptCount := len(yamlWf.Steps)
	ui.SectionHeader("SCRIPTS")

	for i, step := range yamlWf.Steps {
		stepStartTime := time.Now()
		command, err := varResolver.ApplyVariables(step.Command, variables)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to apply variables to step %d: %v", i+1, err))
			os.Exit(1)
		}

		// Display progress with elapsed time
		ui.ScriptProgress(i+1, scriptCount, *step.Description, time.Since(stepStartTime))

		// Execute the command using the execution package
		err = execution.ExecuteCommand(command)

		if err != nil {
			utils.LogError(fmt.Sprintf("Step %d failed: %v", i+1, err))
			os.Exit(1)
		}
		utils.LogInfo("Step completed successfully")
	}

	// Display summary
	totalDuration := time.Since(startTime)
	ui.Summary("SUCCESS", totalDuration, prechecksPassed, prechecksFailed, prechecksWarn, scriptCount, scriptCount, "", "")

	ui.LogSuccessBordered(fmt.Sprintf("YAML workflow '%s' completed successfully", yamlWf.Name))
}

func handleRunProjectWorkflow(cmd *cobra.Command) {
	// Look for migraine.yml or migraine.json in current directory
	projWf, err := workflow.LoadProjectWorkflow()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to load project workflow: %v", err))
		utils.LogInfo("Create a 'migraine.yml' or 'migraine.json' file in your current directory to use this feature")
		os.Exit(1)
	}

	// Validate workflow name
	if err := workflow.ValidateWorkflowName(projWf.Name); err != nil {
		utils.LogError(fmt.Sprintf("Invalid workflow name: %v", err))
		os.Exit(1)
	}

	// Get storage service
	storage := sqlite.GetStorageService()

	// Upsert the workflow to the database
	if err := workflow.UpsertProjectWorkflowToDB(projWf, storage); err != nil {
		utils.LogError(fmt.Sprintf("Failed to upsert project workflow to database: %v", err))
		os.Exit(1)
	}

	utils.LogInfo(fmt.Sprintf("Project workflow '%s' loaded and upserted to database", projWf.Name))

	// Process variables from flags
	flagVars, err := cmd.Flags().GetStringArray("var")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get variables: %v", err))
		os.Exit(1)
	}

	variables := make(map[string]string)
	for _, v := range flagVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			utils.LogError(fmt.Sprintf("Invalid variable format: %s. Use KEY=VALUE format", v))
			os.Exit(1)
		}
		variables[parts[0]] = parts[1]
	}

	// Create variable resolver
	varResolver := workflow.NewVariableResolver(storage)

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(projWf.Name, projWf.UseVault, variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// Execute the project workflow
	executeProjectYAMLWorkflow(projWf, resolvedVars, cmd)
}

func executeProjectYAMLWorkflow(yamlWf *workflow.YAMLWorkflow, variables map[string]string, cmd *cobra.Command) {
	// Display workflow header
	ui.WorkflowHeader(yamlWf.Name, "run")
	startTime := time.Now()

	// Create variable resolver for applying variables
	varResolver := workflow.NewVariableResolver(sqlite.GetStorageService())

	// Check if specific action is requested
	actionFlags, err := cmd.Flags().GetStringArray("action")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get action flags: %v", err))
		os.Exit(1)
	}

	// Track precheck statistics
	precheckCount := 0
	prechecksPassed := 0
	prechecksFailed := 0
	prechecksWarn := 0

	// Run pre-checks section (always run pre-checks regardless of whether actions or main steps are being run)
	ui.SectionHeader("PRECHECKS")

	for i, check := range yamlWf.PreChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		// Execute the command using the execution package
		precheckCount++
		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")
		}
	}

	// Check if specific action is requested (after running pre-checks)
	if len(actionFlags) > 0 {
		// Run specific action instead of main steps (but after pre-checks)
		ui.SectionHeader("ACTIONS")

		for _, actionName := range actionFlags {
			if action, exists := yamlWf.Actions[actionName]; exists {
				actionStartTime := time.Now()
				command, err := varResolver.ApplyVariables(action.Command, variables)
				if err != nil {
					ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to action %s: %v", actionName, err))
					os.Exit(1)
				}

				// Display action progress with elapsed time
				ui.ScriptProgress(1, 1, *action.Description, time.Since(actionStartTime))

				// Execute the command using the execution package
				err = execution.ExecuteCommand(command)

				if err != nil {
					ui.LogErrorBordered(fmt.Sprintf("Action '%s' failed: %v", actionName, err))
					os.Exit(1)
				}
				ui.LogInfoBordered("Action completed successfully")
			} else {
				ui.LogErrorBordered(fmt.Sprintf("Action '%s' not found in workflow", actionName))
				os.Exit(1)
			}
		}
		ui.LogSuccessBordered(fmt.Sprintf("Project workflow actions completed successfully"))
		return
	}

	// Run steps section (only if no actions were specified)
	scriptCount := len(yamlWf.Steps)
	ui.SectionHeader("SCRIPTS")

	for i, step := range yamlWf.Steps {
		stepStartTime := time.Now()
		command, err := varResolver.ApplyVariables(step.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to step %d: %v", i+1, err))
			os.Exit(1)
		}

		// Display progress with elapsed time
		ui.ScriptProgress(i+1, scriptCount, *step.Description, time.Since(stepStartTime))

		// Execute the command using the execution package
		err = execution.ExecuteCommand(command)

		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Step %d failed: %v", i+1, err))
			os.Exit(1)
		}
		ui.LogInfoBordered("Step completed successfully")
	}

	// Display summary
	totalDuration := time.Since(startTime)
	ui.Summary("SUCCESS", totalDuration, prechecksPassed, prechecksFailed, prechecksWarn, scriptCount, scriptCount, "", "")

	ui.LogSuccessBordered(fmt.Sprintf("Project workflow '%s' completed successfully", yamlWf.Name))
}

func handleWorkflowInfoV2(workflowName string) {
	// Try to find the workflow in database first
	storage := sqlite.GetStorageService()
	dbWf, dbErr := storage.WorkflowStore().GetWorkflow(workflowName)

	// Also try to find in file system
	fsWf, fsErr := workflow.FindWorkflowByName(workflowName)

	// If neither is found, error out
	if dbErr != nil && fsErr != nil {
		utils.LogError(fmt.Sprintf("Workflow '%s' not found in database or current directory", workflowName))
		os.Exit(1)
	}

	fmt.Printf("\n%s%s%s\n", utils.BOLD, workflowName, utils.RESET)

	if dbErr == nil {
		// Show database workflow info
		fmt.Printf("Source: Database\n")
		fmt.Printf("Use Vault: %t\n", dbWf.UseVault)
		fmt.Printf("Path: %s\n", dbWf.Path)

		// Parse metadata to show more details
		// For now, just show the raw metadata
		fmt.Printf("Metadata: %s\n", dbWf.Metadata)
		utils.LogInfo("Database workflow details would be shown here")
	} else {
		// Show file-based workflow info
		fmt.Printf("Source: %s\n", fsWf.Path)
		fmt.Printf("Use Vault: %t\n", fsWf.UseVault)

		if fsWf.Description != nil {
			fmt.Printf("Description: %s\n", *fsWf.Description)
		}

		// Show pre-checks
		if len(fsWf.PreChecks) > 0 {
			fmt.Printf("\nPre-checks:\n")
			for i, check := range fsWf.PreChecks {
				fmt.Printf("  %d. %s\n", i+1, *check.Description)
				fmt.Printf("     Command: %s\n", check.Command)
			}
		}

		// Show steps
		if len(fsWf.Steps) > 0 {
			fmt.Printf("\nSteps:\n")
			for i, step := range fsWf.Steps {
				fmt.Printf("  %d. %s\n", i+1, *step.Description)
				fmt.Printf("     Command: %s\n", step.Command)
			}
		}

		// Show actions
		if len(fsWf.Actions) > 0 {
			fmt.Printf("\nActions:\n")
			for name, action := range fsWf.Actions {
				fmt.Printf("  %s: %s\n", name, *action.Description)
				fmt.Printf("     Command: %s\n", action.Command)
			}
		}

		// Show variables
		allContent := ""
		for _, step := range fsWf.Steps {
			allContent += step.Command + "\n"
		}
		for _, check := range fsWf.PreChecks {
			allContent += check.Command + "\n"
		}
		for name, action := range fsWf.Actions {
			allContent += fmt.Sprintf("%s: %s\n", name, action.Command)
		}

		variables := utils.ExtractTemplateVars(allContent)
		if len(variables) > 0 {
			fmt.Printf("\nRequired Variables: %s\n", strings.Join(variables, ", "))
		}
	}
}

func handleRunProjectPreChecks(cmd *cobra.Command) {
	// Look for migraine.yml or migraine.json in current directory
	projWf, err := workflow.LoadProjectWorkflow()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to load project workflow: %v", err))
		utils.LogInfo("Create a 'migraine.yml' or 'migraine.json' file in your current directory to use this feature")
		os.Exit(1)
	}

	// Validate workflow name
	if err := workflow.ValidateWorkflowName(projWf.Name); err != nil {
		utils.LogError(fmt.Sprintf("Invalid workflow name: %v", err))
		os.Exit(1)
	}

	// Check if the workflow has pre-checks
	if len(projWf.PreChecks) == 0 {
		utils.LogInfo("No pre-checks found in the project workflow")
		return
	}

	// Get storage service
	storage := sqlite.GetStorageService()

	// Upsert the workflow to the database
	if err := workflow.UpsertProjectWorkflowToDB(projWf, storage); err != nil {
		utils.LogError(fmt.Sprintf("Failed to upsert project workflow to database: %v", err))
		os.Exit(1)
	}

	utils.LogInfo(fmt.Sprintf("Project workflow '%s' loaded and upserted to database", projWf.Name))

	// Process variables from flags
	flagVars, err := cmd.Flags().GetStringArray("var")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get variables: %v", err))
		os.Exit(1)
	}

	variables := make(map[string]string)
	for _, v := range flagVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			utils.LogError(fmt.Sprintf("Invalid variable format: %s. Use KEY=VALUE format", v))
			os.Exit(1)
		}
		variables[parts[0]] = parts[1]
	}

	// Create variable resolver
	varResolver := workflow.NewVariableResolver(storage)

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(projWf.Name, projWf.UseVault, variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// Execute only the pre-checks of the project workflow
	executeProjectYAMLWorkflowPreChecksOnly(projWf, resolvedVars)
}

func handleRunWorkflowPreChecksFromStoredDirectory(workflowName string, cmd *cobra.Command) {
	storage := sqlite.GetStorageService()

	// Get the workflow from the database
	dbWf, err := storage.WorkflowStore().GetWorkflow(workflowName)
	if err != nil {
		utils.LogError(fmt.Sprintf("Workflow '%s' not found in database", workflowName))
		os.Exit(1)
	}

	// Get the stored working directory from the vault
	var workingDir string
	workingDirEntry, err := storage.VaultStore().GetVariableWithFallback("WORKING_DIR", dbWf.ID)
	if err == nil && workingDirEntry != nil {
		workingDir = workingDirEntry.Value
	} else {
		// If no stored directory found, use current directory
		utils.LogInfo(fmt.Sprintf("No stored working directory found for workflow '%s', using current directory", workflowName))
		workingDir, err = os.Getwd()
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to get current directory: %v", err))
			os.Exit(1)
		}
	}

	// Change to the working directory
	currentDir, err := os.Getwd()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get current directory: %v", err))
		os.Exit(1)
	}

	if err = os.Chdir(workingDir); err != nil {
		utils.LogError(fmt.Sprintf("Failed to change to directory '%s': %v", workingDir, err))
		os.Exit(1)
	}

	// Log the directory change
	utils.LogInfo(fmt.Sprintf("Changed to directory: %s", workingDir))

	// Parse the workflow metadata to get the YAML workflow
	var projectConfig workflow.ProjectConfig
	metadataBytes, err := json.Marshal(dbWf.Metadata)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to marshal workflow metadata: %v", err))
		// Restore original directory before exiting
		os.Chdir(currentDir)
		os.Exit(1)
	}

	if err := json.Unmarshal(metadataBytes, &projectConfig); err != nil {
		utils.LogError(fmt.Sprintf("Failed to unmarshal workflow metadata: %v", err))
		// Restore original directory before exiting
		os.Chdir(currentDir)
		os.Exit(1)
	}

	// Convert to YAMLWorkflow format
	yamlWf := &workflow.YAMLWorkflow{
		Name:        dbWf.Name,
		Description: projectConfig.Description,
		PreChecks:   projectConfig.PreChecks,
		Steps:       projectConfig.Steps,
		Actions:     projectConfig.Actions,
		Config:      projectConfig.Config,
		UseVault:    dbWf.UseVault,
		Path:        dbWf.Path,
	}

	// Check if the workflow has pre-checks
	if len(yamlWf.PreChecks) == 0 {
		utils.LogInfo("No pre-checks found in the workflow")
		// Restore original directory
		os.Chdir(currentDir)
		return
	}

	// Process variables from flags
	flagVars, err := cmd.Flags().GetStringArray("var")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get variables: %v", err))
		// Restore original directory before exiting
		os.Chdir(currentDir)
		os.Exit(1)
	}

	variables := make(map[string]string)
	for _, v := range flagVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			utils.LogError(fmt.Sprintf("Invalid variable format: %s. Use KEY=VALUE format", v))
			// Restore original directory before exiting
			os.Chdir(currentDir)
			os.Exit(1)
		}
		variables[parts[0]] = parts[1]
	}

	// Create variable resolver
	varResolver := workflow.NewVariableResolver(storage)

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(yamlWf.Name, yamlWf.UseVault, variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		// Restore original directory before exiting
		os.Chdir(currentDir)
		os.Exit(1)
	}

	// Execute only the pre-checks of the workflow
	executeProjectYAMLWorkflowPreChecksOnly(yamlWf, resolvedVars)

	// Restore original directory
	os.Chdir(currentDir)
	utils.LogInfo(fmt.Sprintf("Restored to original directory: %s", currentDir))
}

func executeProjectYAMLWorkflowPreChecksOnly(yamlWf *workflow.YAMLWorkflow, variables map[string]string) {
	// Display workflow header
	ui.WorkflowHeader(yamlWf.Name, "pre-checks")
	startTime := time.Now()

	// Create variable resolver for applying variables
	varResolver := workflow.NewVariableResolver(sqlite.GetStorageService())

	// Track precheck statistics
	precheckCount := 0
	prechecksPassed := 0
	prechecksFailed := 0
	prechecksWarn := 0

	// Run pre-checks section
	ui.SectionHeader("PRECHECKS")

	for i, check := range yamlWf.PreChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, variables)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		// Execute the command using the execution package
		precheckCount++
		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")
		}
	}

	// Display summary
	totalDuration := time.Since(startTime)
	ui.Summary("SUCCESS", totalDuration, prechecksPassed, prechecksFailed, prechecksWarn, 0, 0, "", "")

	ui.LogSuccessBordered(fmt.Sprintf("Project workflow pre-checks for '%s' completed successfully", yamlWf.Name))
}
