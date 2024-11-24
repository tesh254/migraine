# Contributing to Migraine

Thank you for your interest in contributing to Migraine! We appreciate your help in making this CLI tool better.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Development Prerequisites](#development-prerequisites)
  - [Setting Up Your Development Environment](#setting-up-your-development-environment)
- [Making Contributions](#making-contributions)
  - [Pull Request Process](#pull-request-process)
  - [Development Guidelines](#development-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the email address specified in the Code of Conduct.

## Getting Started

### Development Prerequisites

- Go 1.23 or higher
- Git
- Basic understanding of CLI applications and workflow automation
- Familiarity with GitHub Actions (for CI/CD related contributions)

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/migraine.git
   cd migraine
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/tesh254/migraine.git
   ```
4. Create a branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Making Contributions

### Pull Request Process

1. Update the documentation when needed (including README.md and code comments)
2. Add or update tests as needed
3. Ensure all tests pass locally
4. Commit your changes using clear commit messages
5. Push to your fork and submit a pull request
6. Wait for review and address any changes requested

### Development Guidelines

1. Follow Go best practices and conventions
2. Use clear, descriptive variable and function names
3. Keep functions focused and maintainable
4. Add appropriate error handling
5. Document public functions and types
6. Follow existing code structure and patterns

#### Code Style

- Use `gofmt` to format your code
- Follow the standard Go code style
- Use meaningful variable names
- Keep functions focused and not too long
- Add comments for complex logic

#### Commit Messages

Write clear, descriptive commit messages:
- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

## Testing

- Add tests for new functionality
- Ensure existing tests pass
- Run tests locally before submitting PR:
  ```bash
  go test -v ./...
  ```
- Write both unit tests and integration tests when appropriate

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported functions and types
- Include examples in documentation when helpful
- Update CHANGELOG.md following the existing format

## Project Structure

The project follows this structure:
- `/cmd` - Command line interface code
- `/workflow` - Core workflow functionality
- `/kv` - Key-value store implementation
- `/utils` - Utility functions and helpers
- `/run` - Command execution logic

## Feature Requests and Bug Reports

- Use GitHub Issues for feature requests and bug reports
- Include as much detail as possible
- For bugs, include steps to reproduce

## Release Process

Releases are automated through GitHub Actions when a new tag is pushed:
1. Version tags should follow semantic versioning (v0.0.x)
2. The CI pipeline will:
   - Run tests
   - Build binaries for supported platforms
   - Create GitHub release
   - Update Homebrew formula

## Getting Help

If you need help with contributing:
1. Check existing issues and documentation
2. Create a new issue for questions
3. Be clear and provide context

Thank you for contributing to Migraine!
