# Migraine - Documentation

Welcome to the complete documentation for Migraine, a command-line tool for organizing and automating complex workflows with templated commands. This documentation will guide you through installation, usage, and best practices.

## Table of Contents

- [Installation](#installation)
- [Getting Started](getting-started/)
- [Workflows](workflows/)
- [Vault](vault/)
- [CLI Reference](cli_reference/)
- [Alternatives Comparison](vs_alternatives/)

## Installation

### Prerequisites
- Operating System: macOS (Apple Silicon/Intel) or Linux
- Required tools: Either curl or wget for shell script installation

### Installation Methods

#### 1. Homebrew (Recommended for macOS/Linux)

```bash
# Add the Homebrew tap
brew tap tesh254/migraine https://github.com/tesh254/homebrew-migraine

# Install migraine
brew install migraine
```

#### 2. Shell Script
Using curl:

```bash
curl -sSL https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash
```

Using wget:

```bash
wget -qO- https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash
```

### Verify Installation

```bash
# Check if migraine is installed correctly
migraine --version

mgr --version
```

### Getting Updates
To update the CLI you will need to run the command based on what you used initially to install it

```bash
# Homebrew update
brew update && brew upgrade migraine

# Wget & Curl
wget -qO- https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash

wget -qO- https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash
```