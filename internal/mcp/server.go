package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

// ServeWithTransport starts an MCP server with the specified transport
func ServeWithTransport(mcpServer *server.MCPServer, transportType string, port int) error {
	switch transportType {
	case "stdio":
		// For stdio transport, we don't need the port
		return server.ServeStdio(mcpServer)
	case "http":
		return fmt.Errorf("HTTP transport is not supported - MCP servers use stdio communication")
	default:
		return fmt.Errorf("unsupported transport: %s. Supported transports: stdio", transportType)
	}
}
