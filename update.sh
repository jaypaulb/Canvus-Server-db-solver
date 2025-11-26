#!/bin/bash
# Update script for kpmg-db-solver-linux
# Downloads the latest release from GitHub and runs lookup-hash --dry-run

rm -f kpmg-db-solver-linux
curl -sL $(curl -s https://api.github.com/repos/jaypaulb/kpmg-db-solver/releases/latest | grep "browser_download_url.*linux" | cut -d '"' -f 4) -o kpmg-db-solver-linux
chmod +x kpmg-db-solver-linux
./kpmg-db-solver-linux lookup-hash --dry-run
