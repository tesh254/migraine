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
    ln -sf "${install_dir}/migraine" "${install_dir}/mgr"

    # Add to PATH if needed
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        echo "export PATH=\\$PATH:$install_dir" >> "$HOME/.bashrc"
        echo "export PATH=\\$PATH:$install_dir" >> "$HOME/.zshrc" 2>/dev/null || true
        print_message "$YELLOW" "Please restart your shell or run:"
        print_message "$YELLOW" "    export PATH=\\$PATH:$install_dir"
    fi

    # Install man page if possible
    if command -v man >/dev/null 2>&1; then
        # Find the system man directory
        MAN_DIR=""
        for dir in "${install_dir}/../share/man" "/usr/local/share/man" "/usr/share/man" "/opt/homebrew/share/man"; do
            if [ -d "$dir" ] && [ -w "$dir" ]; then
                MAN_DIR="$dir"
                break
            fi
        done

        if [ -n "$MAN_DIR" ]; then
            # Create man1 directory if it doesn't exist
            MAN1_DIR="$MAN_DIR/man1"
            sudo mkdir -p "$MAN1_DIR" 2>/dev/null || mkdir -p "$HOME/.local/share/man/man1" 2>/dev/null
            
            # Try to generate man page using the installed binary
            if "${install_dir}/migraine" man generate --output "$tmp_dir" >/dev/null 2>&1; then
                MAN_FILE="$tmp_dir/migraine.1"
                if [ -f "$MAN_FILE" ]; then
                    # Try to install to system location first, fallback to user location
                    if [ -w "$MAN1_DIR" ]; then
                        sudo cp "$MAN_FILE" "$MAN1_DIR/"
                        if command -v gzip >/dev/null 2>&1; then
                            sudo gzip "$MAN1_DIR/migraine.1" 2>/dev/null || true
                        else
                            gzip "$MAN_FILE"
                        fi
                        
                        # Create symlink for mgr alias if possible
                        sudo ln -sf "$MAN1_DIR/migraine.1.gz" "$MAN1_DIR/mgr.1.gz" 2>/dev/null || true
                    else
                        # Fallback: install to user's man directory
                        USER_MAN_DIR="$HOME/.local/share/man/man1"
                        mkdir -p "$USER_MAN_DIR"
                        cp "$MAN_FILE" "$USER_MAN_DIR/migraine.1"
                        if command -v gzip >/dev/null 2>&1; then
                            gzip "$USER_MAN_DIR/migraine.1" 2>/dev/null || true
                        else
                            gzip "$MAN_FILE"
                        fi
                        
                        # Create mgr alias
                        ln -sf "$USER_MAN_DIR/migraine.1.gz" "$USER_MAN_DIR/mgr.1.gz" 2>/dev/null || true
                        
                        # Add manpath to shell config if not already there
                        if ! grep -q "MANPATH.*\\.local.*share.*man" "$HOME/.bashrc" 2>/dev/null && 
                           ! grep -q "MANPATH.*\\.local.*share.*man" "$HOME/.zshrc" 2>/dev/null; then
                            echo "export MANPATH=\\$MANPATH:\\$HOME/.local/share/man" >> "$HOME/.bashrc"
                            echo "export MANPATH=\\$MANPATH:\\$HOME/.local/share/man" >> "$HOME/.zshrc" 2>/dev/null || true
                        fi
                    fi
                    
                    print_message "$GREEN" "✓ Man page installed successfully!"
                    print_message "$BLUE" "You can now run: man migraine or man mgr"
                fi
            fi
        fi
    fi

    print_message "$GREEN" "✓ migraine has been installed successfully!"
    print_message "$GREEN" "✓ You can now use 'migraine' or 'mgr' commands"
    print_message "$BLUE" "Run 'migraine --help' to get started"
}

# Run main function
main