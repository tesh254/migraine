# Best Practices and Advanced Features

## Best Practices

### 1. Workflow Organization

#### Use Descriptive Names
```yaml
# Good
name: deploy-to-production

# Avoid
name: deploy1
name: workflow_2
```

#### Group Related Workflows
Organize your `./workflows/` directory with clear naming:
```
workflows/
├── build/
│   ├── frontend.yaml
│   └── backend.yaml
├── deploy/
│   ├── staging.yaml
│   └── production.yaml
├── test/
│   ├── unit.yaml
│   └── integration.yaml
└── setup.yaml
```

#### Single Responsibility Principle
Each workflow should focus on one main task:

```yaml
# Good: Focused workflow
name: build-frontend
description: Build the frontend application

steps:
  - command: "npm install"
    description: "Install dependencies"
  - command: "npm run build"
    description: "Build the frontend"
```

### 2. Documentation and Comments

#### Describe Each Step
```yaml
steps:
  - command: "npm install"
    description: "Install Node.js dependencies from package-lock.json"
  
  - command: "npm run build"
    description: "Compile TypeScript and bundle assets to dist/ directory"
  
  - command: "npm run test"
    description: "Run unit tests to verify build integrity"
```

#### Use Rich Comments in YAML Workflows
```yaml
# Database migration workflow
# This workflow should be run before deploying database schema changes
# Prerequisites:
# - Database backup has been created
# - Maintenance window established
# - API traffic is routed to standby service
name: db-migrate
description: Apply database migrations with safety checks
```

### 3. Variable Management

#### Use Variable Scoping Appropriately
```bash
# Global variables (available to all workflows)
migraine vars set default_region "us-west-2" -s global

# Project variables (available to all workflows in this project)
migraine vars set project_id "my-project-123" -s project

# Workflow-specific variables
migraine vars set db_password "secret" -s workflow -w db-deploy
```

#### Validate Required Variables
```yaml
pre_checks:
  # Validate that required variables are provided
  - command: "[ -n '{{api_key}}' ] || (echo 'API_KEY is required' && exit 1)"
    description: "Validate API key is provided"
  - command: "[ -n '{{environment}}' ] || (echo 'ENVIRONMENT is required' && exit 1)"
    description: "Validate environment is specified"
```

### 4. Error Handling and Idempotency

#### Write Idempotent Commands
```yaml
# Good: Command can be safely rerun
steps:
  - command: "mkdir -p ./build && chmod 755 ./build"
    description: "Ensure build directory exists with proper permissions"

# Avoid: Command that fails on second run
steps:
  - command: "mkdir ./build"  # Fails if directory already exists
    description: "Create build directory"
```

#### Use Conditional Logic
```yaml
steps:
  # Only run cleanup if previous step created files
  - command: "if [ -d './temp' ]; then rm -rf ./temp; fi"
    description: "Clean up temporary directory if it exists"
```

### 5. Security Considerations

#### Never Hardcode Secrets
```yaml
# Bad: Never do this
steps:
  - command: "curl -H 'Authorization: Bearer secret123' https://api.com"

# Good: Use vault variables
steps:
  - command: "curl -H 'Authorization: Bearer {{api_token}}' https://api.com"
    description: "Call API with vault-stored token"
```

#### Validate Inputs
```yaml
pre_checks:
  - command: "echo '{{user_input}}' | grep -E '^[a-zA-Z0-9_-]+$' || (echo 'Invalid input: only alphanumeric, underscore, hyphen allowed' && exit 1)"
    description: "Validate user input format"
```

## Advanced Features

### 1. Environment-Specific Configurations

Create a `migraine.yml` file at your project root:

```yaml
# migraine.yml
project:
  name: "My Application"
  version: "1.0.0"

environments:
  development:
    variables:
      NODE_ENV: "development"
      API_URL: "http://localhost:3000"
      DEBUG: "true"
  staging:
    variables:
      NODE_ENV: "staging" 
      API_URL: "https://staging-api.example.com"
      DEBUG: "false"
  production:
    variables:
      NODE_ENV: "production"
      API_URL: "https://api.example.com"
      DEBUG: "false"

defaults:
  use_vault: true
```

### 2. Complex Pre-checks

Use pre-checks for comprehensive validation:

```yaml
name: deploy-to-production
description: Production deployment with extensive validation

pre_checks:
  # System requirements
  - command: "which docker && which docker-compose"
    description: "Verify Docker tools are available"
  
  # Service health checks
  - command: "curl -f http://localhost:3000/health || exit 1"
    description: "Verify current service is healthy"
  
  # Configuration validation
  - command: "test -f .env.production && test -f docker-compose.prod.yml"
    description: "Verify deployment configuration exists"
  
  # Permissions check
  - command: "test -w ./logs && test -w ./data"
    description: "Verify write permissions to critical directories"
  
  # Backup verification
  - command: "find ./backups -name 'backup_*.tar.gz' -mmin -60 | head -1"
    description: "Verify recent backup exists (within last hour)"
```

### 3. Advanced Variable Transformations

Use the vault system with transformations:

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

### 4. Action-Based Workflows

Create workflows with multiple actions for different operations:

```yaml
name: app-management
description: Complete application lifecycle management

steps:
  # Default: show status
  - command: "docker-compose ps"
    description: "Show current service status"

actions:
  start:
    command: "docker-compose up -d"
    description: "Start all services"
  
  stop:
    command: "docker-compose down"
    description: "Stop all services"
  
  restart:
    command: "docker-compose restart"
    description: "Restart all services"
  
  logs:
    command: "docker-compose logs -f"
    description: "Follow service logs"
  
  backup:
    command: "scripts/backup.sh"
    description: "Create database backup"
  
  rollback:
    command: "scripts/rollback.sh {{version}}"
    description: "Rollback to specific version"
  
  cleanup:
    command: "docker system prune -f"
    description: "Clean up unused Docker resources"
```

Run specific actions:
```bash
migraine workflow run app-management -a start
migraine workflow run app-management -a logs
migraine workflow run app-management -a rollback -v version=v1.2.3
```

### 5. Conditional Execution Patterns

While Migraine doesn't have built-in conditionals, you can implement them:

```yaml
name: conditional-build
description: Build with environment-specific steps

steps:
  - command: "echo 'Starting build for {{environment}}...'"
    description: "Show build environment"
  
  # Environment-specific steps
  - command: "if [ '{{environment}}' = 'production' ]; then npm run build:prod; else npm run build:dev; fi"
    description: "Build with environment-specific configuration"
  
  # Debug-specific steps
  - command: "if [ '{{debug}}' = 'true' ]; then npm run analyze-bundle; fi"
    description: "Analyze bundle if debug mode enabled"
  
  # Run tests
  - command: "npm run test"
    description: "Run tests"
  
  # Deploy conditionally
  - command: "if [ '{{environment}}' = 'production' ]; then ./deploy.sh; fi"
    description: "Deploy only in production environment"
```

### 6. Workflow Composition

Create workflows that build upon each other:

```yaml
# build.yaml
name: build
description: Build the application
steps:
  - command: "npm install"
    description: "Install dependencies"
  - command: "npm run build"
    description: "Build application"

# test.yaml
name: test
description: Test the application
pre_checks:
  - command: "migraine workflow validate workflows/build.yaml"
    description: "Ensure build workflow is available"
steps:
  - command: "migraine workflow run build"  # Run another workflow
    description: "Build before testing"
  - command: "npm run test"
    description: "Run tests"

# deploy.yaml
name: deploy
description: Deploy the application
pre_checks:
  - command: "migraine workflow run test"  # Ensure tests pass
    description: "Run tests before deployment"
steps:
  - command: "migraine workflow run build"  # Build if not done
    description: "Build application"
  - command: "rsync -av ./dist/ {{server}}:/var/www/"
    description: "Deploy to server"
```

### 7. Advanced Vault Usage

#### Workflow with Multiple Variable Sources

```yaml
name: complex-deployment
description: Deployment with multiple variable sources

# This workflow uses vault, env files, and command-line variables
use_vault: true  # Will look up variables in vault first

# Variables can come from:
# 1. Command-line: -v server=prod -v user=admin
# 2. Vault: api_key, db_password
# 3. Env files: .env, .env.production
# 4. Prompts: if not found elsewhere

steps:
  - command: "echo 'Deploying to {{server}} as {{user}} at $(date)'"
    description: "Log deployment info"
  - command: "rsync -av ./dist/ {{user}}@{{server}}:{{deploy_path}}"
    description: "Deploy files"
  - command: "ssh {{user}}@{{server}} 'cd {{deploy_path}} && ./post-deploy.sh'"
    description: "Run post-deploy script"

actions:
  rollback:
    command: "ssh {{user}}@{{server}} 'cd {{deploy_path}} && ./rollback.sh {{previous_version}}'"
    description: "Rollback to previous version"
    variables:
      - "previous_version"  # Will be prompted if not provided
```

### 8. Monitoring and Logging

Add monitoring to your workflows:

```yaml
hooks:
  pre_workflow:
    - command: "echo '[INFO] Starting workflow: {{workflow_name}} at $(date)' >> ./logs/workflow.log"
      description: "Log workflow start"
  
  post_workflow:
    - command: "echo '[INFO] Completed workflow: {{workflow_name}} at $(date)' >> ./logs/workflow.log"
      description: "Log workflow completion"
    - command: "if [ $? -eq 0 ]; then echo '[SUCCESS] {{workflow_name}}' >> ./logs/success.log; else echo '[ERROR] {{workflow_name}}' >> ./logs/error.log; fi"
      description: "Log success or failure"

# Or implement in workflow steps:
steps:
  - command: "echo 'Workflow started: $(date)' > /tmp/mgr-{{workflow_name}}-start.log"
    description: "Create timestamp file"
  # ... other steps ...
  - command: "echo 'Workflow completed: $(date)' >> /tmp/mgr-{{workflow_name}}-start.log"
    description: "Add completion timestamp"
```

## Performance Optimization

### 1. Parallel Execution Considerations

While Migraine doesn't yet support parallel workflow execution, design your workflows to be efficient:

```yaml
# Group related commands to reduce overhead
steps:
  - command: |
      npm install
      npm run build
      npm run test
    description: "Install, build, and test in one shell session"
```

### 2. Caching and Optimization

Use caching where possible:

```yaml
steps:
  # Check if build already exists to skip unnecessary work
  - command: "if [ ! -f './dist/app.js' ] || [ ./src -nt ./dist/app.js ]; then npm run build; else echo 'Build up to date'; fi"
    description: "Build only if source changed"
```

## Troubleshooting Advanced Issues

### 1. Debugging Variable Resolution

Check variable resolution order:
```bash
# See what variables are available
migraine vars list

# Test specific variable
migraine vars get my_var

# Run with verbose output (if available)
migraine workflow run my-workflow --verbose
```

### 2. Workflow Dependencies

For complex workflows, consider creating dependency validation:

```yaml
pre_checks:
  # Check if previous workflow results exist
  - command: "test -f ./build/dist.tar.gz || (echo 'Build artifact not found, run build workflow first' && exit 1)"
    description: "Verify build artifacts exist"
```

### 3. Environment Isolation

Test workflows in isolated environments:

```yaml
# Use temporary directories for testing
steps:
  - command: |
      TEMP_DIR=$(mktemp -d)
      cd $TEMP_DIR
      # Your commands here
      rm -rf $TEMP_DIR
    description: "Run in temporary isolated environment"
```

## Migration and Integration Tips

### From Legacy Systems

When migrating from makefiles or scripts:
1. Start by converting your most critical workflows first
2. Use the same variable names for consistency
3. Add pre-checks to validate the migration
4. Test thoroughly in a development environment
5. Gradually migrate less critical workflows

### Integration with CI/CD

```yaml
# For CI/CD systems, ensure workflows are self-contained:
name: ci-build
description: CI build for automated systems

pre_checks:
  - command: "[ -n '${CI}' ] || (echo 'This should run in CI only' && exit 1)"
    description: "Verify running in CI environment"

steps:
  - command: "npm ci"  # Use ci instead of install for clean builds
    description: "Install clean dependencies"
  - command: "npm run test:ci"  # Use CI-specific test command
    description: "Run CI tests"
  - command: "npm run build"
    description: "Build application"
  - command: "echo \"Build completed at $(date)\""
    description: "Log build completion"
```

These best practices and advanced features will help you create robust, maintainable, and efficient workflows with Migraine.