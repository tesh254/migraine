package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewFileSystemServer(config ServerConfig) (*server.MCPServer, error) {
	var options FileSystemOptions

	s := server.NewMCPServer(
		config.Name,
		config.Version,
		server.WithToolCapabilities(true),
	)

	registerFileSystemTools(s, options)

	return s, nil
}

func registerFileSystemTools(s *server.MCPServer, options FileSystemOptions) {
	readFileTool := mcp.NewTool(
		"read_file",
		mcp.WithDescription("Read the contents of a file"),
		mcp.WithString("path", mcp.Required()),
	)

	listAllowedContentsTool := mcp.NewTool(
		"list_allowed_contents",
		mcp.WithDescription("List all files and directories within allowed directories"),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to recursively list subdirectories"),
			mcp.DefaultBool(true),
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Maximum depth for recursive listing (0 = unlimited)"),
			mcp.DefaultNumber(0),
		),
	)

	writeFileTool := mcp.NewTool(
		"write_file",
		mcp.WithDescription("Write or update the contents of a file"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to write or update"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to write to the file"),
		),
		mcp.WithBoolean("create_if_missing",
			mcp.Description("Create the file if it doesn't exist"),
			mcp.DefaultBool(true),
		),
	)

	s.AddTool(readFileTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, ok := req.Params.Arguments["path"].(string)

		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		if !isPathAllowed(path, options.AllowedDirs) {
			return nil, fmt.Errorf("access to path is not allowed: %s", path)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}

		return mcp.NewToolResultText(string(content)), nil
	})

	s.AddTool(listAllowedContentsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		recursive := false
		if val, ok := req.Params.Arguments["recursive"].(bool); ok {
			recursive = val
		}

		maxDepth := 0
		if val, ok := req.Params.Arguments["max_depth"].(float64); ok {
			maxDepth = int(val)
		}

		var resultBuilder strings.Builder
		resultBuilder.WriteString("Files and directories in allowed paths:\n\n")

		for _, dir := range options.AllowedDirs {
			resultBuilder.WriteString(fmt.Sprintf("Directory: %s\n", dir))
			resultBuilder.WriteString("------------------------\n")

			entries, err := listDirectory(dir, recursive, maxDepth, 0)
			if err != nil {
				resultBuilder.WriteString(fmt.Sprintf("Error reading directory: %v\n\n", err))
				continue
			}

			formatDirectoryEntries(&resultBuilder, entries, 0)
			resultBuilder.WriteString("\n")
		}

		return mcp.NewToolResultText(resultBuilder.String()), nil
	})

	s.AddTool(writeFileTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if filesystem is in read-only mode
		if options.ReadOnly {
			return nil, fmt.Errorf("filesystem is in read-only mode")
		}

		path, ok := req.Params.Arguments["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		content, ok := req.Params.Arguments["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content must be a string")
		}

		createIfMissing := true
		if val, ok := req.Params.Arguments["create_if_missing"].(bool); ok {
			createIfMissing = val
		}

		// Verify the path is within allowed directories
		if !isPathAllowed(path, options.AllowedDirs) {
			return nil, fmt.Errorf("access to path is not allowed: %s", path)
		}

		// Check if file exists
		fileExists := false
		if _, err := os.Stat(path); err == nil {
			fileExists = true
		}

		// If file doesn't exist and we're not allowed to create it
		if !fileExists && !createIfMissing {
			return nil, fmt.Errorf("file doesn't exist and create_if_missing is false")
		}

		// Ensure the directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory structure: %v", err)
		}

		// Write the file
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file: %v", err)
		}

		// Generate success message
		action := "updated"
		if !fileExists {
			action = "created"
		}

		result := fmt.Sprintf("File successfully %s at %s (%d bytes written)",
			action, path, len(content))

		return mcp.NewToolResultText(result), nil
	})
}

func isPathAllowed(path string, allowedDirs []string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		rel, err := filepath.Rel(absDir, absPath)
		if err == nil && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}

func listDirectory(dirPath string, recursive bool, maxDepth, currentDepth int) ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		entryInfo := map[string]interface{}{
			"name":   entry.Name(),
			"path":   filepath.Join(dirPath, entry.Name()),
			"is_dir": entry.IsDir(),
		}

		// Add file info if it's a file
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				entryInfo["size"] = info.Size()
				entryInfo["modified"] = info.ModTime().Format(time.RFC3339)
			}
		}

		results = append(results, entryInfo)

		// Recursively list subdirectories if requested
		if recursive && entry.IsDir() {
			// Check max depth
			if maxDepth == 0 || currentDepth < maxDepth {
				subPath := filepath.Join(dirPath, entry.Name())
				subEntries, err := listDirectory(subPath, recursive, maxDepth, currentDepth+1)
				if err == nil {
					entryInfo["contents"] = subEntries
				}
			}
		}
	}

	return results, nil
}

func formatDirectoryEntries(builder *strings.Builder, entries []map[string]interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	for _, entry := range entries {
		isDir := entry["is_dir"].(bool)
		name := entry["name"].(string)

		if isDir {
			builder.WriteString(fmt.Sprintf("%s[DIR] %s\n", indentStr, name))

			// If this directory has contents (from recursive listing)
			if contents, ok := entry["contents"].([]map[string]interface{}); ok {
				formatDirectoryEntries(builder, contents, indent+1)
			}
		} else {
			// Format file with size if available
			sizeStr := ""
			if size, ok := entry["size"].(int64); ok {
				sizeStr = formatFileSize(size)
			}

			if sizeStr != "" {
				builder.WriteString(fmt.Sprintf("%s%s (%s)\n", indentStr, name, sizeStr))
			} else {
				builder.WriteString(fmt.Sprintf("%s%s\n", indentStr, name))
			}
		}
	}
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size < KB:
		return fmt.Sprintf("%d B", size)
	case size < MB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	case size < GB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	}
}
