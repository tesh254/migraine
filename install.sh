#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored message
print_message() {
    local color=$1
    local message=$2
    printf "${color}${message}${NC}\n"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect OS and architecture
detect_platform() {
    local platform=""
    local architecture=""

    # Detect OS
    case "$(uname -s)" in
        Darwin*)
            platform="darwin"
            ;;
        Linux*)
            platform="linux"
            ;;
        *)
            print_message "$RED" "Unsupported operating system"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            architecture="amd64"
            ;;
        arm64|aarch64)
            architecture="arm64"
            ;;
        *)
            print_message "$RED" "Unsupported architecture"
            exit 1
            ;;
    esac

    echo "${platform}-${architecture}"
}

# Get the latest release version from GitHub
get_latest_version() {
    if command_exists curl; then
        curl -s https://api.github.com/repos/tesh254/migraine/releases/latest | grep '"tag_name":' | cut -d'"' -f4
    elif command_exists wget; then
        wget -qO- https://api.github.com/repos/tesh254/migraine/releases/latest | grep '"tag_name":' | cut -d'"' -f4
    else
        print_message "$RED" "Either curl or wget is required"
        exit 1
    fi
}

# Main installation function
main() {
    print_message "$BLUE" "Starting migraine installation..."

    # Ensure either curl or wget is available
    if ! command_exists curl && ! command_exists wget; then
        print_message "$RED" "Either curl or wget is required for installation"
        exit 1
    fi

    # Detect platform
    local platform
    platform=$(detect_platform)
    print_message "$BLUE" "Detected platform: $platform"

    # Get the latest version
    local version
    version=$(get_latest_version)
    print_message "$BLUE" "Latest version: $version"

    # Create temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "${tmp_dir}"' EXIT

    # Download URL
    local download_url="https://github.com/tesh254/migraine/releases/download/${version}/migraine-${platform}"
    local binary_path="${tmp_dir}/migraine"

    print_message "$BLUE" "Downloading migraine..."

    # Download the binary
    if command_exists curl; then
        curl -L "$download_url" -o "$binary_path"
    else
        wget -O "$binary_path" "$download_url"
    fi

    # Make binary executable
    chmod +x "$binary_path"

    # Determine install location
    local install_dir="/usr/local/bin"
    if [ ! -w "$install_dir" ]; then
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
    fi

    # Move binary to installation directory
    mv "$binary_path" "${install_dir}/migraine"

    # Create symbolic link
    ln -sf "${install_dir}/migraine" "${install_dir}/mig"

    # Add to PATH if needed
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        echo "export PATH=\$PATH:$install_dir" >> "$HOME/.bashrc"
        echo "export PATH=\$PATH:$install_dir" >> "$HOME/.zshrc" 2>/dev/null || true
        print_message "$YELLOW" "Please restart your shell or run:"
        print_message "$YELLOW" "    export PATH=\$PATH:$install_dir"
    fi

    print_message "$GREEN" "✓ migraine has been installed successfully!"
    print_message "$GREEN" "✓ You can now use 'migraine' or 'mig' commands"
    print_message "$BLUE" "Run 'migraine --help' to get started"
}

# Run main function
main
