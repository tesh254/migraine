package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/workflow"
)

var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wk"},
	Short:   "Manage workflows",
}

var workflowInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new workflow file with commented sections",
	Long: `Create a new workflow file with commented sections.
	
Use without arguments to create a project configuration file (migraine.yml).
Use with --yml or --json flags to create project config files directly.
Use with a name to create a regular workflow in the workflows/ directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		// Check if --yml or --json flags are used
		generateYML, _ := cmd.Flags().GetBool("yml")
		generateJSON, _ := cmd.Flags().GetBool("json")

		if generateYML || generateJSON {
			// Generate project configuration file
			format := "yaml"
			if generateJSON {
				format = "json"
			}

			// Get description from flag or use default
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}

			return workflow.ScaffoldProjectConfig(format, description)
		} else if name == "" {
			// No name provided and no format flags, generate default YAML config
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig("yaml", description)
		} else {
			// Regular workflow with name, generate in workflows/ directory
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = name
			}

			return workflow.ScaffoldYAMLWorkflow(name, description)
		}
	},
}

var workflowListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all workflows (from database and current directory)",
	Run: func(cmd *cobra.Command, args []string) {
		handleListWorkflows()
	},
}

var workflowValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a workflow file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		workflow, err := workflow.LoadYAMLWorkflow(path)
		if err != nil {
			return fmt.Errorf("failed to load workflow: %v", err)
		}

		// Basic validation
		if workflow.Name == "" {
			return fmt.Errorf("workflow name is required")
		}

		fmt.Printf("✓ Workflow '%s' is valid\n", workflow.Name)
		return nil
	},
}

var workflowPreChecksCmd = &cobra.Command{
	Use:   "pre-checks [name]",
	Short: "Run pre-checks from migraine.yaml in current directory or by workflow name",
	Long:  "Run only the pre-checks section of the migraine.yaml file in the current directory, or for a specific workflow by name",
	Args:  cobra.MaximumNArgs(1), // Allow zero args to run project config, or one arg for specific workflow
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No workflow name provided, run project pre-checks from current directory
			handleRunProjectPreChecks(cmd)
		} else {
			// Workflow name provided, run pre-checks for specific workflow from its stored directory
			workflowName := args[0]
			handleRunWorkflowPreChecksFromStoredDirectory(workflowName, cmd)
		}
	},
}

var workflowRunCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Run a workflow (new command for v2)",
	Args:  cobra.MaximumNArgs(1), // Allow zero args to run project config
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No workflow name provided, look for migraine.yml/migraine.json in current directory
			handleRunProjectWorkflow(cmd)
		} else {
			// Workflow name provided, run the existing logic
			handleRunWorkflowV2(args[0], cmd)
		}
	},
}

var workflowInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Display detailed information about a workflow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowInfoV2(args[0])
	},
}

// Add flags to the workflow init command
func init() {
	workflowInitCmd.Flags().StringP("description", "d", "", "Description for the workflow")
	workflowInitCmd.Flags().Bool("yml", false, "Generate project configuration file as migraine.yml")
	workflowInitCmd.Flags().Bool("json", false, "Generate project configuration file as migraine.json")
	workflowPreChecksCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
}

// Create a top-level init command as an alias to workflow init
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new workflow file with commented sections (alias for workflow init)",
	Long: `Create a new workflow file with commented sections.
	
Use without arguments to create a project configuration file (migraine.yml).
Use with --yml or --json flags to create project config files directly.
Use with a name to create a regular workflow in the workflows/ directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call the same logic as workflow init
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		// Check if --yml or --json flags are used
		generateYML, _ := cmd.Flags().GetBool("yml")
		generateJSON, _ := cmd.Flags().GetBool("json")

		if generateYML || generateJSON {
			// Generate project configuration file
			format := "yaml"
			if generateJSON {
				format = "json"
			}

			// Get description from flag or use default
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}

			return workflow.ScaffoldProjectConfig(format, description)
		} else if name == "" {
			// No name provided and no format flags, generate default YAML config
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig("yaml", description)
		} else {
			// Regular workflow with name, generate in workflows/ directory
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = name
			}

			return workflow.ScaffoldYAMLWorkflow(name, description)
		}
	},
}

func init() {
	// Add flags to the root init command
	initCmd.Flags().StringP("description", "d", "", "Description for the workflow")
	initCmd.Flags().Bool("yml", false, "Generate project configuration file as migraine.yml")
	initCmd.Flags().Bool("json", false, "Generate project configuration file as migraine.json")

	// Add flags to workflow run command
	workflowRunCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
	workflowRunCmd.Flags().StringArrayP("action", "a", []string{}, "Action to run")

	// Add commands
	rootCmd.AddCommand(initCmd) // Add top-level init command
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowInitCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowValidateCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowCmd.AddCommand(workflowPreChecksCmd)
	workflowCmd.AddCommand(workflowInfoCmd)
}
