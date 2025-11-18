#!/bin/bash
# Bash script to build KPMG DB Solver for Windows
# Usage: ./scripts/build-windows.sh

echo "🔨 Building KPMG DB Solver for Windows"
echo "======================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    echo "💡 Please install Go from: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go version: $(go version)"

# Set build parameters
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
VERSION="1.1.0"
OUTPUT_DIR="bin"
OUTPUT_FILE="kpmg-db-solver.exe"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo ""
echo "🔧 Build Configuration:"
echo "   Target OS: Windows (amd64)"
echo "   Output: $OUTPUT_DIR/$OUTPUT_FILE"
echo "   Version: $VERSION"
echo "   Build Time: $BUILD_TIME"

echo ""
echo "🚀 Building executable..."

# Build the executable
cd src
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags "-X main.version=$VERSION -X main.buildTime=$BUILD_TIME" \
    -o "../$OUTPUT_DIR/$OUTPUT_FILE" \
    ./cmd/kpmg-db-solver

if [ $? -eq 0 ]; then
    echo "✅ Build completed successfully!"

    # Get file info
    FILE_SIZE=$(du -h "../$OUTPUT_DIR/$OUTPUT_FILE" | cut -f1)

    echo ""
    echo "📊 Build Results:"
    echo "   File: $(realpath "../$OUTPUT_DIR/$OUTPUT_FILE")"
    echo "   Size: $FILE_SIZE"
    echo "   Created: $(date '+%Y-%m-%d %H:%M:%S')"

    echo ""
    echo "🎉 Windows executable ready!"
    echo "💡 You can now copy $OUTPUT_DIR/$OUTPUT_FILE to your Windows machine"

else
    echo "❌ Build failed!"
    exit 1
fi

cd ..

echo ""
echo "🔍 Next Steps:"
echo "   1. Copy the executable to your Windows machine"
echo "   2. Run: .\kpmg-db-solver.exe discover"
echo "   3. Or run: .\kpmg-db-solver.exe run"
