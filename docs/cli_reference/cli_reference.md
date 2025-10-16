# CLI Reference for Migraine

This document provides a complete reference for all Migraine CLI commands and their usage.

## Core Commands

### `migraine`

The main entry point for the Migraine CLI. Shows version information and ASCII art when run without arguments.

```bash
migraine                    # Show version and info
migraine --version         # Show version information
migraine version           # Show version information
```

### `migraine workflow`

Manage workflows - create, list, run, validate, and get info about workflows.

```bash
migraine workflow --help   # Show workflow command help
```

#### `migraine workflow init [name]`

Create a new YAML workflow with comprehensive comments and documentation.

```bash
# Create a project configuration file (migraine.yml)
migraine workflow init

# Create a project configuration file as migraine.yml
migraine workflow init --yml

# Create a project configuration file as migraine.json
migraine workflow init --json

# Create a named workflow in workflows/ directory
migraine workflow init my-workflow -d "Description of my workflow"

# Aliases
migraine init [name]       # Alias for workflow init
```

#### `migraine workflow list`

Show all workflows (both database and file-based).

```bash
migraine workflow list     # List all workflows
migraine workflow ls       # Alias for list
```

#### `migraine workflow validate [path]`

Validate a workflow file.

```bash
migraine workflow validate workflows/my-workflow.yaml
```

#### `migraine workflow run [name]`

Execute a workflow.

```bash
# Run project workflow from current directory
migraine workflow run

# Run specific workflow
migraine workflow run my-workflow

# Run with variables
migraine workflow run my-workflow -v var1=value1 -v var2=value2

# Run specific action
migraine workflow run my-workflow -a deploy

# Run multiple actions
migraine workflow run my-workflow -a deploy -a notify
```

#### `migraine workflow pre-checks [name]`

Run only the pre-checks section of a workflow.

```bash
# Run pre-checks from current directory's migraine.yaml
migraine workflow pre-checks

# Run pre-checks for a specific workflow from its stored directory
migraine workflow pre-checks my-workflow

# Run with variables
migraine workflow pre-checks my-workflow -v var1=value1
```

#### `migraine workflow info [name]`

Show detailed information about a workflow.

```bash
migraine workflow info my-workflow
```

### `migraine run [name]`

Execute a workflow (v2 command - similar to workflow run but at top level).

```bash
# Run project workflow from current directory
migraine run

# Run specific workflow
migraine run my-workflow

# Run with variables
migraine run my-workflow -v var1=value1

# Run specific action
migraine run my-workflow -a deploy
```

### `migraine vars`

Manage variables in the vault system.

```bash
migraine vars --help       # Show variables command help
```

#### `migraine vars set [key] [value]`

Store a variable in the vault.

```bash
# Set global variable
migraine vars set api_key "my-api-key"

# Set project variable
migraine vars set project_path "/path/to/project" -s project

# Set workflow variable
migraine vars set workflow_var "value" -s workflow -w workflow-name
```

#### `migraine vars get [key]`

Retrieve a variable value.

```bash
migraine vars get api_key
```

#### `migraine vars list`

List all variables.

```bash
migraine vars list
```

#### `migraine vars delete [key]`

Remove a variable.

```bash
migraine vars delete api_key
```

### `migraine version`

Show version information in various formats.

```bash
migraine version                    # Detailed version info
migraine version --json             # JSON format
migraine version -s                 # Short format
migraine version -c                 # With commit hash
```

### `migraine buildinfo`

Show detailed build information.

```bash
migraine buildinfo
```

## Common Flags

### Variable Flags
- `-v, --var` - Set variables for workflow execution (format: KEY=VALUE)
```bash
migraine workflow run my-workflow -v name=value -v env=production
```

### Action Flags
- `-a, --action` - Specify actions to run (instead of main steps)
```bash
migraine workflow run my-workflow -a deploy -a cleanup
```

### Scope Flags
- `-s, --scope` - Specify scope for variable operations (global, project, workflow)
```bash
migraine vars set my_var "value" -s project
```

### Workflow Flags
- `-w, --workflow` - Specify workflow for variable operations
```bash
migraine vars set my_var "value" -s workflow -w my-workflow
```

### Description Flags
- `-d, --description` - Add description when creating workflows
```bash
migraine workflow init my-workflow -d "Description here"
```

## Special Files

### Configuration Files
Migraine looks for these project-level configuration files:
- `./migraine.yaml`
- `./migraine.yml`
- `./migraine.json`
- `./workflows/*.yaml`
- `./workflows/*.yml`

### Environment Files
- `.env` - Default environment file
- `./env/[workflow].env` - Workflow-specific environment files

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Workflow or command not found
- `3` - Validation error
- `4` - Pre-check failure

## Examples

### Basic Workflow Management
```bash
# Create a new workflow
migraine workflow init deploy-app -d "Deploy application to server"

# List all workflows
migraine workflow list

# Validate a workflow file
migraine workflow validate workflows/deploy-app.yaml

# Run a workflow
migraine workflow run deploy-app

# Run with variables
migraine workflow run deploy-app -v server=prod -v user=deployer
```

### Variable Management
```bash
# Set variables
migraine vars set api_key "secret123"
migraine vars set db_host "localhost" -s project

# List variables
migraine vars list

# Run workflow using variables
migraine workflow run my-workflow
```

### Advanced Usage
```bash
# Run only specific actions
migraine workflow run deploy-app -a rollback

# Run pre-checks only from stored directory
migraine workflow pre-checks my-workflow

# Run with multiple variables
migraine run deploy-app -v env=staging -v region=us-west-2
```