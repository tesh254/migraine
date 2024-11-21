package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/constants"
)

var rootCmd = &cobra.Command{
	Use:     "migraine",
	Short:   "Migraine - A CLI tool used to organize and automate complex workflows with templated commands. Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.",
	Version: constants.VERSION,
	Aliases: []string{"mig"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().NFlag() == 0 && len(args) == 0 {
			fmt.Print(constants.MIGRAINE_ASCII_V2)
			fmt.Println(constants.CurrentOSWithVersion())
		}
		return nil
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
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
}
