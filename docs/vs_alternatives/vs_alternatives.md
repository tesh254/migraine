# Migraine vs. Alternative Tools: Why Choose Migraine?

This document compares Migraine to other popular automation and task runner tools, highlighting the advantages that make Migraine a superior choice for workflow automation.

## Overview

Task automation is crucial for modern development workflows. While there are many tools available, each has its own strengths and weaknesses. This comparison will help you understand why Migraine may be the best choice for your automation needs.

## Migraine vs. Shell Scripts

### Traditional Shell Scripts

Many developers create custom shell scripts for automation tasks. While functional, this approach has several limitations:

#### Limitations of Shell Scripts
- **Hard to maintain**: Scripts become complex and difficult to update
- **Poor error handling**: Basic error handling requires significant extra code
- **Limited documentation**: No built-in way to document steps and purpose
- **Poor variable management**: Manual parsing of arguments and environment variables
- **Inconsistent structure**: Each script has different conventions
- **No pre-checks**: Validation must be manually implemented
- **Hard to compose**: Difficult to chain or combine scripts

#### Advantages of Migraine over Shell Scripts
- **Structured format**: YAML provides consistent, readable structure
- **Built-in validation**: Pre-checks section for automatic validation
- **Rich documentation**: Comments and descriptions are first-class features
- **Advanced variable management**: Vault system with scope awareness
- **Error handling**: Automatic failure handling and clear error messages
- **Composability**: Easy to run specific actions or steps
- **Automatic discovery**: Finds workflow files automatically
- **Standardized execution**: Consistent command interface

### Example Comparison

#### Shell Script Approach
```bash
#!/bin/bash

# Deploy script
echo "Starting deployment process..."

# Validate inputs
if [ -z "$SERVER" ]; then
    echo "Error: SERVER environment variable required" >&2
    exit 1
fi

if [ -z "$APP_NAME" ]; then
    echo "Error: APP_NAME environment variable required" >&2
    exit 1
fi

# Check if required tools exist
if ! command -v rsync &> /dev/null; then
    echo "Error: rsync is required but not found" >&2
    exit 1
fi

if ! command -v ssh &> /dev/null; then
    echo "Error: ssh is required but not found" >&2
    exit 1
fi

# Test SSH connectivity
if ! ssh -q "$SERVER" exit; then
    echo "Error: Cannot connect to server $SERVER" >&2
    exit 1
fi

# Build the application
echo "Building application..."
npm run build

# Package for deployment
echo "Packaging application..."
tar -czf app.tar.gz dist/

# Deploy
echo "Deploying to $SERVER..."
rsync -av app.tar.gz "$SERVER:/tmp/"

echo "Deployment complete!"
```

#### Migraine YAML Approach
```yaml
name: deploy-app
description: Deploy application to server with full validation

pre_checks:
  - command: "which rsync"
    description: "Verify rsync is available"
  - command: "which ssh"
    description: "Verify ssh is available"
  - command: "ssh -q {{server}} exit"
    description: "Test server connectivity"
  - command: "[ -n '{{server}}' ] || (echo 'SERVER required' && exit 1)"
    description: "Validate server variable"
  - command: "[ -n '{{app_name}}' ] || (echo 'APP_NAME required' && exit 1)"
    description: "Validate app_name variable"

steps:
  - command: "echo 'Starting deployment process...'"
    description: "Show start message"
  - command: "npm run build"
    description: "Build the application"
  - command: "tar -czf app.tar.gz dist/"
    description: "Package for deployment"
  - command: "rsync -av app.tar.gz {{user}}@{{server}}:/tmp/"
    description: "Deploy package to server"

actions:
  rollback:
    command: "ssh {{user}}@{{server}} 'cd /opt/app && git checkout HEAD~1'"
    description: "Rollback to previous version"
  cleanup:
    command: "rm -f app.tar.gz"
    description: "Clean up local package"

config:
  variables:
    server:
      - "required"
    user:
      - "required"
  store_variables: false

use_vault: true
```

## Migraine vs. Makefiles

### Makefiles

Make is a powerful build automation tool, but has specific limitations for general workflow automation:

#### Limitations of Makefiles
- **Unix-focused**: Limited Windows support and compatibility
- **Complex syntax**: Non-intuitive syntax for non-build tasks
- **Dependency-based**: Designed for file generation, not general workflows
- **Limited variable management**: Basic variable substitution only
- **Poor cross-platform**: Challenging to write portable Makefiles
- **No built-in validation**: Pre-checks must be manually implemented

#### Advantages of Migraine over Makefiles
- **YAML syntax**: Modern, readable configuration
- **Cross-platform**: Works on macOS, Linux, and Windows
- **Workflow-focused**: Designed specifically for workflow automation
- **Advanced variable management**: Vault system with scope awareness
- **Built-in pre-checks**: Dedicated validation section
- **Rich documentation**: Descriptions and comments as first-class features
- **Automatic discovery**: Finds workflows automatically
- **Modern CLI**: Intuitive command interface

### Example Comparison

#### Makefile Approach
```makefile
# Deploy targets
.PHONY: deploy precheck build

# Variables
SERVER ?= localhost
APP_NAME ?= myapp

precheck:
	@echo "Checking prerequisites..."
	@which rsync || (echo "rsync not found" && exit 1)
	@ssh -q $(SERVER) exit || (echo "Cannot connect to server" && exit 1)
	@test -n "$(SERVER)" || (echo "SERVER required" && exit 1)

build: precheck
	@echo "Building application..."
	npm run build

deploy: build
	@echo "Deploying to $(SERVER)..."
	tar -czf app.tar.gz dist/
	rsync -av app.tar.gz $(SERVER):/tmp/

rollback:
	@echo "Rolling back..."
	ssh $(SERVER) 'cd /opt/app && git checkout HEAD~1'
```

#### Migraine Approach
```yaml
name: deploy
description: Deploy application with validation and rollback

pre_checks:
  - command: "which rsync"
    description: "Verify rsync is available"
  - command: "ssh -q {{server}} exit"
    description: "Test server connectivity"
  - command: "[ -n '{{server}}' ] || (echo 'SERVER required' && exit 1)"
    description: "Validate server variable"

steps:
  - command: "echo 'Building application...'"
    description: "Show build message"
  - command: "npm run build"
    description: "Build the application"
  - command: "tar -czf app.tar.gz dist/"
    description: "Package for deployment"
  - command: "rsync -av app.tar.gz {{server}}:/tmp/"
    description: "Deploy package"

actions:
  deploy:
    command: "ssh {{server}} 'tar -xzf /tmp/app.tar.gz -C /opt/app && systemctl restart {{app_name}}'"
    description: "Complete deployment"
  rollback:
    command: "ssh {{server}} 'cd /opt/app && git checkout HEAD~1'"
    description: "Rollback to previous version"
  cleanup:
    command: "rm -f app.tar.gz && ssh {{server}} 'rm -f /tmp/app.tar.gz'"
    description: "Clean up files"

variables:
  - server
  - app_name

use_vault: true
```

## Migraine vs. npm Scripts

### npm Scripts

npm scripts are popular in Node.js projects but have limitations outside the JS ecosystem:

#### Limitations of npm Scripts
- **Language-specific**: Designed for the Node.js ecosystem only
- **Poor variable management**: Only environment variables
- **No workflow organization**: All scripts in single package.json
- **Limited validation**: No built-in pre-checks
- **No scope awareness**: All variables global
- **Complex chaining**: Difficult to create complex workflows

#### Advantages of Migraine over npm Scripts
- **Language agnostic**: Works with any technology stack
- **Advanced variable management**: Vault with global/project/workflow scopes
- **Organized workflows**: Multiple files, clear separation
- **Built-in validation**: Dedicated pre-checks section
- **Rich documentation**: Descriptions and comments
- **Better composition**: Actions and workflow dependencies

### Example Comparison

#### npm Scripts Approach
```json
{
  "scripts": {
    "build": "npm run build:frontend && npm run build:backend",
    "build:frontend": "cd frontend && npm run build",
    "build:backend": "cd backend && npm run build",
    "test": "npm run test:unit && npm run test:integration",
    "test:unit": "jest",
    "test:integration": "cypress run",
    "deploy": "npm run build && npm run deploy:server && npm run notify:slack",
    "deploy:server": "rsync -av ./dist/ user@server:/var/www/",
    "notify:slack": "curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"Deployment successful\"}' $SLACK_WEBHOOK"
  }
}
```

#### Migraine Approach
```yaml
name: project-workflows
description: Complete project automation

pre_checks:
  - command: "which npm"
    description: "Verify npm is available"
  - command: "which rsync"
    description: "Verify rsync is available"

workflows:
  build:
    description: Build the entire project
    steps:
      - command: "echo 'Building frontend...'"
        description: "Frontend build start"
      - command: "cd frontend && npm run build"
        description: "Build frontend application"
      - command: "echo 'Building backend...'"
        description: "Backend build start"
      - command: "cd backend && npm run build"
        description: "Build backend application"

  test:
    description: Run complete test suite
    steps:
      - command: "npm run test:unit"
        description: "Run unit tests"
      - command: "npm run test:integration"
        description: "Run integration tests"

  deploy:
    description: Deploy to production
    pre_checks:
      - command: "npm run test"
        description: "Run tests before deployment"
    steps:
      - command: "npm run build"
        description: "Build application"
      - command: "npm run deploy:server"
        description: "Deploy to server"
      - command: "npm run notify:slack"
        description: "Send notification"

  deploy-server:
    description: Deploy files to server
    steps:
      - command: "rsync -av ./dist/ {{deploy_user}}@{{deploy_server}}:/var/www/"
        description: "Sync files to server"

  notify-slack:
    description: Send deployment notification
    steps:
      - command: "curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"Deployment successful\"}' {{slack_webhook}}"
        description: "Send Slack notification"

variables:
  - deploy_server
  - deploy_user
  - slack_webhook

use_vault: true
```

## Migraine vs. Taskfile (Task)

### Taskfiles

Task is a modern task runner with TOML syntax, but still has some limitations:

#### Limitations of Taskfiles
- **TOML syntax**: Less familiar than YAML for many users
- **Limited variable scoping**: No built-in vault system
- **No automatic discovery**: Must specify exact task location
- **Less workflow-focused**: More oriented toward build tasks
- **Smaller ecosystem**: Less tooling and editor support

#### Advantages of Migraine over Taskfiles
- **YAML syntax**: More familiar and better tooling support
- **Advanced variable management**: Built-in vault with scopes
- **Automatic discovery**: Finds workflows automatically
- **Workflow-centric**: Designed specifically for complex workflows
- **Rich documentation**: First-class support for descriptions
- **Pre-checks**: Dedicated validation section

## Migraine vs. Rake

### Rakefiles

Rake is Ruby's make tool, but has Ruby-specific requirements:

#### Limitations of Rakefiles
- **Ruby dependency**: Requires Ruby installation
- **Limited scope**: Primarily for Ruby projects
- **Complex syntax**: Ruby code mixed with build logic
- **Platform issues**: Ruby compatibility across platforms
- **Learning curve**: Requires Ruby knowledge

#### Advantages of Migraine over Rakefiles
- **No language dependency**: Standalone binary
- **Cross-language**: Works with any technology stack
- **Simple syntax**: YAML configuration
- **No installation**: Single binary executable
- **Go-based**: Consistent across platforms

## Why Choose Migraine?

### 1. Modern, Readable Configuration
- YAML syntax is more readable than Make, TOML, or Ruby syntax
- Rich commenting and documentation capabilities
- Clear separation of concerns (pre-checks, steps, actions)

### 2. Advanced Variable Management
- Vault system with global/project/workflow scoping
- Automatic variable resolution with fallbacks
- Secure storage system (with upcoming encryption)

### 3. Built-in Validation
- Dedicated pre-checks section for validation
- Automatic workflow validation
- Clear error reporting

### 4. Automatic Discovery
- Finds workflow files automatically
- No need to specify exact paths
- Supports multiple workflow file locations

### 5. Cross-Platform Consistency
- Built with Go for consistent behavior across platforms
- Single binary executable
- No runtime dependencies

### 6. Comprehensive CLI
- Intuitive command structure
- Rich help and documentation
- Consistent interface across operations

### 7. Workflow-Centric Design
- Specifically designed for complex workflows
- Supports long-running and multi-step processes
- Action-based execution for flexibility

### 8. Professional Features
- Migration from legacy formats
- File-based and database storage
- Environment-specific configurations
- Variable transformation capabilities

## Conclusion

While shell scripts, Make, npm scripts, Task, and Rake all have their place, Migraine is specifically designed for complex workflow automation with features like:
- Structured, readable configuration
- Advanced variable management with vault
- Built-in validation and error handling
- Professional workflow organization
- Cross-platform consistency
- Modern CLI experience

For teams looking to manage complex automation workflows with professional tooling, clear documentation, and robust variable management, Migraine provides significant advantages over traditional approaches.