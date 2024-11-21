![migraine](https://github.com/user-attachments/assets/1f1f90d0-3a85-44c8-b84a-b23838bf35c2)

[![Build and Release](https://github.com/tesh254/migraine/actions/workflows/release.yml/badge.svg)](https://github.com/tesh254/migraine/actions/workflows/release.yml)

# `migraine`

This is a robust CLI tool used to organize and automate complex workflows with templated commands. Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.

## Installation

`migraine` can be installed using either Homebrew (recommended for macOS/Linux) or our shell script installer.

### System Requirements
- Operating Systems: macOS (Apple Silicon/Intel) or Linux
- Dependencies: Either `curl` or `wget` for shell script installation

### Option 1: Homebrew Installation (Recommended for macOS/Linux)

If you don't have Homebrew installed, you can install it with:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Then install migraine:

1. Add the Homebrew tap:
```bash
brew tap tesh254/migraine
```

2. Install migraine:
```bash
brew install migraine
```

To update to the latest version:
```bash
brew update && brew upgrade migraine
```

### Option 2: Shell Script Installation

You can install migraine directly using our installation script. Choose one of these commands based on your preferred download tool:

Using curl:
```bash
curl -sSL https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash
```

Using wget:
```bash
wget -qO- https://raw.githubusercontent.com/tesh254/migraine/main/install.sh | bash
```

The script will:
- Detect your operating system and architecture
- Download the appropriate binary
- Install it to `/usr/local/bin` (or `~/.local/bin` if you don't have sudo privileges)
- Create both `migraine` and `mig` commands
- Add the installation directory to your PATH if needed

### Verifying the Installation

After installation, verify that migraine is installed correctly:

```bash
migraine --version
```

Or using the short alias:
```bash
mig --version
```

### Note for Enterprise Users

If you're installing migraine in an enterprise environment where access to GitHub might be restricted, you can:

1. Download the binary directly from our [releases page](https://github.com/tesh254/migraine/releases)
2. Move it to `/usr/local/bin` (or your preferred binary location)
3. Make it executable with `chmod +x /usr/local/bin/migraine`
4. Create the alias: `ln -s /usr/local/bin/migraine /usr/local/bin/mig`

### Troubleshooting

If you encounter any issues during installation:

1. Ensure you have the required permissions:
   ```bash
   # For Homebrew installation
   sudo chown -R $(whoami) /usr/local/bin

   # For script installation
   sudo chmod +x /usr/local/bin/migraine
   ```

2. If the command isn't found after installation, try:
   ```bash
   # Add to PATH manually
   export PATH=$PATH:/usr/local/bin
   ```

3. Verify your system architecture:
   ```bash
   uname -m
   ```

For any other issues, please check our [issue tracker](https://github.com/tesh254/migraine/issues) or submit a new issue.
