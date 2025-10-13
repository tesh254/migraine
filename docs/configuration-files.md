# Migraine Configuration Files (migraine.yml / migraine.json)

## Overview

Migraine supports project-level configuration files that allow you to define default workflows, environment settings, and project-specific configurations. This document explains how to use `migraine.yml` and `migraine.json` files for enhanced workflow management.

## Project-Level Configuration Files

### File Locations

Migraine looks for configuration files in these locations (in order of preference):

1. `./migraine.yaml` - Project-specific workflows and configuration
2. `./migraine.yml` - Alternative YAML format
3. `./.migraine.yaml` - Hidden configuration file
4. `./.migraine.yml` - Hidden configuration file
5. `./migraine.json` - JSON format (legacy support)

### Multiple Configuration Pattern

You can also organize workflows in multiple files:

- `./workflows/*.yaml` - All YAML workflow files in the workflows directory
- `./workflows/*.yml` - Alternative YAML extension
- `./workflows/*.json` - JSON workflow files (for legacy compatibility)

## Structure of Configuration Files

### YAML Format (migraine.yml)

```yaml
# migraine.yml - Project-level workflow configuration

# Metadata about the project
project:
  name: "My Awesome Project"
  description: "A sample project with automated workflows"
  version: "1.0.0"

# Default configuration for all workflows in this project
defaults:
  use_vault: true
  environment: "development"

# Pre-defined workflows (these can also be in separate files)
workflows:
  setup:
    description: "Set up the development environment"
    pre_checks:
      - command: "which node"
        description: "Check if Node.js is installed"
      - command: "which docker"
        description: "Check if Docker is available"
    steps:
      - command: "npm install"
        description: "Install project dependencies"
      - command: "docker-compose up -d"
        description: "Start required services"
    use_vault: true

  test:
    description: "Run project tests"
    steps:
      - command: "npm run test:unit"
        description: "Run unit tests"
      - command: "npm run test:integration"
        description: "Run integration tests"
    config:
      store_variables: false

  build:
    description: "Build the project"
    steps:
      - command: "npm run build"
        description: "Build the application"
    use_vault: false

  deploy:
    description: "Deploy to specified environment"
    pre_checks:
      - command: "npm run test"
        description: "Run tests before deployment"
    steps:
      - command: "npm run build"
        description: "Build the application"
      - command: "rsync -av ./dist/ {{deployment_target}}:/var/www/"
        description: "Deploy to target server"
    actions:
      rollback:
        command: "rsync -av ./dist_old/ {{deployment_target}}:/var/www/"
        description: "Rollback to previous version"
    variables:
      - "deployment_target"
    use_vault: true

# Environment-specific configuration
environments:
  development:
    variables:
      NODE_ENV: "development"
      DEBUG: "true"
  staging:
    variables:
      NODE_ENV: "staging"
      API_URL: "https://staging-api.example.com"
  production:
    variables:
      NODE_ENV: "production"
      API_URL: "https://api.example.com"

# Hooks for specific events
hooks:
  pre_workflow:
    - command: "echo 'Starting workflow: {{workflow_name}} at $(date)'"
      description: "Log workflow start time"
  post_workflow:
    - command: "echo 'Workflow {{workflow_name}} completed at $(date)'"
      description: "Log workflow completion time"
```

### JSON Format (migraine.json)

```json
{
  "project": {
    "name": "My Project",
    "description": "A sample project with automated workflows",
    "version": "1.0.0"
  },
  "defaults": {
    "use_vault": true,
    "environment": "development"
  },
  "workflows": {
    "setup": {
      "description": "Set up the development environment",
      "pre_checks": [
        {
          "command": "which node",
          "description": "Check if Node.js is installed"
        }
      ],
      "steps": [
        {
          "command": "npm install",
          "description": "Install project dependencies"
        }
      ],
      "use_vault": true
    }
  },
  "environments": {
    "development": {
      "variables": {
        "NODE_ENV": "development"
      }
    }
  }
}
```

## Using Configuration Files

### Auto-Discovery

Migraine automatically discovers and loads configuration files when you run any command in the project directory:

```bash
# Run this in your project directory containing migraine.yml
migraine workflow list  # Will show workflows defined in migraine.yml
migraine workflow run setup  # Runs the 'setup' workflow from the config
```

### Overriding Defaults

You can override project defaults with command-line flags:

```bash
# Override the default environment
migraine workflow run deploy -v environment=production

# Use specific variables instead of vault
migraine workflow run deploy --var server=prod-server --var user=admin
```

### Environment Switching

Use the environment configuration to switch between different deployment targets:

```bash
# Use development settings
migraine workflow run deploy -v environment=development

# Use production settings
migraine workflow run deploy -v environment=production
```

## Advanced Configuration Features

### Conditional Steps

While YAML doesn't support true conditionals, you can achieve similar behavior with variable-based logic:

```yaml
# migraine.yml
workflows:
  conditional-deploy:
    description: "Deploy with conditional steps"
    steps:
      # This command will only do something if the variable is set appropriately
      - command: "if [ '{{debug}}' = 'true' ]; then echo 'Debug mode enabled'; fi"
        description: "Conditional debug step"
      - command: "echo 'Always runs'"
        description: "This always runs"
```

### Variable Validation

You can implement basic validation by using pre-checks:

```yaml
workflows:
  validate-vars:
    description: "Workflow with variable validation"
    pre_checks:
      - command: "[ -n '{{required_var}}' ] || (echo 'Required variable missing' && exit 1)"
        description: "Validate required variable is not empty"
      - command: "echo '{{email_var}}' | grep -E '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}$' || (echo 'Invalid email format' && exit 1)"
        description: "Validate email format"
    steps:
      - command: "echo 'Using email: {{email_var}}'"
        description: "Use validated email"
```

### Environment-Specific Variables

```yaml
# migraine.yml
environments:
  development:
    variables:
      API_URL: "http://localhost:3000"
      DB_HOST: "localhost"
      DEBUG: "true"
  staging:
    variables:
      API_URL: "https://staging.example.com"
      DB_HOST: "staging-db.example.com"
      DEBUG: "false"
  production:
    variables:
      API_URL: "https://api.example.com"
      DB_HOST: "prod-db.example.com"
      DEBUG: "false"

workflows:
  deploy:
    description: "Deploy with environment-specific settings"
    steps:
      - command: "echo 'Deploying to API: {{API_URL}}'"
        description: "Show deployment target"
```

## Working with Multiple Configuration Files

### Organizing Complex Projects

For larger projects, you can split workflows across multiple files:

```
my-project/
├── migraine.yml              # Project metadata and common workflows
├── workflows/
│   ├── frontend.yaml         # Frontend-specific workflows
│   ├── backend.yaml          # Backend-specific workflows
│   └── database.yaml         # Database-related workflows
└── env/
    ├── development.env       # Development environment variables
    └── production.env        # Production environment variables
```

### Inheritance and Merging

When multiple configuration files exist, Migraine handles them as follows:

1. **Project file first**: `migraine.yml` provides base configuration
2. **Workflow files**: Individual workflow files in `./workflows/` are merged
3. **Environment files**: `.env` files provide variable overrides
4. **Command-line**: Flags take highest precedence

## Best Practices for Configuration Files

### 1. Use YAML over JSON

YAML provides better readability and supports comments:

```yaml
# Good: YAML with comments
workflows:
  setup:  # Sets up the development environment
    steps:
      - command: "npm install"  # Install dependencies
        description: "Install project dependencies"
```

### 2. Document Your Workflows

Always include descriptions and comments:

```yaml
workflows:
  deploy:
    # Deploy the application to the specified environment
    # Requires: API_KEY, DEPLOYMENT_TARGET variables
    description: "Production deployment workflow"
    # ... workflow steps
```

### 3. Use Variable Validation

Validate critical variables in pre-checks:

```yaml
pre_checks:
  - command: "[ -n '{{api_key}}' ] || (echo 'API_KEY required' && exit 1)"
    description: "Validate API key is provided"
```

### 4. Organize by Functionality

Group related workflows together:

```yaml
# migraine.yml
workflows:
  # Development workflows
  setup-dev: {}
  run-dev-server: {}
  
  # Testing workflows  
  run-unit-tests: {}
  run-integration-tests: {}
  
  # Deployment workflows
  deploy-staging: {}
  deploy-production: {}
```

### 5. Use Environment Files for Sensitive Data

Never commit secrets to configuration files:

```yaml
# migraine.yml - OK
use_vault: true
steps:
  - command: "curl -H 'Authorization: Bearer {{api_token}}' {{api_url}}"

# .env (add to .gitignore) - OK  
API_TOKEN=secret_token_here

# migraine.yml with hardcoded secrets - NOT OK
# steps:
#   - command: "curl -H 'Authorization: Bearer secret123' https://api.com"
```

## Migration from Legacy Configurations

If you're upgrading from an older version of Migraine:

1. **Automatic migration** will convert existing Badger-stored workflows to the new format
2. **New workflows** should use YAML format in the `./workflows/` directory
3. **Project-level configuration** can be stored in `migraine.yml`
4. **Environment variables** should be managed in `.env` files or the vault system

## Troubleshooting Configuration Files

### Common Issues

1. **Workflow not found**: Check file extension and location
2. **Variables not resolving**: Verify vault configuration and variable scopes
3. **Pre-check failures**: Examine the specific command that failed
4. **YAML syntax errors**: Use a YAML validator to check syntax

### Debugging Tips

1. **Validate YAML syntax**:
   ```bash
   # Use yamllint to validate syntax
   yamllint migraine.yml
   ```

2. **Check workflow information**:
   ```bash
   migraine workflow info workflow-name
   ```

3. **List all discovered workflows**:
   ```bash
   migraine workflow list
   ```

4. **Test variable resolution**:
   ```bash
   migraine vars list
   ```

## Examples

### Full Stack Application Configuration

```yaml
# migraine.yml for a full-stack application
project:
  name: "Full-Stack App"
  description: "A complete web application with frontend and backend"

defaults:
  use_vault: true

workflows:
  setup:
    description: "Complete project setup"
    pre_checks:
      - command: "which node"
        description: "Verify Node.js is available"
      - command: "which docker"
        description: "Verify Docker is available"
    steps:
      - command: "npm install"
        description: "Install root dependencies"
      - command: "cd frontend && npm install"
        description: "Install frontend dependencies"
      - command: "cd backend && npm install"
        description: "Install backend dependencies"
      - command: "docker-compose up -d"
        description: "Start database and other services"

  dev:
    description: "Start development servers"
    steps:
      - command: "cd frontend && npm run dev"
        description: "Start frontend development server"
      - command: "cd backend && npm run dev"
        description: "Start backend development server"

  test:
    description: "Run complete test suite"
    steps:
      - command: "npm run test:frontend"
        description: "Run frontend tests"
      - command: "npm run test:backend"
        description: "Run backend tests"

  build:
    description: "Build the complete application"
    steps:
      - command: "cd frontend && npm run build"
        description: "Build frontend"
      - command: "cd backend && npm run build"
        description: "Build backend"

environments:
  development:
    variables:
      NODE_ENV: "development"
      API_URL: "http://localhost:3001"
  production:
    variables:
      NODE_ENV: "production" 
      API_URL: "https://api.yourapp.com"
```