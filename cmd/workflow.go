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
	Use:     "add",
	Aliases: []string{"-a"},
	Short:   "Add a new workflow",
	Run: func(cmd *cobra.Command, args []string) {
		handleAddWorkflow()
	},
}

var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Template related commands",
	Aliases: []string{"tmpl"},
}

var templateNewCmd = &cobra.Command{
	Use:   "new [template_file]",
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

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowAddCmd)
	templateCmd.AddCommand(templateNewCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateListCmd)
	workflowCmd.AddCommand(templateCmd)
}
