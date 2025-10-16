# Vault & Variables in Migraine

The vault system in Migraine provides secure variable management with scope awareness. This section covers how to use and manage variables in your workflows.

## Overview

The vault system allows you to store, manage, and use variables in your workflows. It provides a secure way to handle sensitive data like API keys, passwords, and other configuration values.

### Security Notice
⚠️ **IMPORTANT**: The current vault implementation stores variables in an unencrypted SQLite database. While the variables are stored locally and not transmitted over networks, they are not encrypted at rest. We are actively working on adding encryption support to the vault system in an upcoming release to enhance security. For now, we recommend avoiding storing highly sensitive information like production API keys in the vault until encryption is implemented.

## Variable Scopes

Migraine supports three variable scopes:

1. **Global** - Available to all workflows
2. **Project** - Available to workflows in the same project
3. **Workflow** - Specific to one workflow

## Managing Variables

### Setting Variables

```bash
# Set global variable
migraine vars set api_key "my-api-key"

# Set project-specific variable
migraine vars set project_path "/path/to/project" -s project

# Set workflow-specific variable
migraine vars set workflow_name "my-workflow" -s workflow -w my-workflow
```

### Listing Variables

```bash
# List all variables
migraine vars list

# List variables with details
$ migraine vars list
Key: database_url
Scope: global
Value: postgresql://localhost/mydb
Created: 2024-01-15 10:30:45
Updated: 2024-01-15 10:30:45
```

### Getting Variables

```bash
# Get a specific variable
migraine vars get api_key
```

### Deleting Variables

```bash
# Delete a variable
migraine vars delete api_key
```

## Using Variables in Workflows

### Configuration

When `use_vault: true` is set in a workflow:
- Variables are resolved from the vault according to scope precedence
- Fallback to environment files if vault doesn't contain the variable
- Prompt for missing variables if no fallback exists

When `use_vault: false`:
- Variables are loaded from environment files (`.env` or workflow-specific files)
- Prompt for missing variables

### Variable Resolution Order

When a workflow runs, variables are resolved in this order:
1. Command-line flags: `migraine run workflow -v var=value`
2. Workflow scope (in vault)
3. Project scope (in vault)
4. Global scope (in vault)
5. Environment files (`.env`, `./env/[workflow].env`)
6. Prompt user for missing variables

## WORKING_DIR Feature

As of recent updates, Migraine automatically stores the working directory of each workflow as a vault variable:
- The `WORKING_DIR` variable is stored with workflow scope
- This enables the `migraine workflow pre-checks <workflow_name>` command to run pre-checks from the stored directory
- The system changes to the stored directory, executes the pre-checks, then restores the original directory

## Practical Examples

### Example 1: Storing and Using API Keys

```bash
# First, store your API key in the vault
migraine vars set openai_api_key "sk-..."

# Create a workflow that uses the variable
# (workflows/ai-request.yaml)
use_vault: true

steps:
  - command: "curl -H 'Authorization: Bearer {{openai_api_key}}' https://api.openai.com/v1/models"
    description: "Request available models"
```

### Example 2: Environment-Specific Variables

```bash
# Set different environment variables
migraine vars set api_url "https://dev-api.example.com" -s project
migraine vars set db_host "dev-db.example.com" -s project

# Workflow file (workflows/test-api.yaml)
name: test-api
use_vault: true

steps:
  - command: "curl -X GET {{api_url}}/health"
    description: "Test API health endpoint"
  - command: "ping -c 1 {{db_host}}"
    description: "Test database connectivity"
```

## Advanced Features

### Variable Transformations

You can apply transformations to variables in the workflow configuration:

```yaml
config:
  variables:
    project_name:
      - "slugify"        # Convert to lowercase, replace spaces with hyphens
    deploy_path:
      - "validate_path"  # Custom validation
    user_email:
      - "validate_email" # Format validation
```

### Validation in Workflows

Use pre-checks to validate that required variables are provided:

```yaml
pre_checks:
  # Validate that required variables are provided
  - command: "[ -n '{{api_key}}' ] || (echo 'API_KEY is required' && exit 1)"
    description: "Validate API key is provided"
  - command: "[ -n '{{environment}}' ] || (echo 'ENVIRONMENT is required' && exit 1)"
    description: "Validate environment is specified"
```

## Best Practices

### 1. Use Appropriate Scoping
- Use global variables for values needed across all workflows
- Use project variables for values specific to a project
- Use workflow variables for values specific to a single workflow

### 2. Secure Handling
- Avoid storing highly sensitive information until encryption is implemented
- Use environment variables for production secrets when possible
- Regularly audit stored variables

### 3. Documentation
- Document the purpose of variables in comments
- Use descriptive variable names
- Maintain consistency in naming conventions

### 4. Validation
- Validate required variables in pre-checks
- Use transformations to ensure proper formatting
- Test workflows with different variable values