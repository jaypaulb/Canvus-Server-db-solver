#!/bin/bash
# Update script for kpmg-db-solver-linux
# Downloads the latest release from GitHub and runs lookup-hash --dry-run

set -e

echo "Removing old binary..."
rm -f kpmg-db-solver-linux

echo "Fetching latest release URL..."
DOWNLOAD_URL=$(curl -s https://api.github.com/repos/jaypaulb/kpmg-db-solver/releases/latest | grep -o '"browser_download_url": *"[^"]*linux[^"]*"' | cut -d'"' -f4)

echo "Downloading from: $DOWNLOAD_URL"
curl -sL "$DOWNLOAD_URL" -o kpmg-db-solver-linux

echo "Setting executable permission..."
chmod +x kpmg-db-solver-linux

echo "Running lookup-hash --dry-run..."
./kpmg-db-solver-linux lookup-hash --dry-run
