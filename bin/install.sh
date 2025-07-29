#!/bin/bash

# Install script for rcon-mcp-server
# Downloads and installs a specific version from GitHub releases

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if version argument is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Version number required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION="$1"
REPO="mjmorales/rcon-mcp-server"
BINARY_NAME="rcon-mcp-server"

# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Map to Go's naming convention
case "$OS" in
    Darwin) OS_NAME="Darwin" ;;
    Linux) OS_NAME="Linux" ;;
    *) 
        echo -e "${RED}Error: Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64) ARCH_NAME="x86_64" ;;
    arm64|aarch64) ARCH_NAME="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Construct download URL
FILENAME="${BINARY_NAME}_${OS_NAME}_${ARCH_NAME}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo -e "${GREEN}Installing rcon-mcp-server ${VERSION}${NC}"
echo "Platform: ${OS_NAME}_${ARCH_NAME}"
echo ""

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download the release
echo "Downloading from: $URL"
if ! curl -L -o "$TMP_DIR/$FILENAME" "$URL"; then
    echo -e "${RED}Error: Failed to download release${NC}"
    echo "Please check that version $VERSION exists"
    exit 1
fi

# Extract the binary
echo "Extracting binary..."
cd "$TMP_DIR"
if ! tar -xzf "$FILENAME"; then
    echo -e "${RED}Error: Failed to extract archive${NC}"
    exit 1
fi

# Find the binary (it should be in the root of the tar)
if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${RED}Error: Binary not found in archive${NC}"
    exit 1
fi

# Make it executable
chmod +x "$BINARY_NAME"

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Warning: Cannot write to $INSTALL_DIR${NC}"
    echo "Installing to current directory instead"
    INSTALL_DIR="."
fi

# Install the binary
echo "Installing to: $INSTALL_DIR"
if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
    # Use sudo if installing to system directory
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""

# Verify installation
if command -v "$BINARY_NAME" &> /dev/null; then
    echo "Installed version:"
    "$BINARY_NAME" --version 2>/dev/null || echo "Version information not available"
else
    echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"
    echo "You may need to add $INSTALL_DIR to your PATH"
fi