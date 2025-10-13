# Migraine Documentation Index

Welcome to the Migraine documentation! This directory contains comprehensive guides to help you understand and use Migraine effectively.

## Table of Contents

### Getting Started
- [Quick Start Guide](quick-start.md) - Get up and running quickly with Migraine
- [Main Documentation](README.md) - Complete overview of Migraine features and usage

### Core Concepts
- [Workflows and Templates](workflows-and-templates.md) - Understanding workflows, the relationship between workflows and templates, and best practices
- [Configuration Files](configuration-files.md) - Using migraine.yml and other configuration files for project-level settings

### Comparisons & Context
- [Comparison with Other Tools](comparison.md) - Migraine vs Make, npm scripts, Taskfile, and Rake

### Advanced Usage
- [Best Practices and Advanced Features](best-practices.md) - Professional usage patterns, security considerations, and advanced features

## About Migraine

Migraine is a modern command-line tool for organizing and automating complex workflows with templated commands. It allows users to define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.

### Key Features
- **YAML-based workflows** with comprehensive documentation
- **Vault-backed variable management** with scope awareness
- **File-based workflow discovery** in the current working directory  
- **Template scaffolding** with commented examples
- **Migration capabilities** from legacy formats

## Quick Navigation

### Essential Commands
```bash
# Get started
migraine workflow init my-workflow -d "Description"

# List all workflows
migraine workflow list

# Run a workflow
migraine workflow run my-workflow

# Manage variables
migraine vars set my_var "value"
migraine vars list

# Get help
migraine --help
migraine workflow --help
```

### File Structure
When working with Migraine, you'll typically create:

```
my-project/
├── migraine.yml              # (Optional) Project configuration
├── workflows/                # Your workflow files
│   ├── deploy.yaml           # Deployment workflow
│   ├── build.yaml            # Build workflow  
│   └── test.yaml             # Testing workflow
├── .env                      # (Optional) Environment variables
└── src/                      # Your project files
```

## Next Steps

1. **New to Migraine?** Start with the [Quick Start Guide](quick-start.md)
2. **Want to understand workflows?** Read [Workflows and Templates](workflows-and-templates.md)  
3. **Comparing tools?** Check out [Comparison with Other Tools](comparison.md)
4. **Ready for advanced features?** See [Best Practices](best-practices.md)

## Support and Community

If you encounter issues or have questions:
- Check the documentation in this directory
- Look at the examples and best practices
- File an issue on the GitHub repository
- Check the changelog for version-specific information

Happy automating with Migraine!