package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/handlers"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage Model Context Protocol (MCP) servers",
	Long:  `Create and manage Model Context Protocol (MCP) servers for AI model integration.`,
}

var mcpStart = &cobra.Command{
	Use:   "start [config]",
	Short: "Start an MCP server",
	Long:  `Start an MCP server using the provided configuration file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handlers.HandleStartMCPServer(args[0])
	},
}

// var mcpListCmd = &cobra.Command{
// 	Use:   "list",
// 	Short: "List all MCP servers",
// 	Long:  `List all running MCP servers.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		handlers.HandleListMCPServers()
// 	},
// }

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpStart)
	// mcpCmd.AddCommand(mcpListCmd)
}

// var mcpRegisterCmd = &cobra.Command{
//     Use:   "register [client] [server-name]",
//     Short: "Register an MCP server with a client",
//     Args:  cobra.ExactArgs(2),
//     Run: func(cmd *cobra.Command, args []string) {
//         configFile, _ := cmd.Flags().GetString("config")
//         handleRegisterMCPServer(args[0], args[1], configFile)
//     },
// }
