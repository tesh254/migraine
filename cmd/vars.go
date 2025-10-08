package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/storage/sqlite"
	"github.com/tesh254/migraine/pkg/utils"
)

var varsCmd = &cobra.Command{
	Use:     "vars",
	Aliases: []string{"var"},
	Short:   "Manage vault variables",
}

var varsGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a variable value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		
		scope, _ := cmd.Flags().GetString("scope")
		workflowID, _ := cmd.Flags().GetString("workflow")
		
		var workflowIDPtr *string
		if workflowID != "" {
			workflowIDPtr = &workflowID
		}
		
		storage := sqlite.GetStorageService()
		
		if scope != "" && scope != "global" {
			// Get variable with specific scope
			entry, err := storage.VaultStore().GetVariable(key, scope, workflowIDPtr)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to get variable: %v", err))
				return
			}
			
			fmt.Printf("%s\n", entry.Value)
		} else {
			// Use fallback logic: workflow -> project -> global
			var workflowIDStr string
			if workflowID != "" {
				workflowIDStr = workflowID
			}
			
			entry, err := storage.VaultStore().GetVariableWithFallback(key, workflowIDStr)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to get variable with fallback: %v", err))
				return
			}
			
			fmt.Printf("%s\n", entry.Value)
		}
	},
}

var varsSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a variable value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		
		scope, _ := cmd.Flags().GetString("scope")
		workflowID, _ := cmd.Flags().GetString("workflow")
		
		var workflowIDPtr *string
		if workflowID != "" {
			workflowIDPtr = &workflowID
		}
		
		storage := sqlite.GetStorageService()
		
		// Check if variable already exists
		_, err := storage.VaultStore().GetVariable(key, scope, workflowIDPtr)
		if err == nil {
			// Update existing variable
			err = storage.VaultStore().UpdateVariable(key, scope, workflowIDPtr, value)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to update variable: %v", err))
				return
			}
			utils.LogSuccess(fmt.Sprintf("Variable '%s' updated successfully", key))
		} else {
			// Create new variable
			entry := sqlite.VaultEntry{
				Key:        key,
				Value:      value,
				Scope:      scope,
				WorkflowID: workflowIDPtr,
			}
			
			err = storage.VaultStore().CreateVariable(entry)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to create variable: %v", err))
				return
			}
			utils.LogSuccess(fmt.Sprintf("Variable '%s' created successfully", key))
		}
	},
}

var varsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all variables",
	Run: func(cmd *cobra.Command, args []string) {
		scope, _ := cmd.Flags().GetString("scope")
		workflowID, _ := cmd.Flags().GetString("workflow")
		
		var workflowIDPtr *string
		if workflowID != "" {
			workflowIDPtr = &workflowID
		}
		
		storage := sqlite.GetStorageService()
		
		variables, err := storage.VaultStore().ListVariables(scope, workflowIDPtr)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to list variables: %v", err))
			return
		}
		
		if len(variables) == 0 {
			fmt.Println("No variables found")
			return
		}
		
		fmt.Printf("\nVariables:\n")
		for _, variable := range variables {
			scopeInfo := variable.Scope
			if variable.WorkflowID != nil {
				scopeInfo = fmt.Sprintf("%s:%s", variable.Scope, *variable.WorkflowID)
			}
			
			fmt.Printf("  %s (%s)\n", variable.Key, scopeInfo)
		}
	},
}

var varsDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a variable",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		
		scope, _ := cmd.Flags().GetString("scope")
		workflowID, _ := cmd.Flags().GetString("workflow")
		
		var workflowIDPtr *string
		if workflowID != "" {
			workflowIDPtr = &workflowID
		}
		
		storage := sqlite.GetStorageService()
		
		err := storage.VaultStore().DeleteVariable(key, scope, workflowIDPtr)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to delete variable: %v", err))
			return
		}
		
		utils.LogSuccess(fmt.Sprintf("Variable '%s' deleted successfully", key))
	},
}

func init() {
	// Add flags to all commands
	scopeFlag := "global"
	workflowFlag := ""
	
	varsGetCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "global", "Variable scope (global, project, workflow)")
	varsGetCmd.Flags().StringVarP(&workflowFlag, "workflow", "w", "", "Workflow ID (for workflow scope)")
	
	varsSetCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "global", "Variable scope (global, project, workflow)")
	varsSetCmd.Flags().StringVarP(&workflowFlag, "workflow", "w", "", "Workflow ID (for workflow scope)")
	
	varsListCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "", "Variable scope (global, project, workflow)")
	varsListCmd.Flags().StringVarP(&workflowFlag, "workflow", "w", "", "Workflow ID (for workflow scope)")
	
	varsDeleteCmd.Flags().StringVarP(&scopeFlag, "scope", "s", "global", "Variable scope (global, project, workflow)")
	varsDeleteCmd.Flags().StringVarP(&workflowFlag, "workflow", "w", "", "Workflow ID (for workflow scope)")
	
	// Add commands to root
	rootCmd.AddCommand(varsCmd)
	varsCmd.AddCommand(varsGetCmd)
	varsCmd.AddCommand(varsSetCmd)
	varsCmd.AddCommand(varsListCmd)
	varsCmd.AddCommand(varsDeleteCmd)
}