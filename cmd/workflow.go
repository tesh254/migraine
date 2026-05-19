package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/editor"
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
Use with a name to create a regular workflow in the workflows/ directory.
Use with --editor to configure editor LSP integration instead.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		editorFlag, _ := cmd.Flags().GetString("editor")
		if editorFlag != "" {
			return runEditorSetup(cmd, editorFlag)
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		generateYML, _ := cmd.Flags().GetBool("yml")
		generateJSON, _ := cmd.Flags().GetBool("json")

		if generateYML || generateJSON {
			format := "yaml"
			if generateJSON {
				format = "json"
			}
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig(format, description)
		} else if name == "" {
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig("yaml", description)
		} else {
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
	workflowInitCmd.Flags().StringP("editor", "e", "", "Configure editor LSP integration (vscode, neovim, vim, helix)")
	workflowPreChecksCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
}

// Create a top-level init command as an alias to workflow init
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new workflow file or configure editor integration",
	Long: `Create a new workflow file with commented sections.
	
Use without arguments to create a project configuration file (migraine.yml).
Use with --yml or --json flags to create project config files directly.
Use with a name to create a regular workflow in the workflows/ directory.
Use with --editor to configure editor LSP integration (vscode, neovim, vim, helix).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		editorFlag, _ := cmd.Flags().GetString("editor")
		if editorFlag != "" {
			return runEditorSetup(cmd, editorFlag)
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		generateYML, _ := cmd.Flags().GetBool("yml")
		generateJSON, _ := cmd.Flags().GetBool("json")

		if generateYML || generateJSON {
			format := "yaml"
			if generateJSON {
				format = "json"
			}
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig(format, description)
		} else if name == "" {
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = "Project-level workflow configuration"
			}
			return workflow.ScaffoldProjectConfig("yaml", description)
		} else {
			description, _ := cmd.Flags().GetString("description")
			if description == "" {
				description = name
			}
			return workflow.ScaffoldYAMLWorkflow(name, description)
		}
	},
}

func runEditorSetup(cmd *cobra.Command, editorFlag string) error {
	if editorFlag == "auto" || editorFlag == "detect" {
		detected := editor.Detect()
		if len(detected) == 0 {
			fmt.Println("No supported editors detected.")
			fmt.Printf("Supported editors: %s\n", strings.Join(editor.Supported(), ", "))
			fmt.Println("Use --editor <name> to configure a specific editor.")
			return nil
		}
		fmt.Printf("Detected editors: %s\n\n", strings.Join(detected, ", "))
		for _, e := range detected {
			fmt.Printf("Configuring %s...\n", e)
			if err := editor.Setup(e, ""); err != nil {
				fmt.Printf("  ✗ %v\n", err)
			} else {
				fmt.Printf("  ✓ %s configured\n", e)
			}
		}
		return nil
	}

	fmt.Printf("Configuring %s...\n", editorFlag)
	if err := editor.Setup(editorFlag, ""); err != nil {
		return err
	}
	fmt.Printf("  ✓ %s configured\n", editorFlag)
	fmt.Println("\nDone! Open a .mg file in your editor to get started.")
	return nil
}

func init() {
	// Add flags to the root init command
	initCmd.Flags().StringP("description", "d", "", "Description for the workflow")
	initCmd.Flags().Bool("yml", false, "Generate project configuration file as migraine.yml")
	initCmd.Flags().Bool("json", false, "Generate project configuration file as migraine.json")
	initCmd.Flags().StringP("editor", "e", "", "Configure editor LSP (vscode, neovim, vim, helix, or 'auto' to detect)")

	// Add flags to workflow run command
	workflowRunCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
	workflowRunCmd.Flags().StringArrayP("action", "a", []string{}, "Action to run")

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowInitCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowValidateCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowCmd.AddCommand(workflowPreChecksCmd)
	workflowCmd.AddCommand(workflowInfoCmd)
}
