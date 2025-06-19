package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/constants"
	"github.com/tesh254/migraine/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "migraine",
	Short:   "Migraine - A CLI tool used to organize and automate complex workflows with templated commands. Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.",
	Version: constants.VERSION(),
	Aliases: []string{"mgr"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle version flag specially to show detailed info
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Println(constants.DETAILED_VERSION())
			return nil
		}

		if cmd.Flags().NFlag() == 0 && len(args) == 0 {
			fmt.Print(constants.MIGRAINE_ASCII_V2)
			fmt.Println(constants.CurrentOSWithVersion())
			fmt.Printf("\n%s\n", constants.GetReleaseInfo())
		}
		return nil
	},
}

// Version command with multiple output formats
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Show version information for migraine.

This command displays version information extracted automatically from 
the Go build system, including Git commit, build date, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFlag, _ := cmd.Flags().GetBool("json")
		shortFlag, _ := cmd.Flags().GetBool("short")
		commitFlag, _ := cmd.Flags().GetBool("commit")

		switch {
		case jsonFlag:
			fmt.Println(version.GetJSONVersion())
		case shortFlag:
			fmt.Println(version.GetShortVersion())
		case commitFlag:
			fmt.Println(version.GetVersionWithCommit())
		default:
			fmt.Println(version.GetDetailedVersion())

			// Add extra info for development builds
			if version.IsDevelopment() {
				fmt.Printf("\n%sNote:%s This is a development build.\n",
					"\033[33m", "\033[0m")
			}
		}
	},
}

// Build info command for detailed build information
var buildInfoCmd = &cobra.Command{
	Use:   "buildinfo",
	Short: "Show detailed build information",
	Long:  `Show comprehensive build information including module details, VCS info, and build settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()

		fmt.Printf("Build Information:\n")
		fmt.Printf("==================\n")
		fmt.Printf("Version:      %s\n", info.Version)
		fmt.Printf("Git Commit:   %s\n", info.GitCommit)
		if info.GitTag != "unknown" {
			fmt.Printf("Git Tag:      %s\n", info.GitTag)
		}
		fmt.Printf("Build Date:   %s\n", info.BuildDate)
		fmt.Printf("Go Version:   %s\n", info.GoVersion)
		fmt.Printf("Platform:     %s\n", info.Platform)
		fmt.Printf("Compiler:     %s\n", info.Compiler)
		fmt.Printf("Modified:     %t\n", info.IsModified)
		if info.ModulePath != "" {
			fmt.Printf("Module Path:  %s\n", info.ModulePath)
		}
		if info.ModuleSum != "" {
			fmt.Printf("Module Sum:   %s\n", info.ModuleSum)
		}

		// Show build type
		fmt.Printf("\nBuild Type:   ")
		if version.IsRelease() {
			fmt.Printf("%sRelease%s\n", "\033[32m", "\033[0m")
		} else {
			fmt.Printf("%sDevelopment%s\n", "\033[33m", "\033[0m")
		}
	},
}

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(constants.VERSION())); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Root command flags
	rootCmd.Flags().BoolP("version", "v", false, "Print detailed version information")

	// Version command flags
	versionCmd.Flags().Bool("json", false, "Output version information in JSON format")
	versionCmd.Flags().BoolP("short", "s", false, "Output short version only")
	versionCmd.Flags().BoolP("commit", "c", false, "Output version with commit hash")
	rootCmd.AddCommand(buildInfoCmd)
	rootCmd.AddCommand(versionCmd)
}
