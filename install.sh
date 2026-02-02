#!/bin/bash
# Chrono CLI Installer

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

BINARY_NAME="chrono-${OS}-${ARCH}"

# Get latest version
LATEST_VERSION=$(curl -s https://api.github.com/repos/ChronoAIProject/chrono-cli/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')

if [ -z "$LATEST_VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

DOWNLOAD_URL="https://github.com/ChronoAIProject/chrono-cli/releases/download/v${LATEST_VERSION}/${BINARY_NAME}"

echo "Downloading Chrono CLI v${LATEST_VERSION} for ${OS}-${ARCH}..."
curl -sSL -o chrono "$DOWNLOAD_URL"

chmod +x chrono

# Install to /usr/local/bin if possible, otherwise ~/go/bin
if [ -w /usr/local/bin ]; then
    sudo mv chrono /usr/local/bin/
    echo "✓ Installed to /usr/local/bin/chrono"
else
    mkdir -p ~/go/bin
    mv chrono ~/go/bin/
    echo "✓ Installed to ~/go/bin/chrono"
    echo "  Add ~/go/bin to your PATH if not already there:"
    echo "  export PATH=\"\$PATH:~/go/bin\""
fi

echo "Run 'chrono --help' to get started"
