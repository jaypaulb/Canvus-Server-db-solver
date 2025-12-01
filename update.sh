#!/bin/bash
# Update script for canvus-server-db-solver-linux
# Downloads the latest release from GitHub and runs lookup-hash --dry-run

set -e

echo "Removing old binary..."
rm -f canvus-server-db-solver-linux

echo "Fetching latest release URL..."
DOWNLOAD_URL=$(curl -s https://api.github.com/repos/jaypaulb/canvus-server-db-solver/releases/latest | grep -o '"browser_download_url": *"[^"]*linux[^"]*"' | cut -d'"' -f4)

echo "Downloading from: $DOWNLOAD_URL"
curl -sL "$DOWNLOAD_URL" -o canvus-server-db-solver-linux

echo "Setting executable permission..."
chmod +x canvus-server-db-solver-linux

echo "Running lookup-hash --dry-run..."
./canvus-server-db-solver-linux lookup-hash --dry-run
