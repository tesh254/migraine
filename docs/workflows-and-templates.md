# Understanding Workflows and Templates in Migraine

## Overview

Migraine uses a powerful workflow system that allows you to automate complex sequences of shell commands. This document explains how workflows and templates work, their relationship, and best practices for using them effectively.

## Workflows vs Templates

### Workflows
A **workflow** in Migraine is a complete automation sequence that includes:
- Pre-checks (validation steps)
- Main execution steps
- Optional actions
- Configuration options
- Variable definitions

Workflows can be stored in two ways:
1. **File-based** (YAML files in your project directory)
2. **Database-stored** (in the SQLite database)

### Templates
A **template** is a legacy concept from Migraine v1.x that allowed users to create workflow definitions that could be customized with variables. In v2.0.0, direct template support has been replaced by the improved YAML workflow system, but the concepts remain relevant:

- **YAML workflows** serve the same purpose as the old templates
- **Variables** work similarly to customize workflow behavior
- **Migration** automatically converts old templates to workflow entries

## YAML Workflows

### Basic Structure

A typical YAML workflow looks like this:

```yaml
# my-project-deploy.yaml
name: deploy
description: Deploy the application to production servers

# Pre-checks ensure prerequisites are met before execution
pre_checks:
  # Verify git is available
  - command: "which git"
    description: "Verify git is installed"
  # Check if the deployment server is accessible
  - command: "ssh -q {{server_host}} exit"
    description: "Test SSH connectivity to server"

# Main steps of the workflow
steps:
  # Build the application
  - command: "npm run build"
    description: "Build the application"
  # Create deployment package
  - command: "tar -czf app.tar.gz dist/"
    description: "Package the application for deployment"
  # Upload to server
  - command: "scp app.tar.gz {{server_user}}@{{server_host}}:/tmp/"
    description: "Upload package to server"

# Optional actions that can be run independently
actions:
  rollback:
    command: "ssh {{server_user}}@{{server_host}} 'cd /opt/app && git checkout HEAD~1'"
    description: "Rollback to previous version"
  cleanup:
    command: "rm -f app.tar.gz && ssh {{server_user}}@{{server_host}} 'rm -f /tmp/app.tar.gz'"
    description: "Clean up temporary files"

# Configuration options
config:
  variables:
    server_host:
      - "slugify"  # Apply transformations to the variable
  store_variables: false

# Whether to use the vault for variable resolution
use_vault: true
```

### Workflow Sections Explained

#### `name`
- The unique identifier for the workflow
- Used when running the workflow: `migraine workflow run deploy`

#### `description`
- Human-readable description of what the workflow does
- Shown in workflow lists and help output

#### `pre_checks`
- Commands that run before the main workflow steps
- Used for validation and verification
- If any pre-check fails, the workflow execution stops
- Great for checking dependencies, connectivity, or prerequisites

#### `steps`
- The main commands that make up the workflow
- Execute in the order specified
- Each step can have a description and command

#### `actions`
- Optional commands that can be run independently of the main workflow
- Often used for cleanup, rollback, or deployment operations
- Can be run separately from the main workflow

#### `config`
- Configuration options for the workflow
- Defines variables and their transformations
- Controls whether variables are stored in the workflow

#### `use_vault`
- Boolean flag to determine variable resolution method
- When `true`, resolves variables from the vault system
- When `false`, uses environment files or prompts

### Variable Handling

#### In Workflows
Variables in workflows use the `{{variable_name}}` syntax:

```yaml
steps:
  - command: "echo 'Deploying {{app_name}} to {{environment}}'"
    description: "Show deployment info"
  - command: "git clone {{repo_url}} {{project_dir}}"
    description: "Clone the repository"
```

#### Variable Resolution Order
When a workflow is executed, Migraine resolves variables in this order:

1. **Command-line flags**: `migraine run my-workflow -v var1=value1`
2. **Workflow scope** (in vault): Specific to this workflow
3. **Project scope** (in vault): For the current project
4. **Global scope** (in vault): Available to all workflows
5. **Environment files**: `.env`, `./env/[workflow].env`
6. **Prompt user**: If no value is found

### Creating Workflows

#### Scaffolding New Workflows
Use the `init` command to create a new workflow with comments:

```bash
migraine workflow init deploy-app -d "Production deployment workflow"
```

This creates `./workflows/deploy-app.yaml` with a fully documented template.

#### Workflow Discovery
Migraine automatically discovers workflows in:

1. `./workflows/` directory
2. Current directory (`.`)
3. Files with `.yaml` or `.yml` extensions

### Running Workflows

#### Basic Execution
```bash
migraine workflow run my-workflow
```

#### With Variables
```bash
migraine workflow run my-workflow -v server=prod -v user=admin
```

#### Running Specific Actions
```bash
migraine run my-workflow -a cleanup
```

## Templates Migration

### Legacy Templates
In Migraine v1.x, templates were stored in the database and used to create workflows:

```json
{
  "name": "generic-deploy",
  "steps": [
    {
      "command": "git clone {{repo_url}} {{project_dir}}",
      "description": "Clone the repository"
    }
  ],
  "variables": {
    "project_dir": ["slugify"]
  }
}
```

### Migration Process
When you upgrade to v2.0.0:

1. **Automatic migration** converts old templates and workflows to the new SQLite format
2. **YAML files** become the primary format for new workflows
3. **Backward compatibility** is maintained for existing data

### Transitioning to YAML
New workflows should use the YAML format as it provides:
- Better documentation capabilities
- Version control friendliness
- Rich commenting support
- Improved variable management

## Best Practices

### Workflow Design
1. **Single Responsibility**: Each workflow should focus on one main task
2. **Descriptive Names**: Use clear, action-oriented names
3. **Comprehensive Pre-checks**: Verify all prerequisites before main execution
4. **Clear Descriptions**: Document what each step does
5. **Idempotent Operations**: Design workflows that can be safely rerun

### Variable Management
1. **Use Descriptive Names**: `server_host` instead of just `host`
2. **Group Related Variables**: Use prefixes like `db_`, `api_`, etc.
3. **Secure Sensitive Data**: Never hardcode secrets; use the vault system
4. **Provide Defaults**: Document expected variable values in comments

### Security Considerations
1. **Don't Hardcode Secrets**: Use vault variables for sensitive information
2. **Sanitize Inputs**: Be cautious with user-provided variable values
3. **Verify Permissions**: Ensure commands run with appropriate permissions
4. **Audit Trail**: Consider logging workflow execution for security review

## Example Workflow Patterns

### CI/CD Pipeline
```yaml
name: ci-cd-pipeline
description: Complete CI/CD pipeline for the application

pre_checks:
  - command: "which npm"
    description: "Verify npm is available"
  - command: "git status"
    description: "Verify git repo is clean"

steps:
  - command: "npm install"
    description: "Install dependencies"
  - command: "npm test"
    description: "Run tests"
  - command: "npm run build"
    description: "Build application"
  - command: "npm run deploy -- --env={{environment}}"
    description: "Deploy to environment"

actions:
  rollback:
    command: "npm run rollback -- --env={{environment}}"
    description: "Rollback deployment"
  notify:
    command: "echo 'Deployment to {{environment}} completed' | mail -s 'Deploy Complete' team@example.com"
    description: "Notify team of completion"

config:
  variables:
    environment:
      - "slugify"
  store_variables: false

use_vault: true
```

### Development Environment Setup
```yaml
name: dev-setup
description: Set up development environment

pre_checks:
  - command: "which docker"
    description: "Verify Docker is installed"
  - command: "which docker-compose"
    description: "Verify Docker Compose is available"

steps:
  - command: "cp .env.example .env"
    description: "Create environment file from example"
  - command: "docker-compose up -d"
    description: "Start development services"
  - command: "npm install"
    description: "Install project dependencies"
  - command: "npm run seed"
    description: "Seed development database"

config:
  store_variables: false

use_vault: false
```

## Troubleshooting

### Common Issues
1. **Variable Not Found**: Check the variable resolution order and make sure the variable is defined
2. **Workflow Not Found**: Verify the workflow file is in the right location with the correct extension
3. **Pre-check Failure**: Investigate the specific pre-check that failed and its dependencies
4. **Permission Issues**: Ensure the commands have appropriate permissions to execute

### Debugging Workflows
Use the `workflow info` command to see workflow details:
```bash
migraine workflow info my-workflow
```

This shows all the steps, variables, and configuration in the workflow.