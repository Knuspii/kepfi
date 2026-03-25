#!/usr/bin/env bash

# kepfi - Zero-dependency Install Script
# Usage: curl -sSL https://raw.githubusercontent.com/Knuspii/kepfi/main/install.sh | sudo bash

set -e

REPO="Knuspii/kepfi"
BINARY_NAME="kepfi"
INSTALL_PATH="/usr/local/bin/$BINARY_NAME"

echo "Installing kepfi..."

# Detect OS and Architecture
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
    ARCH=""
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
    ARCH="-arm64"
fi

URL="https://github.com/$REPO/releases/latest/download/${BINARY_NAME}${ARCH}"

echo "Downloading $BINARY_NAME for ${ARCH}..."
curl -L "$URL" -o "$BINARY_NAME"

echo "Setting permissions..."
chmod +x "$BINARY_NAME"

echo "Moving to $INSTALL_PATH..."
mv "$BINARY_NAME" "$INSTALL_PATH"

echo "Done! You can now use 'kepfi'."