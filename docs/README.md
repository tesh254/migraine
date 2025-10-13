# Migraine - Complete Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Getting Started](#getting-started)
4. [CLI Usage](#cli-usage)
5. [YAML Workflows](#yaml-workflows)
6. [Vault & Variables](#vault--variables)
7. [Comparison with Other Tools](#comparison-with-other-tools)
8. [Best Practices](#best-practices)
9. [Advanced Features](#advanced-features)

## Introduction

Migraine is a command-line tool for organizing and automating complex workflows with templated commands. It allows users to define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.

### Key Features
- **YAML-based workflows** with comprehensive documentation
- **Vault-backed variable management** with scope awareness (global, project, workflow)
- **File-based workflow discovery** in the current working directory
- **Template scaffolding** with commented examples
- **Migration from legacy formats** while maintaining compatibility

## Installation

### Prerequisites
- Go 1.24 or higher

### Install from Source
```bash
git clone https://github.com/tesh254/migraine.git
cd migraine
go build -o migraine .
sudo install migraine /usr/local/bin/migraine
```

### Using Go Install
```bash
go install github.com/tesh254/migraine@latest
```

## Getting Started

### Basic Commands
```bash
# Check version
migraine version

# View help
migraine --help

# Create your first workflow
migraine workflow init my-first-workflow -d "My first automated workflow"
```

### Quick Example
After installation, create a simple workflow:

1. Create a new workflow:
   ```bash
   migraine workflow init hello-world -d "A simple greeting workflow"
   ```

2. Edit the generated file in `./workflows/hello-world.yaml`

3. Run the workflow:
   ```bash
   migraine workflow run hello-world
   ```

## CLI Usage

### Core Commands

#### Workflow Management
- `migraine workflow init [name]` - Create a new YAML workflow with comments
- `migraine workflow list` - Show all workflows (database + file-based)
- `migraine workflow validate [path]` - Validate a workflow file
- `migraine workflow run [name]` - Execute a workflow
- `migraine workflow info [name]` - Show workflow details

#### Variable Management
- `migraine vars set [key] [value]` - Store a variable in the vault
- `migraine vars get [key]` - Retrieve a variable value
- `migraine vars list` - List all variables
- `migraine vars delete [key]` - Remove a variable

#### General Commands
- `migraine run [name]` - Execute a workflow (v2 command)
- `migraine version` - Show version information
- `migraine buildinfo` - Detailed build information

### Common Flags
- `-v, --var` - Set variables for workflow execution
- `-s, --scope` - Specify scope for variables (global, project, workflow)
- `-w, --workflow` - Specify workflow for variable operations
- `-d, --description` - Add description when creating workflows

## YAML Workflows

YAML workflows provide a structured, well-documented format for defining automated processes.

### Basic Structure
```yaml
# Example workflow file with comprehensive comments
name: my-workflow
description: A sample workflow to demonstrate features

# Pre-checks run before the main workflow steps
pre_checks:
  # Verify dependencies exist
  - command: "which git"
    description: "Check if git is installed"
  - command: "test -d {{project_dir}}"
    description: "Verify project directory exists"

# Main workflow steps executed in order
steps:
  - command: "echo 'Hello from workflow {{workflow_name}}!'"
    description: "Example step with variable substitution"
  - command: "git status"
    description: "Show git status"

# Optional actions that can be run independently
actions:
  cleanup:
    command: "rm -rf ./temp"
    description: "Clean up temporary files"
  deploy:
    command: "rsync -av ./dist/ user@server:/path/to/deploy/"
    description: "Deploy built files to server"

# Configuration for the workflow
config:
  variables:
    project_name:
      - "slugify"  # Apply transformations to variable values
  store_variables: false  # Whether to store variables in the workflow

# Whether to use the vault for variable resolution
use_vault: false
```

### Variable Substitution
Workflows support variable substitution using `{{variable_name}}` syntax:

1. Variables can be set via CLI flags: `migraine run my-workflow -v name=project`
2. Variables can be stored in the vault system
3. Variables can be loaded from environment files
4. Variables can be prompted during execution

### File Discovery
Migraine automatically discovers workflow files in:
- `./workflows/` directory
- Current directory (`.`)
- Files with `.yaml` or `.yml` extensions

## Vault & Variables

### Variable Scopes
Migraine supports three variable scopes:

1. **Global** - Available to all workflows
2. **Project** - Available to workflows in the same project
3. **Workflow** - Specific to one workflow

### Setting Variables
```bash
# Set global variable
migraine vars set api_key "my-secret-key"

# Set project-specific variable
migraine vars set project_path "/path/to/project" -s project

# Set workflow-specific variable
migraine vars set workflow_name "my-workflow" -s workflow -w my-workflow
```

### Using Variables in Workflows
When `use_vault: true` is set in a workflow:
- Variables are resolved from the vault according to scope precedence
- Fallback to environment files if vault doesn't contain the variable
- Prompt for missing variables if no fallback exists

When `use_vault: false`:
- Variables are loaded from environment files (`.env` or workflow-specific files)
- Prompt for missing variables

## Comparison with Other Tools

### vs Makefile

| Feature | Migraine | Makefile |
|---------|----------|----------|
| Language | YAML/CLI | Make syntax |
| Variable Management | Advanced vault system | Simple environment/shell variables |
| Workflow Discovery | Automatic in directory | Manual target specification |
| Variable Scoping | Global/Project/Workflow | Project-wide only |
| Comments | Rich documentation support | Comment lines |
| Dependencies | Built-in pre-checks | Explicit dependency declarations |
| Execution Context | Isolated shell | Same shell context |
| IDE Support | YAML with syntax highlighting | Specialized Makefile editors |
| Learning Curve | Low | Moderate (Make-specific syntax) |
| Complex Actions | Yes (actions section) | Requires shell scripting |
| Configuration | Per-workflow or centralized | Per-Makefile |

### When to Use Migraine vs Makefile

**Use Migraine when:**
- You want YAML-based configuration with rich documentation
- You need advanced variable management with scoping
- You want automatic workflow discovery and execution
- You need vault-backed secure variable storage
- You want a more modern CLI experience
- You need complex variable substitution and transformations

**Use Makefile when:**
- You need complex build dependency management
- You're building C/C++ projects or similar
- You want minimal dependencies (make is in most Unix systems)
- You need file generation rules
- You're comfortable with Make's syntax and conventions

### vs Other Task Runners

Compared to tools like npm scripts, Rake, or Task, Migraine offers:
- Cross-language support (not tied to specific ecosystems)
- Vault-backed variable management
- File-based workflow discovery
- YAML configuration with rich comments
- Scope-aware variables

## Best Practices

### Workflow Organization
1. **Group related workflows** in the `./workflows/` directory
2. **Use descriptive names** that clearly indicate the workflow's purpose
3. **Include detailed descriptions** in YAML files
4. **Document variables** with clear usage instructions
5. **Use pre-checks** to verify prerequisites before execution

### Variable Management
1. **Use appropriate scopes** - global for shared secrets, workflow for specific values
2. **Encrypt sensitive data** in the vault
3. **Use environment files** for local development variations
4. **Document variable requirements** in workflow files

### Security Considerations
1. **Don't hardcode secrets** in workflow files
2. **Use vault for sensitive variables** like API keys
3. **Review environment files** to avoid committing secrets
4. **Use appropriate file permissions** for workflow directories

## Advanced Features

### Custom Workflow Discovery
Migraine looks for workflows in specific patterns:
- `./workflows/*.yaml`
- `./workflows/*.yml`
- `./migraine.yaml` or `./migraine.yml` (for project-specific workflows)
- `./.migraine.yaml` or `./.migraine.yml` (hidden workflow files)

### Migration from Legacy Formats
Migraine automatically migrates from older Badger-based storage to the new SQLite system. Existing workflows and templates are preserved during migration.

### Environment Integration
Variables can be loaded from:
- Command-line flags
- Vault storage
- Environment files (`./.env`, `./env/.env`, `./env/[workflow].env`)
- Default values specified in workflows