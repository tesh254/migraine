package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [workflow_name]",
	Short: "Run a workflow (new v2 command)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No workflow name provided, look for migraine.yml/migraine.json
			handleRunProjectWorkflow(cmd)
		} else {
			// Workflow name provided, run the existing logic
			handleRunWorkflowV2(args[0], cmd)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
	runCmd.Flags().StringArrayP("action", "a", []string{}, "Action to run")
}
