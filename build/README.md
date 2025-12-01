# Build System

This directory contains the build system and tools for the Canvus Server DB Solver project.

## Files

- `Makefile` - Main build automation with cross-platform support
- `build-windows.sh` - Bash script for Windows builds
- `sync-gitlab.sh` - Script to sync with GitLab repository
- `README.md` - This documentation

## Quick Start

### Using Make (Recommended)

```bash
# Build for current platform
make build

# Build for Windows
make build-windows

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

### Using Scripts

```bash
# Windows build script
./build-windows.sh
```

## Build Targets

| Target | Description |
|--------|-------------|
| `build` | Build for current platform |
| `build-windows` | Build Windows executable (.exe) |
| `build-linux` | Build Linux executable |
| `build-mac` | Build macOS executable |
| `build-all` | Build for all platforms |
| `test` | Run all tests |
| `test-coverage` | Run tests with coverage report |
| `lint` | Run code formatting and linting |
| `clean` | Remove all build artifacts |
| `deps` | Update Go dependencies |
| `help` | Show available targets |

## Output

All built executables are placed in the `../bin/` directory:

- `canvus-server-db-solver` - Linux/macOS executable
- `canvus-server-db-solver.exe` - Windows executable
- `canvus-server-db-solver-linux` - Linux executable (cross-compiled)
- `canvus-server-db-solver-mac` - macOS executable (cross-compiled)

## Requirements

- Go 1.21 or later
- Make (for Makefile targets)
- Bash (for build scripts)

## Cross-Compilation

The build system supports cross-compilation for all major platforms:

- **Windows**: `GOOS=windows GOARCH=amd64`
- **Linux**: `GOOS=linux GOARCH=amd64`
- **macOS**: `GOOS=darwin GOARCH=amd64`

All builds use `CGO_ENABLED=0` for static linking and better compatibility.
