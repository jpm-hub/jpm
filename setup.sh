#!/bin/sh
set -e

VERSION="1.1.5"

# JPM Unix Setup Script
echo "==============================================="
echo "      JPM Installation Script (v$VERSION)      "
echo "==============================================="
echo ""

# Detect OS and architecture
detect_os_arch() {
    # Detect OS
    if [ "$(uname -s)" = "Darwin" ]; then
        OS="darwin"
    else
        OS="linux"
    fi
    
    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)
            ARCH_TYPE="amd64"
            ;;
        arm64|aarch64)
            ARCH_TYPE="arm64"
            ;;
        *)
            ARCH_TYPE="amd64"
            echo "! Could not detect architecture, defaulting to AMD64"
            ;;
    esac
}

# Set temporary directory for download
TEMP_DIR="/tmp/jpm-install-$$"

echo "This script will install JPM $VERSION "
echo ""
echo "The following actions will be performed:"
echo ""
echo "1. Download JPM from GitHub releases"
echo "   https://github.com/jpm-hub/jpm/releases"
echo ""
echo "2. Extract and install JPM binaries to /usr/local/bin"
echo "   - /usr/local/bin/jpm"
echo "   - /usr/local/bin/jpx"
echo ""
printf "Press Enter to continue or Ctrl+C to cancel... "
read -r ok
echo ""
echo " ❗❗This script might require sudo privileges to copy jpm to /usr/local/bin"
echo ""
printf "Press Enter to continue or Ctrl+C to cancel... "
read -r ok

# Create temporary directory
echo "Creating temporary directory..."
mkdir -p "$TEMP_DIR"
echo "- Temporary directory created successfully"
echo ""

# Detect system
echo "Detecting system..."
detect_os_arch
echo "- Detected OS: $OS"
echo "- Detected architecture: $ARCH_TYPE"

# Download JPM archive file
echo "Downloading JPM from GitHub releases..."
ARCHIVE_FILE="jpm-$OS-$ARCH_TYPE.zip"
DOWNLOAD_URL="https://github.com/jpm-hub/jpm/releases/download/v$VERSION/jpm-$VERSION-$OS-$ARCH_TYPE.zip"

# Remove existing archive file if it exists
[ -f "$TEMP_DIR/$ARCHIVE_FILE" ] && rm "$TEMP_DIR/$ARCHIVE_FILE"

# Download with curl
if command -v curl >/dev/null 2>&1; then
    curl --location -o "$TEMP_DIR/$ARCHIVE_FILE" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$TEMP_DIR/$ARCHIVE_FILE" "$DOWNLOAD_URL"
else
    echo "ERROR: Neither curl nor wget is available. Please install one of them and try again."
    exit 1
fi

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to download JPM archive file"
    echo "Please check your internet connection and try again."
    exit 1
fi
echo "- Downloaded $ARCHIVE_FILE successfully"



# Extract archive file
echo "Extracting $ARCHIVE_FILE..."
cd "$TEMP_DIR"
if command -v unzip >/dev/null 2>&1; then
    unzip -q "$ARCHIVE_FILE"
else
    echo "ERROR: unzip is not available. Please install unzip and try again."
    exit 1
fi

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to extract archive file"
    exit 1
fi
rm "$ARCHIVE_FILE"
echo "- Files extracted successfully"
echo ""

# Install JPM binaries to /usr/local/bin
echo "Installing JPM binaries to /usr/local/bin... (this might require elevated privileges)"
cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
if [ -f "jpm" ]; then
    cd /usr/local/bin/
    rm -f /tmp/jpm.old >/dev/null 2>&1 || sudo rm -f /tmp/jpm.old >/dev/null 2>&1
    mv jpm /tmp/jpm.old >/dev/null 2>&1 || sudo mv jpm /tmp/jpm.old >/dev/null 2>&1 || true
    cd -
    cp jpm /usr/local/bin/jpm >/dev/null 2>&1 || sudo cp jpm /usr/local/bin/jpm 
    chmod +x /usr/local/bin/jpm >/dev/null 2>&1 || sudo chmod +x /usr/local/bin/jpm
    echo "- Installed jpm to /usr/local/bin"
else
    echo "ERROR: jpm binary not found after extraction"
    exit 1
fi
if [ -f "jpx" ]; then
    cd /usr/local/bin/
    rm -f /tmp/jpx.old >/dev/null 2>&1 || sudo rm -f /tmp/jpx.old >/dev/null 2>&1
    mv jpx /tmp/jpx.old >/dev/null 2>&1 || sudo mv jpx /tmp/jpx.old >/dev/null 2>&1 || true
    cd -
    cp jpx /usr/local/bin/jpx >/dev/null 2>&1 || sudo cp jpx /usr/local/bin/jpx
    chmod +x /usr/local/bin/jpx >/dev/null 2>&1 || sudo chmod +x /usr/local/bin/jpx
    echo "- Installed jpx to /usr/local/bin"
else
    echo "ERROR: jpx executable not found after extraction"
    exit 1
fi

# Clean up
echo "Cleaning up..."
rm -rf "$TEMP_DIR"
echo "- Temporary files removed"

echo ""
echo "==============================================="
echo "           Installation Complete!"
echo "==============================================="
echo ""
echo "JPM has been successfully installed to /usr/local/bin"
echo ""
echo "To verify installation, run:"
echo "  jpm"
echo "  jpm -h"
echo ""
