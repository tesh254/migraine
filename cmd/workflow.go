package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/utils"
)

var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wk"},
	Short:   "Manage workflow templates",
}

var workflowAddCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new workflow from a template",
	Run: func(cmd *cobra.Command, args []string) {
		handleAddWorkflow()
	},
}

var workflowListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all generate workflows",
	Run: func(cmd *cobra.Command, args []string) {
		handleListWorkflows()
	},
}

var workflowDeleteCmd = &cobra.Command{
	Use:     "delete [workflow_id]",
	Aliases: []string{"del"},
	Short:   "Delete a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		handleDeleteWorkflow(args[0])
	},
}

var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Template related commands",
	Aliases: []string{"tmpl"},
}

var templateNewCmd = &cobra.Command{
	Use:   "add [template_file]",
	Short: "Create a new workflow template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		templatePath := args[0]
		err := handleNewTemplate(templatePath)
		if err != nil {
			utils.LogError(fmt.Sprintf("%v", err))
		}
	},
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	Run: func(cmd *cobra.Command, args []string) {
		handleListTemplates()
	},
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete [template_name]",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDeleteTemplate(args[0])
	},
}

var templateLoadRemoteCmd = &cobra.Command{
	Use:   "load [url]",
	Short: "Load a workflow template from a remote URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleLoadRemoteTemplate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowAddCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowDeleteCmd)
	templateCmd.AddCommand(templateNewCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateLoadRemoteCmd)
	workflowCmd.AddCommand(templateCmd)
}
