# Migraine Plugin System Research

## Overview

This document captures extensive research and design concepts for implementing a comprehensive plugin system for Migraine. The system should enable extension of Migraine's functionality without modifying core source code, supporting both workflow-integrated plugins and standalone management tools.

## 1. Hybrid Plugin Architecture (Bash + Python)

### 1.1 Plugin Configuration Format

```yaml
# plugins/log-processor.yaml
name: log-processor
version: 1.0.0
type: command-output-interceptor
description: Processes command output and logs to file

hooks:
  - event: command-output
    pattern: ".*"  # Apply to all commands
    handler:
      type: python  # or 'bash'
      module: log_processor  # Python module name
      function: process_log
    script: |
      #!/bin/bash
      # Fallback bash script if Python handler fails
      echo "$(date): $MIGRAINE_CMD_OUTPUT" >> "$MIGRAINE_LOG_FILE"

config:
  log_file: "./workflow.log"
  max_file_size: "10MB"
```

### 1.2 Python Plugin Interface

```python
# log_processor.py
import json
import os
from datetime import datetime

def process_log(plugin_config: dict, context: dict) -> dict:
    """
    Process log data from Migraine workflow
    
    Args:
        plugin_config: Configuration from the plugin YAML
        context: Runtime context with workflow data
        
    Returns:
        dict: Results of processing, can contain 'status', 'output', etc.
    """
    # Extract data from context
    command_output = context.get('output', '')
    command_name = context.get('command', '')
    workflow_name = context.get('workflow', {})
    
    # Extract plugin configuration
    log_file = plugin_config.get('log_file', './default.log')
    max_size = plugin_config.get('max_file_size', '10MB')
    
    # Process the data
    timestamp = datetime.now().isoformat()
    log_entry = f"[{timestamp}] [{workflow_name.get('name', 'unknown')}] [{command_name}]: {command_output}\n"
    
    # Write to file with size check
    if should_rotate(log_file, max_size):
        rotate_log(log_file)
    
    with open(log_file, 'a') as f:
        f.write(log_entry)
    
    return {
        'status': 'success',
        'output': f'Logged to {log_file}',
        'processed_lines': 1
    }

def should_rotate(log_file: str, max_size: str) -> bool:
    """Check if log file should be rotated based on size"""
    if not os.path.exists(log_file):
        return False
    
    size_map = {'KB': 1024, 'MB': 1024**2, 'GB': 1024**3}
    max_bytes = int(max_size[:-2]) * size_map[max_size[-2:]]
    
    return os.path.getsize(log_file) > max_bytes

def rotate_log(log_file: str):
    """Rotate the log file"""
    import shutil
    backup_file = f"{log_file}.old"
    if os.path.exists(backup_file):
        os.remove(backup_file)
    shutil.move(log_file, backup_file)
```

### 1.3 Context Interface

The system passes consistent context to both bash and Python plugins:

```json
{
  "workflow": {
    "name": "deploy-app",
    "id": "workflow-123",
    "start_time": "2024-01-15T10:30:45Z"
  },
  "step": {
    "name": "build-app",
    "description": "Build the application",
    "index": 2
  },
  "command": {
    "original": "npm run build",
    "resolved": "npm run build",
    "exit_code": 0
  },
  "output": "Build completed successfully",
  "timestamp": "2024-01-15T10:31:22Z",
  "variables": {
    "env": "production",
    "server": "server-01"
  },
  "plugin_config": {
    "log_file": "/var/log/build.log",
    "format": "json"
  }
}
```

## 2. Plugin Hook System

### 2.1 Hook Points

- **Command-level Hooks**:
  - `pre-command`: Before command execution
  - `post-command`: After command execution
  - `command-output`: When command produces output
  - `command-error`: When command fails

- **Workflow-level Hooks**:
  - `pre-workflow`: Before workflow starts
  - `post-workflow`: After workflow completes
  - `workflow-progress`: During workflow execution
  - `workflow-error`: When workflow fails

- **System-level Hooks**:
  - `workflow-load`: When workflow is loaded
  - `variable-resolution`: During variable processing
  - `pre-check-status`: After pre-check completion

### 2.2 Example Hook Configuration

```yaml
hooks:
  - event: post-command
    pattern: "server.*start"
    handler:
      type: python
      module: server_monitor
      function: monitor_server
    on_error:
      type: bash
      script: |
        # Fallback bash script
        echo "Server monitoring failed, using bash fallback" >> /tmp/migraine.log
```

## 3. Python Plugin Security and Sandboxing

### 3.1 Migraine CLI-Managed Python Sandbox

```go
// Go code to execute Python plugins with security restrictions
func ExecutePythonPlugin(moduleName string, functionName string, pluginConfig map[string]interface{}, context map[string]interface{}) (map[string]interface{}, error) {
    // Create temporary Python execution script
    pythonCode := createPythonExecutionScript(moduleName, functionName, pluginConfig, context)
    
    // Create temporary file with resource limits
    tempDir, err := os.MkdirTemp("", "migraine_plugin_*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tempDir) // Clean up after execution
    
    scriptPath := filepath.Join(tempDir, "plugin_executor.py")
    if err := os.WriteFile(scriptPath, []byte(pythonCode), 0600); err != nil {
        return nil, err
    }
    
    // Execute with timeout and resource limits
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "python3", scriptPath)
    
    // Set resource limits using syscall
    if cmd.SysProcAttr == nil {
        cmd.SysProcAttr = &syscall.SysProcAttr{}
    }
    
    output, err := cmd.CombinedOutput()
    
    if ctx.Err() == context.DeadlineExceeded {
        return nil, fmt.Errorf("plugin execution timed out")
    }
    
    if err != nil {
        return nil, fmt.Errorf("plugin execution failed: %v, output: %s", err, string(output))
    }
    
    // Parse the result as JSON
    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("plugin returned invalid JSON: %v", err)
    }
    
    return result, nil
}
```

### 3.2 Python Security Implementation

```python
# Security features for Python plugins
import sys
import json
import resource
import os
import time
import math
import re
import urllib.parse
import collections
import itertools
import datetime
import hashlib
import base64
import uuid
import enum
import dataclasses
import functools
import operator
import statistics
import decimal
import fractions
import array
import heapq
import bisect
import copy
import pprint
import traceback
import warnings
import weakref
import gc
import atexit
import logging
import pathlib
import shutil
import subprocess
import csv
import zipfile
import gzip
import fileinput
import io
import codecs
import ast
import inspect
import site
import tokenize
import token

# SECURITY: Restrict available modules
def secure_import(name, globals=None, locals=None, fromlist=(), level=0):
    allowed_modules = {
        'json', 'os', 'sys', 'time', 'datetime', 're', 'urllib.parse', 
        'collections', 'itertools', 'math', 'random', 'string', 'pathlib',
        'shutil', 'hashlib', 'base64', 'uuid', 'enum', 'dataclasses',
        'functools', 'operator', 'statistics', 'decimal', 'fractions',
        'array', 'heapq', 'bisect', 'copy', 'pprint', 'csv', 'zipfile',
        'gzip', 'fileinput', 'io', 'codecs', 'ast', 'inspect', 'site',
        'tokenize', 'token', 'traceback', 'warnings', 'weakref', 'gc',
        'atexit', 'logging', 'subprocess'
    }
    
    module_parts = name.split('.')
    if module_parts[0] not in allowed_modules:
        raise ImportError(f"Module '{name}' is not allowed")
    
    return __import__(name, globals, locals, fromlist, level)

# Override built-in __import__ to restrict imports
__builtins__['__import__'] = secure_import

# Set resource limits
try:
    resource.setrlimit(resource.RLIMIT_AS, (256 * 1024 * 1024, 256 * 1024 * 1024))  # 256MB
    resource.setrlimit(resource.RLIMIT_CPU, (30, 30))  # 30 seconds
    resource.setrlimit(resource.RLIMIT_STACK, (10 * 1024 * 1024, 10 * 1024 * 1024))  # 10MB stack
except:
    pass
```

## 4. Standalone Plugin System for External Management

### 4.1 Plugin Structure

```
plugins/
├── docker-manager/
│   ├── plugin.yaml          # Plugin metadata and capabilities
│   ├── docker-manager.go    # Main plugin executable code
│   └── Dockerfile           # For containerized plugin (optional)
├── kubernetes-manager/
│   ├── plugin.yaml
│   ├── k8s-manager.go
│   └── config.example
├── file-operations/
│   ├── plugin.yaml
│   └── file-plugin.go
└── notification/
    ├── plugin.yaml
    ├── notifier.go
    └── webhook-handler.go
```

### 4.2 Plugin Definition Example

```yaml
# plugins/docker-manager/plugin.yaml
name: docker-manager
version: "1.0.0"
description: "Docker container and image management tools"
author: "Plugin Developer"
license: "MIT"

# Plugin capabilities and commands it provides
commands:
  - name: docker-list-containers
    description: "List running Docker containers"
    usage: "migraine plugin docker-manager docker-list-containers [options]"
  - name: docker-build-image
    description: "Build Docker image from context"
    usage: "migraine plugin docker-manager docker-build-image --file Dockerfile --tag myimage:latest"
  - name: docker-deploy
    description: "Deploy container to Docker"
    usage: "migraine plugin docker-manager docker-deploy --image myapp:latest --port 8080"
  - name: docker-logs
    description: "Get logs from container"
    usage: "migraine plugin docker-manager docker-logs --container myapp"

# Dependencies this plugin requires
dependencies:
  - docker
  - docker-socket  # Unix socket access

# Configuration options
config:
  docker_socket: "/var/run/docker.sock"
  default_registry: "docker.io"

hooks:
  # Can still have hooks that integrate with workflows if needed
  - event: "workflow:post-deploy"
    command: "docker-logs --container {{container_name}}"

# Installation requirements
install_requires:
  - "docker CLI tools"
  - "Docker daemon running"
```

### 4.3 Go-Based Plugin Example

```go
// plugins/docker-manager/docker-manager.go
package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
    
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }
    
    command := os.Args[1]
    
    switch command {
    case "docker-list-containers":
        listContainers()
    case "docker-build-image":
        buildImage()
    case "docker-deploy":
        deployContainer()
    case "docker-logs":
        getLogs()
    case "--help", "-h":
        printUsage()
    default:
        fmt.Printf("Unknown command: %s\n", command)
        printUsage()
        os.Exit(1)
    }
}

func listContainers() {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        fmt.Printf("Error connecting to Docker: %v\n", err)
        os.Exit(1)
    }
    
    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
    if err != nil {
        fmt.Printf("Error listing containers: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("CONTAINER ID\tIMAGE\t\tSTATUS\t\tNAMES")
    for _, container := range containers {
        fmt.Printf("%s\t%s\t%s\t%s\n",
            container.ID[:12],
            container.Image,
            container.Status,
            strings.Join(container.Names, ","))
    }
}
```

### 4.4 Plugin Management Commands

```bash
# Install a Docker management plugin
migraine plugin install https://github.com/user/docker-manager-plugin

# List available plugins
migraine plugin list

# Use Docker plugin to list containers
migraine plugin execute docker-manager docker-list-containers

# Build Docker image using plugin
migraine plugin execute docker-manager docker-build-image --file Dockerfile --tag myapp:latest

# Deploy container
migraine plugin execute docker-manager docker-deploy --image myapp:latest --port 8080

# Get logs
migraine plugin execute docker-manager docker-logs --container myapp
```

## 5. Plugin Registry and Management

### 5.1 PluginInfo Structure

```go
type PluginInfo struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Description  string            `json:"description"`
    Commands     []PluginCommand   `json:"commands"`
    Dependencies []string          `json:"dependencies"`
    Config       map[string]string `json:"config"`
}

type PluginCommand struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Usage       string `json:"usage"`
}
```

### 5.2 Plugin Registry Implementation

```go
// PluginRegistry manages installed plugins
type PluginRegistry struct {
    pluginsDir string
}

func NewPluginRegistry() *PluginRegistry {
    // Find plugins directory (could be in ~/.migraine/plugins or ./plugins)
    homeDir, _ := os.UserHomeDir()
    pluginsDir := filepath.Join(homeDir, ".migraine", "plugins")
    
    return &PluginRegistry{
        pluginsDir: pluginsDir,
    }
}

func (pr *PluginRegistry) ListPlugins() ([]PluginInfo, error) {
    var plugins []PluginInfo
    
    entries, err := ioutil.ReadDir(pr.pluginsDir)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        if entry.IsDir() {
            pluginDir := filepath.Join(pr.pluginsDir, entry.Name())
            pluginInfo, err := pr.loadPluginInfo(pluginDir)
            if err != nil {
                fmt.Printf("Warning: Could not load plugin %s: %v\n", entry.Name(), err)
                continue
            }
            plugins = append(plugins, pluginInfo)
        }
    }
    
    return plugins, nil
}
```

## 6. Security Considerations

### 6.1 Resource Limits
- Memory limit: 256MB per plugin
- CPU time limit: 30 seconds per plugin
- Stack size limit: 10MB per plugin

### 6.2 Import Restrictions
- Only standard library modules allowed
- Whitelist-based module access
- No third-party library imports

### 6.3 File System Access
- Plugins operate in temporary directories
- Whitelist for file system access
- No direct access to Migraine internals

### 6.4 Network Access
- Configurable network access (default: disabled)
- Outbound connection restrictions
- Proxy configuration support

## 7. Plugin Development Best Practices

### 7.1 Standard Library Only
- Use only Python standard library for Python plugins
- No external dependencies for security and portability
- Follow security guidelines for input validation

### 7.2 Error Handling
- Proper error reporting with JSON output
- Graceful degradation for missing features
- Clear error messages for debugging

### 7.3 Testing
- Include test cases with plugins
- Validate plugin configuration
- Test resource limit compliance

## 8. Integration with Workflow System

### 8.1 Hook Integration
- Plugins can hook into workflow events
- Context passed from workflows to plugins
- Result handling and workflow continuation

### 8.2 Variable Handling
- Plugins can access workflow variables
- Variable substitution in plugin configuration
- Secure variable handling and storage

### 8.3 Logging and Monitoring
- Plugin activity logging
- Performance monitoring
- Error tracking and reporting

## Conclusion

The proposed plugin system provides a comprehensive extension mechanism for Migraine that supports:
1. Workflow-integrated plugins (Bash + Python standard library)
2. Standalone system management plugins (Go executables)
3. Robust security and sandboxing
4. Flexible hook system
5. Centralized plugin management
6. Easy distribution and installation

This architecture enables powerful extensibility while maintaining security and stability of the core Migraine system.