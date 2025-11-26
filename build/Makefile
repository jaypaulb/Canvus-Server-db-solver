# KPMG DB Solver Build Makefile
# Usage: make build, make build-windows, make clean, etc.

.PHONY: build build-windows build-linux build-mac clean test lint help

# Build configuration
BINARY_NAME=kpmg-db-solver
VERSION=1.3.3
BUILD_TIME=$(shell date '+%Y-%m-%d_%H:%M:%S')
BUILD_DIR=../bin
SRC_DIR=../src

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: build

# Build for current platform
build:
	@echo "🔨 Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kpmg-db-solver
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Windows
build-windows:
	@echo "🔨 Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/kpmg-db-solver
	@echo "✅ Windows build completed: $(BUILD_DIR)/$(BINARY_NAME).exe"

# Build for Linux
build-linux:
	@echo "🔨 Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/kpmg-db-solver
	@echo "✅ Linux build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# Build for macOS
build-mac:
	@echo "🔨 Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-mac ./cmd/kpmg-db-solver
	@echo "✅ macOS build completed: $(BUILD_DIR)/$(BINARY_NAME)-mac"

# Build for all platforms
build-all: build-windows build-linux build-mac
	@echo "✅ All platform builds completed"

# Run tests
test:
	@echo "🧪 Running tests..."
	cd $(SRC_DIR) && go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	cd $(SRC_DIR) && go test -cover ./...

# Run linting
lint:
	@echo "🔍 Running linters..."
	cd $(SRC_DIR) && go fmt ./...
	cd $(SRC_DIR) && go vet ./...
	@echo "✅ Linting completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/*
	@echo "✅ Clean completed"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	cd $(SRC_DIR) && go mod tidy
	@echo "✅ Dependencies updated"

# Show help
help:
	@echo "KPMG DB Solver Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build for current platform"
	@echo "  build-windows  Build for Windows (amd64)"
	@echo "  build-linux    Build for Linux (amd64)"
	@echo "  build-mac      Build for macOS (amd64)"
	@echo "  build-all      Build for all platforms"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  lint           Run linters (fmt, vet)"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Update dependencies"
	@echo "  help           Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build-windows    # Build Windows executable"
	@echo "  make test-coverage    # Run tests with coverage"
	@echo "  make clean            # Clean all build artifacts"
