package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use: "run [workflow_id]",
	Short: "Run a workflow",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
