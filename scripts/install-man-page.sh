#!/bin/bash

# Script to install the man page for migraine
# This allows users to run: man migraine or man mgr

set -e

# Check if man command exists
if ! command -v man &> /dev/null; then
    echo "Error: man command not found. Please install man pages system."
    exit 1
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Generate the man page
echo "Generating man page..."
go run "$PROJECT_ROOT/migraine.go" man generate --output "$PROJECT_ROOT"

if [ ! -f "$PROJECT_ROOT/migraine.1" ]; then
    echo "Error: Failed to generate man page"
    exit 1
fi

# Find the system man directory
MAN_DIR=""
for dir in "/usr/local/share/man" "/usr/share/man" "/opt/homebrew/share/man"; do
    if [ -d "$dir" ] && [ -w "$dir" ]; then
        MAN_DIR="$dir"
        break
    fi
done

if [ -z "$MAN_DIR" ]; then
    echo "Error: Could not find writable man directory. Try running with sudo or installing to user directory."
    exit 1
fi

# Create man1 directory if it doesn't exist
MAN1_DIR="$MAN_DIR/man1"
sudo mkdir -p "$MAN1_DIR"

# Copy the man page to the man1 directory
sudo cp "$PROJECT_ROOT/migraine.1" "$MAN1_DIR/"
sudo gzip "$MAN1_DIR/migraine.1"

# Create symlink for mgr alias
sudo ln -sf "$MAN1_DIR/migraine.1.gz" "$MAN1_DIR/mgr.1.gz"

echo "Man page installed successfully!"
echo "You can now run: man migraine or man mgr"