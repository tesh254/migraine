package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/constants"
)

var rootCmd = &cobra.Command{
	Use:     "mg",
	Short:   "Migraine - A CLI for managing personal and server workflows",
	Version: constants.VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(constants.MIGRAINE_ASCII_V2)
		fmt.Println(constants.CurrentOSWithVersion())
		fmt.Print(constants.MIGRAINE_USAGE)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
