package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/lsp"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the Migraine Language Server Protocol server",
	Long: `Start the Migraine LSP server for editor integration.

The LSP server provides:
  - Syntax highlighting via semantic tokens
  - Diagnostics (parse errors)
  - Autocompletion for keywords, blocks, and properties
  - Hover documentation
  - Document symbols (outline)

It communicates over stdio using the Language Server Protocol.
Use with a VS Code extension or other LSP-compatible editor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.ErrOrStderr(), "Starting Migraine LSP server on stdio...")
		server := lsp.NewServer()
		return lsp.RunStdioServer(server)
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)
}