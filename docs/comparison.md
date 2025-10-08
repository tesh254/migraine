# Comparison: Migraine vs Other Task Automation Tools

## Overview

Migraine is a modern task automation tool designed for workflow orchestration. This document compares Migraine to other popular tools like Make, npm scripts, Taskfile, and Rake, highlighting the advantages and use cases for each.

## Migraine vs Make

### Feature Comparison

| Feature | Migraine | Make |
|---------|----------|------|
| **Configuration Format** | YAML with rich comments | Makefile syntax (POSIX shell) |
| **Variable Management** | Advanced vault system with scopes | Shell variables and environment |
| **Variable Substitution** | `{{variable_name}}` with validation | `$(VARIABLE)` or `${VARIABLE}` |
| **Dependency Management** | Pre-checks and actions | Explicit target dependencies |
| **Documentation** | Rich YAML comments and descriptions | Comment lines (not standardized) |
| **Cross-platform** | Yes (Go-based) | Unix-focused, limited Windows support |
| **Learning Curve** | Low (YAML syntax) | Moderate (Make-specific syntax) |
| **Concurrency** | Built-in support | Limited (Make -j) |
| **File Discovery** | Automatic in `./workflows/` | Manual target specification |
| **Execution Context** | Isolated shell commands | Same shell context |
| **Variable Scoping** | Global/Project/Workflow | Project-wide only |
| **IDE Support** | YAML with standard tooling | Specialized Makefile editors |

### Code Examples

#### Makefile Example
```makefile
# Build the application
build:
	@echo "Building the application..."
	npm install
	npm run build

# Run tests
test:
	@echo "Running tests..."
	npm run test:unit
	npm run test:integration

# Deploy to production
deploy: build test
	@echo "Deploying to production..."
	@if [ -z "${API_KEY}" ]; then echo "API_KEY required"; exit 1; fi
	scp ./dist/* user@server:/var/www/

# Setup development environment
setup:
	@echo "Setting up development environment..."
	npm install
	cp .env.example .env

.PHONY: build test deploy setup
```

#### Migraine YAML Example
```yaml
# workflows/deploy.yaml
name: deploy
description: Deploy the application to production servers

pre_checks:
  - command: "which npm"
    description: "Verify npm is available"
  - command: "[ -n '{{api_key}}' ] || (echo 'API_KEY required' && exit 1)"
    description: "Validate API key is provided"

steps:
  - command: "echo 'Building the application...'"
    description: "Show build start message"
  - command: "npm install"
    description: "Install dependencies"
  - command: "npm run build"
    description: "Build the application"
  - command: "npm run test"
    description: "Run tests before deployment"

actions:
  deploy:
    command: "scp ./dist/* {{deploy_user}}@{{deploy_server}}:/var/www/"
    description: "Deploy to production server"
  rollback:
    command: "ssh {{deploy_user}}@{{deploy_server}} 'cd /var/www && git checkout HEAD~1'"
    description: "Rollback to previous version"

config:
  variables:
    api_key:
      - "required"
  store_variables: false

use_vault: true
```

### When to Choose Make

**Choose Make when:**
- Building C/C++ projects or similar compiled languages
- Need sophisticated dependency tracking between files
- Working in Unix-only environments
- Want minimal tooling dependencies (make is standard)
- Need file generation rules (`%.o: %.c` pattern)
- Working with complex build systems that require dependency graphs

**Choose Migraine when:**
- Need modern, readable configuration (YAML)
- Want advanced variable management with vault system
- Working with cross-platform projects
- Need automatic workflow discovery
- Want rich documentation capabilities
- Need scope-aware variables (global/project/workflow)
- Want a modern CLI experience

## Migraine vs npm Scripts

### Feature Comparison

| Feature | Migraine | npm Scripts |
|---------|----------|-------------|
| **Ecosystem** | Cross-language | JavaScript/Node.js only |
| **Configuration** | YAML (standalone) | package.json scripts |
| **Variable Management** | Advanced vault system | Environment variables only |
| **Workflow Organization** | Multi-file, categorized | Single package.json |
| **Cross-project** | Yes | Limited |
| **Pre-checks** | Built-in section | Manual implementation |
| **Actions** | Separate actions section | Combined in scripts |
| **Documentation** | Rich comments in YAML | Package.json description only |
| **Execution** | `migraine workflow run name` | `npm run name` |
| **Variable Scoping** | Global/Project/Workflow | Environment only |
| **Language Agnostic** | Yes | Node.js ecosystem |

### Code Examples

#### Package.json Scripts Example
```json
{
  "scripts": {
    "setup": "npm install && cp .env.example .env",
    "build": "npm run build:frontend && npm run build:backend",
    "build:frontend": "cd frontend && npm run build",
    "build:backend": "cd backend && npm run build", 
    "test": "npm run test:unit && npm run test:e2e",
    "test:unit": "jest",
    "test:e2e": "cypress run",
    "deploy": "npm run build && npm run deploy:server",
    "deploy:server": "rsync -av ./dist/ user@server:/var/www/"
  }
}
```

#### Migraine Equivalent
```yaml
# workflows/project-workflows.yaml
name: project-workflows
description: Complete project automation workflows

workflows:
  setup:
    description: Set up development environment
    steps:
      - command: "npm install"
        description: "Install dependencies"
      - command: "cp .env.example .env" 
        description: "Create environment file"

  build:
    description: Build the entire project
    steps:
      - command: "npm run build:frontend"
        description: "Build frontend"
      - command: "npm run build:backend" 
        description: "Build backend"

  build-frontend:
    description: Build frontend only
    steps:
      - command: "cd frontend && npm run build"
        description: "Build frontend application"

  build-backend: 
    description: Build backend only
    steps:
      - command: "cd backend && npm run build"
        description: "Build backend application"

  test:
    description: Run complete test suite
    steps:
      - command: "npm run test:unit"
        description: "Run unit tests"
      - command: "npm run test:e2e"
        description: "Run end-to-end tests"

  deploy:
    description: Deploy to production
    pre_checks:
      - command: "npm run build"
        description: "Verify build succeeds"
    steps:
      - command: "npm run deploy:server"
        description: "Deploy to server"

  deploy-server:
    description: Deploy files to server
    steps:
      - command: "rsync -av ./dist/ {{deploy_user}}@{{deploy_server}}:/var/www/"
        description: "Sync files to server"
    variables:
      - "deploy_server"
      - "deploy_user"
```

### When to Choose npm Scripts

**Choose npm Scripts when:**
- Working in a Node.js-only project
- Want integration with package management
- Need simple task execution
- Already using Node.js tooling
- Want to keep everything in package.json

**Choose Migraine when:**
- Working with multi-language projects
- Need sophisticated variable management
- Want automatic workflow discovery
- Need cross-project workflow sharing
- Want better documentation capabilities
- Need pre-checks and actions separation

## Migraine vs Taskfile (Task)

### Feature Comparison

| Feature | Migraine | Taskfile |
|---------|----------|----------|
| **Configuration Format** | YAML | YAML |
| **Variable Management** | Vault system with scopes | Environment and command-line |
| **Pre-checks** | Built-in section | Manual implementation |
| **Actions** | Separate actions section | Same as tasks |
| **Cross-platform** | Excellent | Excellent |
| **Dependency Management** | Pre-checks concept | Explicit dependencies |
| **File Discovery** | Automatic in `./workflows/` | Single Taskfile |
| **Task Organization** | Multi-file or single file | Single Taskfile |
| **Documentation** | Rich comments in YAML | Comments in Taskfile |
| **Language Agnostic** | Yes | Yes |
| **Execution Context** | Isolated commands | Configurable |

### Code Examples

#### Taskfile Example
```yaml
# Taskfile.yml
version: '3'

vars:
  BINARY_NAME: 'myapp'
  BUILD_DIR: './build'

tasks:
  build:
    desc: Build the application
    deps: [install]
    cmds:
      - echo "Building {{.BINARY_NAME}}..."
      - go build -o {{.BUILD_DIR}}/{{.BINARY_NAME}} .
  
  install:
    desc: Install dependencies
    cmds:
      - go mod tidy
      - go mod download

  test:
    desc: Run tests
    deps: [install]
    cmds:
      - go test ./...
  
  run:
    desc: Run the application  
    deps: [build]
    cmds:
      - {{.BUILD_DIR}}/{{.BINARY_NAME}}
    
  deploy:
    desc: Deploy the application
    deps: [build, test]
    cmds:
      - ./scripts/deploy.sh
    vars:
      TARGET: '{{.TARGET | default "staging"}}'
```

#### Migraine Equivalent
```yaml
# workflows/app-workflows.yaml
name: app-workflows
description: Application development and deployment workflows

pre_checks:
  - command: "which go"
    description: "Verify Go is installed"

steps:
  build:
    description: Build the application
    commands:
      - command: "echo 'Building {{binary_name}}...'"
        description: "Show build message"
      - command: "go build -o {{build_dir}}/{{binary_name}} ."
        description: "Build the Go application"

  install:
    description: Install dependencies
    commands:
      - command: "go mod tidy"
        description: "Tidy Go modules"
      - command: "go mod download" 
        description: "Download dependencies"

  test:
    description: Run tests
    commands:
      - command: "go test ./..."
        description: "Run Go tests"

  run:
    description: Run the application
    commands:
      - command: "{{build_dir}}/{{binary_name}}"
        description: "Execute the binary"

actions:
  deploy:
    command: "./scripts/deploy.sh"
    description: "Deploy the application"
    variables:
      - "target"

config:
  variables:
    binary_name: "myapp"
    build_dir: "./build"
  store_variables: false

use_vault: true
```

## Migraine vs Rake

### Feature Comparison

| Feature | Migraine | Rake |
|---------|----------|------|
| **Language** | Go (compiled) | Ruby (interpreted) |
| **Configuration** | YAML | Ruby code |
| **Variable Management** | Vault system | Ruby variables/methods |
| **Cross-platform** | Excellent | Dependent on Ruby |
| **Learning Curve** | Low (YAML) | Medium (Ruby knowledge needed) |
| **Performance** | Fast (compiled) | Moderate (interpreted) |
| **Dependency Management** | Pre-checks concept | Ruby dependency system |
| **Documentation** | YAML comments | RDoc comments |
| **IDE Support** | YAML editors | Ruby-specific editors |
| **File Discovery** | Automatic | Rakefile only |

## Pros and Cons Summary

### Migraine Advantages
✅ **Modern configuration**: YAML with rich comments and structure  
✅ **Advanced variable management**: Vault system with scope awareness  
✅ **Automatic discovery**: Finds workflows in `./workflows/` directory  
✅ **Cross-platform**: Go-based, runs anywhere  
✅ **Rich documentation**: Built-in workflow info and descriptions  
✅ **Flexible execution**: Multiple variable resolution methods  
✅ **Pre-checks concept**: Validation before execution  
✅ **Action separation**: Optional steps that can run independently  

### Migraine Disadvantages  
❌ **Additional dependency**: Requires separate tool installation  
❌ **Less mature**: Newer tool with fewer plugins/ecosystem  
❌ **YAML parsing**: Possible indentation issues  

### Make Advantages
✅ **Ubiquitous**: Available on most Unix systems  
✅ **Powerful dependency system**: Sophisticated file dependency tracking  
✅ **Mature**: Well-established with extensive documentation  
✅ **Fast execution**: Optimized for build systems  

### Make Disadvantages
❌ **Unix-focused**: Limited Windows support  
❌ **Makefile syntax**: Steep learning curve for new users  
❌ **Platform-specific**: Heavily Unix-oriented syntax  
❌ **Poor documentation**: Limited inline documentation options  

### npm Scripts Advantages
✅ **Integrated**: Part of existing npm workflow  
✅ **JavaScript-focused**: Natural for Node.js projects  
✅ **Simple**: Easy to get started with  

### npm Scripts Disadvantages
❌ **JavaScript-only**: Limited to Node.js ecosystem  
❌ **Package.json clutter**: Can make package.json unwieldy  
❌ **Limited variable management**: Only environment variables  

## Choosing the Right Tool

### Choose Migraine if:

- ✅ You want modern, readable workflow definitions
- ✅ You need sophisticated variable management (vault system)
- ✅ You're working with multi-language projects
- ✅ You want automatic workflow discovery
- ✅ You need cross-platform compatibility
- ✅ You want rich documentation capabilities
- ✅ You want scope-aware variables (global/project/workflow)
- ✅ You want pre-checks and actions separation

### Choose Make if:

- ✅ You're building C/C++ or similar compiled projects
- ✅ You need sophisticated file dependency tracking
- ✅ You're working in Unix-only environments
- ✅ You want minimal dependencies
- ✅ You're comfortable with Make syntax

### Choose npm Scripts if:

- ✅ You're working in a JavaScript/Node.js-only project
- ✅ You want integration with npm/package management
- ✅ You prefer keeping everything in package.json
- ✅ You have simple task requirements

### Choose Taskfile if:

- ✅ You want a modern, cross-platform alternative to Make
- ✅ You prefer YAML configuration (like Migraine)
- ✅ You want simpler dependency management
- ✅ You don't need the advanced variable system of Migraine

## Migration Guide

### From Make to Migraine

1. Convert Makefile targets to YAML workflows
2. Move variables to vault or environment files
3. Add pre-checks for current Makefile prerequisites
4. Separate actions that can run independently
5. Replace `$(VAR)` syntax with `{{var}}`

### From npm Scripts to Migraine

1. Extract scripts from package.json to YAML workflows
2. Add descriptions and documentation
3. Move environment variable definitions to vault
4. Add pre-checks for validation
5. Separate actions from main steps

### From Other Tools to Migraine

1. Identify your current task patterns
2. Group related tasks into workflows
3. Define variables using vault system
4. Add validation as pre-checks
5. Configure automatic discovery in `./workflows/`