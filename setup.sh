#!/bin/sh
set -e
VERSION="1.1.5"

# Default install directory
INSTALL_DIR="/usr/local/bin"

# Parse optional argument for install directory
# ...existing code...
# Parse optional argument for install directory
if [ $# -ge 1 ]; then
    INSTALL_DIR="$1"
    INSTALL_DIR="${INSTALL_DIR#\\}"
fi
# ...existing code...

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
echo "2. Extract and install JPM binaries to $INSTALL_DIR"
echo "   - $INSTALL_DIR/jpm"
echo "   - $INSTALL_DIR/jpx"
echo ""
if [ "$INSTALL_DIR" != "/usr/local/bin" ]; then
    echo " ⚠️  Note: You are installing to a custom directory ($INSTALL_DIR)."
    echo "     Make sure $INSTALL_DIR is in your PATH to use 'jpm' and 'jpx' from the command line."
    echo ""
    printf "Press Enter to continue or Ctrl+C to cancel... "
    read -r ok
else 
echo ""
echo " ❗❗This script might require sudo privileges to copy jpm to $INSTALL_DIR"
echo ""
echo "    -  to install in a different directory, run: sh $0 <install_directory>"
echo ""
printf "Press Enter to continue or Ctrl+C to cancel... "
read -r ok
fi

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

echo "Installing JPM binaries to $INSTALL_DIR... (this might require elevated privileges)"
cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
if [ ! -f "jpm" ] || [ ! -f "jpx" ]; then
    echo "ERROR: jpm or jpx binary not found after extraction"
    exit 1
fi
if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
    cd "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR" && cd "$INSTALL_DIR"
    rm -f /tmp/jpm.old >/dev/null 2>&1 || sudo rm -f /tmp/jpm.old >/dev/null 2>&1
    mv jpm /tmp/jpm.old >/dev/null 2>&1 || sudo mv jpm /tmp/jpm.old >/dev/null 2>&1 || true
    cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
    cp jpm "$INSTALL_DIR/jpm" >/dev/null 2>&1 || sudo cp jpm "$INSTALL_DIR/jpm" >/dev/null 2>&1
    chmod +x "$INSTALL_DIR/jpm" >/dev/null 2>&1 || sudo chmod +x "$INSTALL_DIR/jpm" >/dev/null 2>&1
    echo "- Installed jpm to $INSTALL_DIR"
    cd "$INSTALL_DIR" 2>/dev/null
    rm -f /tmp/jpx.old >/dev/null 2>&1 || sudo rm -f /tmp/jpx.old >/dev/null 2>&1
    mv jpx /tmp/jpx.old >/dev/null 2>&1 || sudo mv jpx /tmp/jpx.old >/dev/null 2>&1 || true
    cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
    cp jpx "$INSTALL_DIR/jpx" >/dev/null 2>&1 || sudo cp jpx "$INSTALL_DIR/jpx"
    chmod +x "$INSTALL_DIR/jpx" >/dev/null 2>&1 || sudo chmod +x "$INSTALL_DIR/jpx"
    echo "- Installed jpx to $INSTALL_DIR"
else
    cd "$INSTALL_DIR" 2>/dev/null || mkdir -p "$INSTALL_DIR" && cd "$INSTALL_DIR"
    rm -f jpm.old >/dev/null 2>&1
    mv jpm jpm.old >/dev/null 2>&1 || true
    cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
    cp jpm "$INSTALL_DIR/jpm" >/dev/null 2>&1
    chmod +x "$INSTALL_DIR/jpm" >/dev/null 2>&1
    echo "- Installed jpm to $INSTALL_DIR"
    cd "$INSTALL_DIR" 2>/dev/null
    rm -f jpx.old >/dev/null 2>&1
    mv jpx jpx.old >/dev/null 2>&1 || true
    cd "$TEMP_DIR/jpm-$VERSION-$OS-$ARCH_TYPE/bin/"
    cp jpx "$INSTALL_DIR/jpx" >/dev/null 2>&1
    chmod +x "$INSTALL_DIR/jpx" >/dev/null 2>&1
    echo "- Installed jpx to $INSTALL_DIR"
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
echo "JPM has been successfully installed to $INSTALL_DIR"
echo ""
echo "To verify installation, run:"
echo "  jpm"
echo "  jpm -h"
echo ""
