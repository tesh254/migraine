package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/constants"
)

var manCmd = &cobra.Command{
	Use:     "man",
	Aliases: []string{"manual"},
	Short:   "Display manual page for migraine",
	Long: `Display manual page for migraine.

This command shows the manual page for the migraine CLI tool,
providing comprehensive usage information and documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		manpage := generateManPage()
		fmt.Print(manpage)
	},
}

var generateManPageCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate man page file",
	Long: `Generate the man page file to the specified directory.

This command generates the man page file (migraine.1) to the specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			outputDir = "." // Default to current directory
		}

		manpage := generateManPageAsRoff()
		manpagePath := filepath.Join(outputDir, "migraine.1")

		err := os.WriteFile(manpagePath, []byte(manpage), 0644)
		if err != nil {
			fmt.Printf("Error writing man page: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Man page generated successfully at: %s\n", manpagePath)
	},
}

// generateManPage creates and returns the manual page content in Markdown format
func generateManPage() string {
	manpage := fmt.Sprintf(`%% MIGRAINE(1) Version %s | Migraine Manual
%% Tesh Teckie
%% October 2025

# NAME
migraine, mgr - A CLI tool used to organize and automate complex workflows with templated commands

# SYNOPSIS
**migraine** [COMMAND] [FLAGS]

**mgr** [COMMAND] [FLAGS]

# DESCRIPTION
Migraine is a robust CLI tool used to organize and automate complex workflows with templated commands. 
Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution, 
pre-flight checks, and discrete actions.

# COMMANDS
**help** [command]
    Help about any command

**run** [workflow-name]
    Execute a workflow by name

**vars** [subcommand]
    Manage vault variables with scope flags
    - vars list
    - vars get [key]
    - vars set [key] [value]
    - vars delete [key]

**workflow** [subcommand]
    Manage workflows
    - workflow list
    - workflow get [name]
    - workflow create [name]
    - workflow delete [name]
    - workflow info [name]

**version** [flags]
    Show version information
    - --json: Output version information in JSON format
    - -s, --short: Output short version only
    - -c, --commit: Output version with commit hash

**buildinfo**
    Show detailed build information

**man**
    Display this manual page

# OPTIONS
-v, --version
    Print detailed version information

# EXAMPLES
    # Run a workflow
    $ migraine run my-workflow

    # List all workflows
    $ migraine workflow list

    # Set a variable
    $ migraine vars set API_KEY mysecretkey

    # Get version information
    $ migraine version

# AUTHOR
Erick Wachira <email@bywachira.com>

# COPYRIGHT
MIT License - Copyright (c) 2024-2025 Erick Wachira

`,
		constants.VERSION())

	return manpage
}

// generateManPageAsRoff creates and returns the manual page content in ROFF format (standard man page format)
func generateManPageAsRoff() string {
	manpage := fmt.Sprintf(`.TH MIGRAINE 1 "October 2025" "Version %s" "Migraine Manual"
.SH NAME
migraine, mgr \\- A CLI tool used to organize and automate complex workflows with templated commands
.SH SYNOPSIS
.B migraine
.RB [COMMAND]
.RB [FLAGS]
.PP
.B mgr
.RB [COMMAND]
.RB [FLAGS]
.SH DESCRIPTION
Migraine is a robust CLI tool used to organize and automate complex workflows with templated commands.
Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution,
pre-flight checks, and discrete actions.
.SH COMMANDS
.TP
.B help [command]
Help about any command
.TP
.B run [workflow-name]
Execute a workflow by name
.TP
.B vars [subcommand]
Manage vault variables with scope flags
.RS
.TP
.B vars list
List all variables
.TP
.B vars get [key]
Get a variable by key
.TP
.B vars set [key] [value]
Set a variable with key and value
.TP
.B vars delete [key]
Delete a variable by key
.RE
.TP
.B workflow [subcommand]
Manage workflows
.RS
.TP
.B workflow list
List all workflows
.TP
.B workflow get [name]
Get a workflow by name
.TP
.B workflow create [name]
Create a new workflow
.TP
.B workflow delete [name]
Delete a workflow by name
.TP
.B workflow info [name]
Show detailed information about a workflow
.RE
.TP
.B version [flags]
Show version information
.RS
.TP
.B \\-\\-json
Output version information in JSON format
.TP
.B \\-s, \\-\\-short
Output short version only
.TP
.B \\-c, \\-\\-commit
Output version with commit hash
.RE
.TP
.B buildinfo
Show detailed build information
.TP
.B man
Display this manual page
.SH OPTIONS
.TP
.B \\-v, \\-\\-version
Print detailed version information
.SH EXAMPLES
.TP
.B $ migraine run my-workflow
Run a workflow
.TP
.B $ migraine workflow list
List all workflows
.TP
.B $ migraine vars set API_KEY mysecretkey
Set a variable
.TP
.B $ migraine version
Get version information
.SH AUTHOR
Erick Wachira <email@bywachira.com>
.SH COPYRIGHT
MIT License \\- Copyright (c) 2024\\-2025 Erick Wachira
`,
		constants.VERSION())

	return manpage
}

func init() {
	manCmd.AddCommand(generateManPageCmd)
	generateManPageCmd.Flags().StringP("output", "o", "", "Output directory for the man page file (default is current directory)")
	rootCmd.AddCommand(manCmd)
}