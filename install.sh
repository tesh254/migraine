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
        # Try to find a writable man directory (try user directory first, then system directories)
        MAN1_DIR=""
        
        # Check user directory first and create the full path if needed
        USER_MAN_DIR="$HOME/.local/share/man/man1"
        USER_MAN_PARENT_DIR="$HOME/.local/share/man"
        
        # Create the parent directories to ensure full path exists
        if mkdir -p "$USER_MAN_PARENT_DIR" 2>/dev/null && mkdir -p "$USER_MAN_DIR" 2>/dev/null; then
            MAN1_DIR="$USER_MAN_DIR"
        else
            # Try system directories, using the first one we can write to
            for dir in "/usr/local/share/man/man1" "/opt/homebrew/share/man/man1" "/usr/share/man/man1"; do
                if [ -d "$dir" ] && [ -w "$dir" ]; then
                    MAN1_DIR="$dir"
                    break
                fi
            done
        fi
        
        # If no writable directory found, try to use sudo for system directory or fallback to user
        if [ -z "$MAN1_DIR" ]; then
            # Check if we can create user directory path with mkdir if needed
            if mkdir -p "$USER_MAN_PARENT_DIR" 2>/dev/null && mkdir -p "$USER_MAN_DIR" 2>/dev/null; then
                MAN1_DIR="$USER_MAN_DIR"
            else
                # If system directories exist but aren't writable, we'll try sudo later
                for dir in "/usr/local/share/man/man1" "/opt/homebrew/share/man/man1" "/usr/share/man/man1"; do
                    if [ -d "$dir" ]; then
                        MAN1_DIR="$dir"
                        break
                    fi
                done
            fi
        fi

        # Generate and install the man page
        if "${install_dir}/migraine" man generate --output "$tmp_dir" >/dev/null 2>&1; then
            MAN_FILE="$tmp_dir/migraine.1"
            if [ -f "$MAN_FILE" ]; then
                # Install to appropriate location
                if [ -n "$MAN1_DIR" ]; then
                    # Create the directory if it doesn't exist
                    if [ ! -d "$MAN1_DIR" ]; then
                        if [ -w "$(dirname "$MAN1_DIR")" ]; then
                            mkdir -p "$MAN1_DIR"
                        else
                            # Need sudo for system directories
                            sudo mkdir -p "$MAN1_DIR" 2>/dev/null || true
                        fi
                    fi
                    
                    # Copy the man page
                    if [ -w "$MAN1_DIR" ]; then
                        cp "$MAN_FILE" "$MAN1_DIR/migraine.1"
                    else
                        # Use sudo for system directories
                        sudo cp "$MAN_FILE" "$MAN1_DIR/migraine.1" 2>/dev/null || {
                            # If sudo fails, fall back to user directory
                            MAN1_DIR="$USER_MAN_DIR"
                            mkdir -p "$MAN1_DIR"
                            cp "$MAN_FILE" "$MAN1_DIR/migraine.1"
                        }
                    fi
                    
                    # Compress the man page if gzip is available
                    if command -v gzip >/dev/null 2>&1; then
                        if [ -w "$MAN1_DIR" ]; then
                            gzip -f "$MAN1_DIR/migraine.1" 2>/dev/null || true
                        else
                            sudo gzip -f "$MAN1_DIR/migraine.1" 2>/dev/null || {
                                # If sudo gzip fails, try with user directory
                                if [[ "$MAN1_DIR" == *"$HOME"* ]]; then
                                    gzip -f "$MAN1_DIR/migraine.1" 2>/dev/null || true
                                fi
                            }
                        fi
                    fi
                    
                    # Create symlink for mgr alias
                    MANPAGE_FILE="migraine.1"
                    if [ -f "$MAN1_DIR/migraine.1.gz" ]; then
                        MANPAGE_FILE="migraine.1.gz"
                    elif [ -f "$MAN1_DIR/migraine.1" ]; then
                        MANPAGE_FILE="migraine.1"
                    fi
                    
                    if [ -w "$MAN1_DIR" ]; then
                        ln -sf "$MANPAGE_FILE" "$MAN1_DIR/mgr.1" 2>/dev/null || true
                    else
                        sudo ln -sf "$MANPAGE_FILE" "$MAN1_DIR/mgr.1" 2>/dev/null || {
                            # If sudo fails for link, try user directory
                            if [[ "$MAN1_DIR" == *"$HOME"* ]]; then
                                ln -sf "$MANPAGE_FILE" "$MAN1_DIR/mgr.1" 2>/dev/null || true
                            fi
                        }
                    fi
                    
                    print_message "$GREEN" "✓ Man page installed successfully!"
                    print_message "$BLUE" "You can now run: man migraine or man mgr"
                    
                    # Add MANPATH to shell config if installed to user directory
                    if [[ "$MAN1_DIR" == *"$HOME"* ]]; then
                        if ! grep -q "MANPATH.*\\.local.*share.*man" "$HOME/.bashrc" 2>/dev/null && 
                           ! grep -q "MANPATH.*\\.local.*share.*man" "$HOME/.zshrc" 2>/dev/null; then
                            echo "export MANPATH=\\$MANPATH:\\$HOME/.local/share/man" >> "$HOME/.bashrc"
                            echo "export MANPATH=\\$MANPATH:\\$HOME/.local/share/man" >> "$HOME/.zshrc" 2>/dev/null || true
                        fi
                    fi
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