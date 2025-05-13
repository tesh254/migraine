package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tesh254/migraine/internal/mcp"
	"github.com/tesh254/migraine/pkg/utils"
)

func HandleStartMCPServer(configPath string) {
	// Read and parse the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to read config file: %v", err))
		return
	}

	var config mcp.ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		utils.LogError(fmt.Sprintf("Failed to parse config file: %v", err))
		return
	}

	// Start the server based on the type
	utils.LogInfo(fmt.Sprintf("Starting %s MCP server...", config.Name))

	switch config.Type {
	case "filesystem":
		// Create filesystem server
		fsServer, err := mcp.NewFileSystemServer(config)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to create filesystem server: %v", err))
			return
		}

		// MCP servers only support stdio transport
		// Serve the filesystem server
		if err := mcp.ServeWithTransport(fsServer, "stdio", 0); err != nil {
			utils.LogError(fmt.Sprintf("Failed to serve filesystem server: %v", err))
			return
		}

	default:
		utils.LogError(fmt.Sprintf("Unsupported server type: %s", config.Type))
		return
	}
}
