package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [workflow_id]",
	Short: "Run a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		handleRunWorkflow(args[0], cmd)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayP("var", "v", []string{}, "Variables in KEY=VALUE format")
	runCmd.Flags().StringArrayP("action", "a", []string{}, "Action to run")
}
