package cmd

import (
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

func readLine() string {
	var buf [1]byte
	var line []byte
	for {
		n, err := os.Stdin.Read(buf[:])
		if n > 0 {
			if buf[0] == '\n' {
				break
			}
			if buf[0] != '\r' {
				line = append(line, buf[0])
			}
		}
		if err != nil {
			break
		}
	}
	return string(line)
}

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
		utils.LogError(fmt.Sprintf("Failed to list local workflows: %v", err))
		// Continue even if local discovery fails
	}

	if len(dbWorkflows) == 0 && len(fsWorkflows) == 0 {
		fmt.Println("No workflows found.")
		return
	}

	if len(dbWorkflows) > 0 {
		ui.SectionHeader("DATABASE WORKFLOWS")
		for _, wf := range dbWorkflows {
			// Extract description from metadata if possible
			desc := "No description"
			if wf.Metadata != nil {
				if d, ok := wf.Metadata["description"].(string); ok {
					desc = d
				}
			}

			fmt.Printf("  • %s - %s\n", wf.Name, desc)
			if wf.UseVault {
				fmt.Printf("    (uses vault)\n")
			}
			fmt.Println()
		}
	}

	if len(fsWorkflows) > 0 {
		ui.SectionHeader("LOCAL WORKFLOWS")
		for _, wf := range fsWorkflows {
			desc := "No description"
			if wf.Description != nil {
				desc = *wf.Description
			}
			fmt.Printf("  • %s - %s\n", wf.Name, desc)
			if wf.UseVault {
				fmt.Printf("    (uses vault)\n")
			}
			if wf.Path != "" {
				fmt.Printf("    Path: %s\n", wf.Path)
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
	var configVariables map[string]interface{}

	if dbErr == nil {
		workflowID = dbWf.ID
		// Try to extract config variables from metadata if possible
		var config workflow.ProjectConfig
		metadataBytes, _ := json.Marshal(dbWf.Metadata)
		if err := json.Unmarshal(metadataBytes, &config); err == nil {
			configVariables = config.Config.Variables
		}
	} else {
		workflowID = workflowName
		configVariables = fsWf.Config.Variables
	}

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(workflowID, useVault, variables, configVariables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// If there are still missing variables, prompt for them if not using vault
	if !useVault {
		requiredVars := utils.ExtractTemplateVars(workflowContent)

		for _, v := range requiredVars {
			if _, exists := resolvedVars[v]; !exists {
				fmt.Printf("%s: ", v)
				resolvedVars[v] = readLine()
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

			// Run on_fail hook if present
			if check.OnFail != "" {
				if hookErr := executeHook(check.OnFail, config.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")

			// Run on_success hook if present
			if check.OnSuccess != "" {
				if hookErr := executeHook(check.OnSuccess, config.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_success hook failed: %v", i+1, hookErr))
					os.Exit(1)
				}
			}
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

			// Run on_fail hook if present
			if step.OnFail != "" {
				if hookErr := executeHook(step.OnFail, config.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Step %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			os.Exit(1)
		}
		ui.LogInfoBordered("Step completed successfully")

		// Run on_success hook if present
		if step.OnSuccess != "" {
			if hookErr := executeHook(step.OnSuccess, config.Actions, variables, varResolver); hookErr != nil {
				ui.LogErrorBordered(fmt.Sprintf("Step %d on_success hook failed: %v", i+1, hookErr))
				os.Exit(1)
			}
		}
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

			// Run on_fail hook if present
			if check.OnFail != "" {
				if hookErr := executeHook(check.OnFail, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")

			// Run on_success hook if present
			if check.OnSuccess != "" {
				if hookErr := executeHook(check.OnSuccess, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_success hook failed: %v", i+1, hookErr))
					os.Exit(1)
				}
			}
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

			// Run on_fail hook if present
			if step.OnFail != "" {
				if hookErr := executeHook(step.OnFail, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Step %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			os.Exit(1)
		}
		utils.LogInfo("Step completed successfully")

		// Run on_success hook if present
		if step.OnSuccess != "" {
			if hookErr := executeHook(step.OnSuccess, yamlWf.Actions, variables, varResolver); hookErr != nil {
				ui.LogErrorBordered(fmt.Sprintf("Step %d on_success hook failed: %v", i+1, hookErr))
				os.Exit(1)
			}
		}
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

	// Determine workflow ID (for project workflow, use name as ID for variable resolution)
	workflowID := projWf.Name

	// Resolve variables based on workflow configuration
	resolvedVars, err := varResolver.ResolveVariables(workflowID, projWf.UseVault, variables, projWf.Config.Variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// If there are still missing variables, prompt for them if not using vault
	if !projWf.UseVault {
		var workflowContent string
		for _, step := range projWf.Steps {
			workflowContent += step.Command + "\n"
		}
		for _, check := range projWf.PreChecks {
			workflowContent += check.Command + "\n"
		}
		for _, action := range projWf.Actions {
			workflowContent += action.Command + "\n"
		}

		requiredVars := utils.ExtractTemplateVars(workflowContent)

		for _, v := range requiredVars {
			if _, exists := resolvedVars[v]; !exists {
				fmt.Printf("%s: ", v)
				resolvedVars[v] = readLine()
			}
		}
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

			// Run on_fail hook if present
			if check.OnFail != "" {
				if hookErr := executeHook(check.OnFail, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			ui.LogInfoBordered("Pre-check completed successfully")

			// Run on_success hook if present
			if check.OnSuccess != "" {
				if hookErr := executeHook(check.OnSuccess, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_success hook failed: %v", i+1, hookErr))
					os.Exit(1)
				}
			}
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

					// Run on_fail hook if present
					if action.OnFail != "" {
						if hookErr := executeHook(action.OnFail, yamlWf.Actions, variables, varResolver); hookErr != nil {
							ui.LogErrorBordered(fmt.Sprintf("Action '%s' on_fail hook failed: %v", actionName, hookErr))
						}
					}

					os.Exit(1)
				}
				ui.LogInfoBordered("Action completed successfully")

				// Run on_success hook if present
				if action.OnSuccess != "" {
					if hookErr := executeHook(action.OnSuccess, yamlWf.Actions, variables, varResolver); hookErr != nil {
						ui.LogErrorBordered(fmt.Sprintf("Action '%s' on_success hook failed: %v", actionName, hookErr))
						os.Exit(1)
					}
				}
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

			// Run on_fail hook if present
			if step.OnFail != "" {
				if hookErr := executeHook(step.OnFail, yamlWf.Actions, variables, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Step %d on_fail hook failed: %v", i+1, hookErr))
				}
			}

			os.Exit(1)
		}
		ui.LogInfoBordered("Step completed successfully")

		// Run on_success hook if present
		if step.OnSuccess != "" {
			if hookErr := executeHook(step.OnSuccess, yamlWf.Actions, variables, varResolver); hookErr != nil {
				ui.LogErrorBordered(fmt.Sprintf("Step %d on_success hook failed: %v", i+1, hookErr))
				os.Exit(1)
			}
		}
	}

	// Display summary
	totalDuration := time.Since(startTime)
	ui.Summary("SUCCESS", totalDuration, prechecksPassed, prechecksFailed, prechecksWarn, scriptCount, scriptCount, "", "")

	ui.LogSuccessBordered(fmt.Sprintf("Project workflow '%s' completed successfully", yamlWf.Name))
}

func executeHook(hook string, actions map[string]workflow.YAMLStep, variables map[string]string, varResolver *workflow.VariableResolver) error {
	if hook == "" {
		return nil
	}

	if strings.HasPrefix(hook, "action:") {
		actionName := strings.TrimPrefix(hook, "action:")
		action, ok := actions[actionName]
		if !ok {
			return fmt.Errorf("action '%s' not found", actionName)
		}

		ui.LogInfoBordered(fmt.Sprintf("Executing hook action: %s", actionName))

		command, err := varResolver.ApplyVariables(action.Command, variables)
		if err != nil {
			return fmt.Errorf("failed to apply variables to action %s: %v", actionName, err)
		}

		return execution.ExecuteCommand(command)
	} else if strings.HasPrefix(hook, "run:") {
		commandRaw := strings.TrimPrefix(hook, "run:")

		ui.LogInfoBordered(fmt.Sprintf("Executing hook command: %s", commandRaw))

		command, err := varResolver.ApplyVariables(commandRaw, variables)
		if err != nil {
			return fmt.Errorf("failed to apply variables to hook command: %v", err)
		}

		return execution.ExecuteCommand(command)
	}

	return fmt.Errorf("unknown hook format: %s (must start with 'action:' or 'run:')", hook)
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

	// Get storage service
	storage := sqlite.GetStorageService()

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

	// Resolve variables
	resolvedVars, err := varResolver.ResolveVariables(projWf.Name, projWf.UseVault, variables, projWf.Config.Variables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// If there are still missing variables, prompt for them if not using vault
	if !projWf.UseVault {
		workflowContent := ""
		for _, check := range projWf.PreChecks {
			workflowContent += check.Command + "\n"
		}
		requiredVars := utils.ExtractTemplateVars(workflowContent)
		for _, v := range requiredVars {
			if _, exists := resolvedVars[v]; !exists {
				fmt.Printf("%s: ", v)
				resolvedVars[v] = readLine()
			}
		}
	}

	// Run pre-checks
	ui.WorkflowHeader(projWf.Name, "pre-check")
	
	prechecksPassed := 0
	prechecksFailed := 0
	
	ui.SectionHeader("PRECHECKS")

	for i, check := range projWf.PreChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, resolvedVars)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			
			if check.OnFail != "" {
				if hookErr := executeHook(check.OnFail, projWf.Actions, resolvedVars, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_fail hook failed: %v", i+1, hookErr))
				}
			}
			
			prechecksFailed++
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			prechecksPassed++
			
			if check.OnSuccess != "" {
				if hookErr := executeHook(check.OnSuccess, projWf.Actions, resolvedVars, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_success hook failed: %v", i+1, hookErr))
					os.Exit(1)
				}
			}
		}
	}

	ui.LogSuccessBordered("All pre-checks passed successfully")
}

func handleRunWorkflowPreChecksFromStoredDirectory(workflowName string, cmd *cobra.Command) {
	// First, try to find the workflow in the database
	storage := sqlite.GetStorageService()
	dbWf, dbErr := storage.WorkflowStore().GetWorkflow(workflowName)

	// Also try to find the workflow in current directory
	fsWf, fsErr := workflow.FindWorkflowByName(workflowName)

	if dbErr != nil && fsErr != nil {
		utils.LogError(fmt.Sprintf("Workflow '%s' not found in database or current directory", workflowName))
		os.Exit(1)
	}

	// Prefer database workflow
	var useVault bool
	var configVariables map[string]interface{}
	var workflowID string
	var preChecks []workflow.YAMLStep
	var actions map[string]workflow.YAMLStep

	if dbErr == nil {
		useVault = dbWf.UseVault
		workflowID = dbWf.ID
		
		metadataBytes, _ := json.Marshal(dbWf.Metadata)
		var config workflow.ProjectConfig
		if err := json.Unmarshal(metadataBytes, &config); err == nil {
			configVariables = config.Config.Variables
			preChecks = config.PreChecks
			actions = config.Actions
		}
	} else {
		useVault = fsWf.UseVault
		workflowID = workflowName
		configVariables = fsWf.Config.Variables
		preChecks = fsWf.PreChecks
		actions = fsWf.Actions
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

	// Resolve variables
	resolvedVars, err := varResolver.ResolveVariables(workflowID, useVault, variables, configVariables)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to resolve variables: %v", err))
		os.Exit(1)
	}

	// If missing variables, prompt
	if !useVault {
		workflowContent := ""
		for _, check := range preChecks {
			workflowContent += check.Command + "\n"
		}
		requiredVars := utils.ExtractTemplateVars(workflowContent)
		for _, v := range requiredVars {
			if _, exists := resolvedVars[v]; !exists {
				fmt.Printf("%s: ", v)
				resolvedVars[v] = readLine()
			}
		}
	}

	// Run pre-checks
	ui.WorkflowHeader(workflowName, "pre-check")
	
	ui.SectionHeader("PRECHECKS")

	for i, check := range preChecks {
		precheckStartTime := time.Now()
		command, err := varResolver.ApplyVariables(check.Command, resolvedVars)
		if err != nil {
			ui.LogErrorBordered(fmt.Sprintf("Failed to apply variables to pre-check %d: %v", i+1, err))
			os.Exit(1)
		}

		err = execution.ExecuteCommand(command)
		duration := time.Since(precheckStartTime)

		if err != nil {
			ui.PrecheckResult(*check.Description, "fail", duration, "")
			ui.LogErrorBordered(fmt.Sprintf("Pre-check %d failed: %v", i+1, err))
			
			if check.OnFail != "" {
				if hookErr := executeHook(check.OnFail, actions, resolvedVars, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_fail hook failed: %v", i+1, hookErr))
				}
			}
			
			os.Exit(1)
		} else {
			ui.PrecheckResult(*check.Description, "ok", duration, "")
			
			if check.OnSuccess != "" {
				if hookErr := executeHook(check.OnSuccess, actions, resolvedVars, varResolver); hookErr != nil {
					ui.LogErrorBordered(fmt.Sprintf("Pre-check %d on_success hook failed: %v", i+1, hookErr))
					os.Exit(1)
				}
			}
		}
	}

	ui.LogSuccessBordered("All pre-checks passed successfully")
}

func handleWorkflowInfoV2(workflowName string) {
	// First, try to find the workflow in the database
	storage := sqlite.GetStorageService()
	dbWf, dbErr := storage.WorkflowStore().GetWorkflow(workflowName)
	
	// Also try to find the workflow in current directory
	fsWf, fsErr := workflow.FindWorkflowByName(workflowName)
	
	if dbErr != nil && fsErr != nil {
		utils.LogError(fmt.Sprintf("Workflow '%s' not found", workflowName))
		os.Exit(1)
	}
	
	ui.SectionHeader(fmt.Sprintf("WORKFLOW INFO: %s", workflowName))
	
	if dbErr == nil {
		fmt.Printf("Source: Database\n")
		fmt.Printf("ID: %s\n", dbWf.ID)
		fmt.Printf("Vault Enabled: %v\n", dbWf.UseVault)
		
		metadataBytes, _ := json.Marshal(dbWf.Metadata)
		var config workflow.ProjectConfig
		if err := json.Unmarshal(metadataBytes, &config); err == nil {
			fmt.Printf("Pre-checks: %d\n", len(config.PreChecks))
			fmt.Printf("Steps: %d\n", len(config.Steps))
			fmt.Printf("Actions: %d\n", len(config.Actions))
		}
	} else {
		fmt.Printf("Source: Local File\n")
		if fsWf.Path != "" {
			fmt.Printf("Path: %s\n", fsWf.Path)
		}
		if fsWf.Description != nil {
			fmt.Printf("Description: %s\n", *fsWf.Description)
		}
		fmt.Printf("Vault Enabled: %v\n", fsWf.UseVault)
		fmt.Printf("Pre-checks: %d\n", len(fsWf.PreChecks))
		fmt.Printf("Steps: %d\n", len(fsWf.Steps))
		fmt.Printf("Actions: %d\n", len(fsWf.Actions))
	}
}
