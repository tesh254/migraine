# Workflows in Migraine

Workflows are the core concept in Migraine that allow you to define, organize, and execute sequences of shell commands. This section covers everything you need to know about workflows.

## Overview

A workflow in Migraine is a YAML file that defines a sequence of commands to be executed. Workflows can include pre-checks, main steps, and optional actions, providing a complete automation solution.

## Workflow Structure

A typical workflow includes several key sections:

### `name`
The unique identifier for the workflow. This is used when running the workflow: `migraine workflow run <name>`

### `description`
A human-readable description of what the workflow does. This is shown in workflow lists and help output.

### `pre_checks`
Commands that run before the main workflow steps. These are used for validation and verification. If any pre-check fails, the workflow execution stops. This is great for checking dependencies, connectivity, or prerequisites.

### `steps`
The main commands that make up the workflow. These execute in the order specified. Each step can have a description and command.

### `actions`
Optional commands that can be run independently of the main workflow. These are often used for cleanup, rollback, or deployment operations.

### `config`
Configuration options for the workflow, including variable definitions and transformations.

### `use_vault`
A boolean flag to determine variable resolution method. When `true`, resolves variables from the vault system. When `false`, uses environment files or prompts.

## Example Workflow

```yaml
name: deploy-app
description: Deploy application with validation and rollback

pre_checks:
  - command: "which rsync"
    description: "Verify rsync is available"
  - command: "ssh -q {{server}} exit"
    description: "Test server connectivity"

steps:
  - command: "npm run build"
    description: "Build the application"
  - command: "tar -czf app.tar.gz dist/"
    description: "Package for deployment"
  - command: "rsync -av app.tar.gz {{user}}@{{server}}:/tmp/"
    description: "Upload package to server"

actions:
  rollback:
    command: "ssh {{user}}@{{server}} 'cd /opt/app && git checkout HEAD~1'"
    description: "Rollback to previous version"
  cleanup:
    command: "rm -f app.tar.gz && ssh {{user}}@{{server}} 'rm -f /tmp/app.tar.gz'"
    description: "Clean up temporary files"

config:
  variables:
    server:
      - "slugify"
    user:
      - "validate_user"
  store_variables: false

use_vault: true
```

## Best Practices

### 1. Use Descriptive Names
Choose clear, descriptive names for your workflows that indicate their purpose.

### 2. Document Each Step
Include meaningful descriptions for each step to make the workflow self-documenting.

### 3. Leverage Pre-checks
Use pre-checks to validate prerequisites before executing main steps, preventing failures later in the workflow.

### 4. Organize Files
Keep workflow files organized in the `./workflows/` directory with clear naming conventions.

### 5. Handle Errors
Design workflows with error handling and validation to make them robust.

## File Discovery

Migraine automatically discovers workflow files in:
- `./workflows/` directory
- Current directory (`.`)
- Files with `.yaml` or `.yml` extensions

## Variable Substitution

Workflows support variable substitution using `{{variable_name}}` syntax:
1. Variables can be set via CLI flags: `migraine run my-workflow -v name=project`
2. Variables can be stored in the vault system
3. Variables can be loaded from environment files
4. Variables can be prompted during execution

## New Pre-checks Command

As of recent updates, Migraine includes a new `pre-checks` command that allows you to run only the pre-checks section of a workflow:
- `migraine workflow pre-checks` - Run pre-checks from current directory's migraine.yaml
- `migraine workflow pre-checks <workflow_name>` - Run pre-checks for a specific workflow from its stored directory