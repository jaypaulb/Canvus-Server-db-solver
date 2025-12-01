# Technology Stack Decisions

## Architecture Overview

The Canvus Server DB Solver is designed as a standalone Go application that integrates with the Canvus Server API to identify and restore missing assets. The architecture follows a modular design with clear separation of concerns:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Interface │    │  Core Engine     │    │  Canvus SDK     │
│                 │    │                  │    │                 │
│ • User Input    │◄──►│ • Asset Discovery│◄──►│ • API Client    │
│ • Configuration │    │ • Backup Search  │    │ • Authentication│
│ • Progress UI   │    │ • File Operations│    │ • Data Models   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  File System    │
                       │                 │
                       │ • Asset Scanning│
                       │ • Backup Search │
                       │ • File Copying  │
                       └─────────────────┘
```

## Technology Choices

### Backend: Go 1.21+ - **Rationale**
- **Cross-platform compilation**: Single codebase compiles to Windows executable
- **Performance**: Excellent for concurrent operations and file I/O
- **Standard library**: Rich filesystem and HTTP client libraries
- **Deployment**: Single binary with no external dependencies
- **Concurrency**: Native goroutines for parallel API calls and file operations

### API Integration: Canvus Go SDK (Git Subtree) - **Rationale**
- **Proven compatibility**: Already tested with Canvus Server 3.3.0
- **Type safety**: Strongly typed request/response models
- **Authentication**: Built-in username/password and token refresh support
- **Error handling**: Centralized error handling and retry logic
- **Maintenance**: Leverages existing, maintained SDK

### File Operations: Go Standard Library - **Rationale**
- **Cross-platform**: Works consistently on Windows
- **Performance**: Efficient file copying and directory traversal
- **Concurrency**: Safe concurrent file operations with proper locking
- **Error handling**: Comprehensive error reporting for file system issues

### CLI Framework: Cobra + Viper - **Rationale**
- **Professional CLI**: Industry-standard CLI framework
- **Configuration**: Built-in support for config files and environment variables
- **Help system**: Automatic help generation and command structure
- **Validation**: Input validation and error handling

### Logging: Logrus - **Rationale**
- **Structured logging**: JSON and text output formats
- **Log levels**: Configurable verbosity (debug, info, warn, error)
- **Performance**: Efficient logging with minimal overhead
- **Windows compatibility**: Works seamlessly on Windows

### Reporting: Go Templates + CSV - **Rationale**
- **Flexibility**: Template-based reporting for customizable output
- **CSV export**: Standard format for data analysis and tracking
- **Performance**: Efficient generation of large reports
- **Compatibility**: Standard formats readable by any system

## Architecture Decision Records

### ADR-001: Go Language Selection
- **Status**: Accepted
- **Context**: Need for cross-platform Windows executable with high performance
- **Decision**: Use Go 1.21+ as the primary development language
- **Consequences**:
  - ✅ Single binary deployment
  - ✅ Excellent concurrency support
  - ✅ Rich standard library
  - ⚠️ Learning curve if team not familiar with Go

### ADR-002: Canvus SDK Integration Method
- **Status**: Accepted
- **Context**: Need to integrate with existing Canvus Go SDK
- **Decision**: Use Git subtree to embed SDK as dependency
- **Consequences**:
  - ✅ Version control of SDK version
  - ✅ No external dependency management
  - ✅ Full control over SDK modifications if needed
  - ⚠️ Manual updates when SDK changes

### ADR-003: API-Only Database Access
- **Status**: Accepted
- **Context**: Risk of database corruption with direct access
- **Decision**: Use Canvus API exclusively for data retrieval
- **Consequences**:
  - ✅ Zero risk of database corruption
  - ✅ Future-proof against database schema changes
  - ✅ Leverages existing authentication and authorization
  - ⚠️ Potentially slower than direct database access
  - ⚠️ Dependent on API availability and performance

### ADR-004: Parallel Processing Architecture
- **Status**: Accepted
- **Context**: Need to process thousands of assets efficiently
- **Decision**: Implement parallel processing for API calls and file operations
- **Consequences**:
  - ✅ Significant performance improvement
  - ✅ Better resource utilization
  - ✅ Responsive user interface during long operations
  - ⚠️ Increased complexity in error handling
  - ⚠️ Need for proper synchronization

### ADR-005: CLI-First Interface
- **Status**: Accepted
- **Context**: Speed of development more important than aesthetics
- **Decision**: Implement CLI interface first, GUI as future enhancement
- **Consequences**:
  - ✅ Faster development and deployment
  - ✅ Easier automation and scripting
  - ✅ Lower resource requirements
  - ⚠️ Less user-friendly for non-technical users
  - ⚠️ May require additional GUI development later

### ADR-006: Windows-Only Deployment
- **Status**: Accepted
- **Context**: Target environment is Windows 11/Windows Server
- **Decision**: Focus on Windows compatibility and deployment
- **Consequences**:
  - ✅ Optimized for target environment
  - ✅ Simplified testing and deployment
  - ✅ Windows-specific optimizations possible
  - ⚠️ No cross-platform support
  - ⚠️ Limited to Windows ecosystem

## Development Environment

### Required Tools
- **Go 1.21+**: Primary development language
- **Git**: Version control and subtree management
- **Windows 11/Server**: Target deployment environment
- **VS Code/GoLand**: Recommended IDE with Go support

### Build Configuration
```go
// Build tags for Windows
//go:build windows

// Cross-compilation for Windows from Linux
GOOS=windows GOARCH=amd64 go build -o canvus-server-db-solver.exe
```

### Project Structure
```
canvus-server-db-solver/
├── cmd/
│   └── canvus-server-db-solver/     # Main CLI application
├── internal/
│   ├── api/               # Canvus API integration
│   ├── filesystem/        # File operations
│   ├── backup/            # Backup search logic
│   ├── reporting/         # Report generation
│   └── config/            # Configuration management
├── pkg/
│   └── canvus/            # Canvus SDK (git subtree)
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
└── tests/                 # Test files
```

## Performance Considerations

### Concurrency Strategy
- **API Calls**: Concurrent goroutines for workspace/canvas queries
- **File Operations**: Concurrent file scanning and copying
- **Backup Search**: Parallel search across multiple backup locations
- **Rate Limiting**: Respect API rate limits to avoid server overload

### Memory Management
- **Streaming**: Process large datasets in chunks to avoid memory issues
- **Garbage Collection**: Optimize for Go's garbage collector
- **File Buffering**: Efficient file copying with appropriate buffer sizes

### Error Handling
- **Retry Logic**: Automatic retry for transient API failures
- **Graceful Degradation**: Continue operation despite individual failures
- **Comprehensive Logging**: Detailed error reporting for troubleshooting

## Security Considerations

### Authentication
- **Credential Storage**: Secure handling of Canvus credentials
- **Token Management**: Automatic token refresh and expiration handling
- **Network Security**: HTTPS-only communication with Canvus API

### File Operations
- **Permission Handling**: Proper Windows file permission management
- **Path Validation**: Prevent directory traversal attacks
- **Backup Verification**: Verify backup file integrity before restoration

### Logging
- **Sensitive Data**: Avoid logging credentials or sensitive file paths
- **Audit Trail**: Comprehensive logging for compliance requirements
- **Log Rotation**: Prevent log files from growing too large

## Deployment Strategy

### Build Process
1. **Development**: Local development with Go modules
2. **Testing**: Automated testing on Windows environment
3. **Cross-compilation**: Build Windows executable from Linux
4. **Validation**: Test executable on target Windows environment

### Distribution
- **Single Binary**: Self-contained executable with no dependencies
- **Documentation**: Comprehensive user guide and troubleshooting
- **Configuration**: Example configuration files and templates
- **Scripts**: Helper scripts for common operations

### Maintenance
- **Version Control**: Semantic versioning for releases
- **Update Mechanism**: Simple binary replacement for updates
- **Monitoring**: Log analysis for performance and error tracking
