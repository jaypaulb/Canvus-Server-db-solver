# Canvus Server DB Solver - Project Tasks

## Phase 1: Foundation & Setup

### 1.1 Project Initialization
- [✅] 1.1.1: Create project documentation (PRD, TECH_STACK, TASKS)
- [✅] 1.1.2: Initialize Go module and project structure
- [✅] 1.1.3: Set up Git repository with proper .gitignore
- [✅] 1.1.4: Add Canvus Go SDK as git subtree
- [✅] 1.1.5: Configure build scripts for Windows cross-compilation

### 1.2 Development Environment
- [✅] 1.2.1: Set up Go development environment (1.21+)
- [✅] 1.2.2: Configure IDE with Go support and debugging
- [ ] 1.2.3: Set up testing framework and test data
- [✅] 1.2.4: Create development configuration templates

## Phase 2: Core Infrastructure

### 2.1 CLI Framework
- [✅] 2.1.1: Implement Cobra CLI structure with commands
- [✅] 2.1.2: Add configuration management with Viper
- [✅] 2.1.3: Implement user input prompts for paths and credentials
- [ ] 2.1.4: Add progress indicators and verbose logging options
- [✅] 2.1.5: Create help system and command documentation

### 2.2 Configuration Management
- [✅] 2.2.1: Design configuration structure for user settings
- [✅] 2.2.2: Implement configuration file loading and validation
- [✅] 2.2.3: Add environment variable support for sensitive data
- [✅] 2.2.4: Create configuration templates and examples
- [✅] 2.2.5: Simplify configuration for localhost-only usage (HTTPS port 443)

### 2.3 Logging System
- [ ] 2.3.1: Integrate Logrus for structured logging
- [ ] 2.3.2: Implement log levels (debug, info, warn, error)
- [ ] 2.3.3: Add file and console logging options
- [ ] 2.3.4: Create log rotation and cleanup mechanisms

## Phase 3: Canvus API Integration

### 3.1 SDK Integration
- [ ] 3.1.1: Test Canvus SDK connectivity with test server
- [ ] 3.1.2: Implement authentication (username/password)
- [ ] 3.1.3: Add session management and token refresh
- [ ] 3.1.4: Create API client wrapper with error handling

### 3.2 Data Retrieval
- [ ] 3.2.1: Implement direct canvas listing functionality (no workspaces)
- [ ] 3.2.2: Add widget enumeration for each canvas
- [ ] 3.2.3: Create media asset extraction from widget data (video, image, PDF only)
- [ ] 3.2.4: Implement parallel canvas API calls for performance
- [ ] 3.2.5: Add retry logic and rate limiting (100/sec → 50/sec → 25/sec)

### 3.3 Asset Discovery
- [✅] 3.3.1: Parse widget JSON to extract asset hashes from "hash" field
- [✅] 3.3.2: Create asset metadata structure (hash, widget_type, original_filename, canvas info)
- [✅] 3.3.3: Implement asset deduplication across canvases
- [✅] 3.3.4: Add media asset type detection (video, image, PDF only)
- [✅] 3.3.5: Collect assets without hash values for database lookup

## Phase 4: Filesystem Operations

### 4.1 Asset Folder Scanning
- [ ] 4.1.1: Implement recursive directory scanning
- [ ] 4.1.2: Create file hash extraction from filenames (hash-only matching)
- [ ] 4.1.3: Add file existence checking and validation (ignore file extensions)
- [ ] 4.1.4: Implement parallel filesystem scanning

### 4.4 Dual Asset Storage Investigation
- [ ] 4.4.1: Investigate server-side asset storage: `C:\ProgramData\MultiTaction\canvus\assets`
- [ ] 4.4.2: Investigate client-side asset storage: `C:\Users\multi\AppData\Roaming\MultiTaction\canvus-data\data-300\asset-storage`
- [ ] 4.4.3: Understand interaction between server and client asset storage
- [ ] 4.4.4: Determine which storage location is primary for asset restoration
- [ ] 4.4.5: Update asset discovery and restoration logic based on findings

### 4.2 Missing Asset Detection
- [✅] 4.2.1: Compare API asset hashes with filesystem contents (hash-only matching)
- [✅] 4.2.2: Create missing asset identification logic
- [✅] 4.2.3: Add asset metadata correlation (canvas, widget info, no workspace)
- [✅] 4.2.4: Implement missing asset reporting structure

### 4.5 Database Hash Lookup (NEW)
- [✅] 4.5.1: Read PostgreSQL configuration from mt-canvus-server.ini
- [✅] 4.5.2: Connect to PostgreSQL database
- [✅] 4.5.3: Query asset_files table by original_filename
- [✅] 4.5.4: Retrieve public and private hash values from database
- [✅] 4.5.5: Search for private hash in assets folder (first 2 chars subfolder)
- [✅] 4.5.6: Search for private hash in backup folder if not found
- [✅] 4.5.7: Implement dry-run mode for hash lookup
- [✅] 4.5.8: Generate hash lookup report

### 4.3 File Operations
- [ ] 4.3.1: Create file copying utilities with progress tracking
- [ ] 4.3.2: Add Windows file permission handling
- [ ] 4.3.3: Implement file integrity verification
- [ ] 4.3.4: Add error handling for file system operations

## Phase 5: Backup Search & Recovery

### 5.1 Backup Discovery
- [ ] 5.1.1: Implement backup folder enumeration (newest first)
- [ ] 5.1.2: Create recursive backup directory scanning
- [ ] 5.1.3: Add backup timestamp parsing and sorting
- [ ] 5.1.4: Implement backup integrity checking

### 5.2 Asset Search
- [ ] 5.2.1: Create parallel backup searching across multiple locations
- [ ] 5.2.2: Implement hash-based file matching in backups
- [ ] 5.2.3: Add file version selection (newest available)
- [ ] 5.2.4: Create backup search result tracking

### 5.3 Asset Restoration
- [ ] 5.3.1: Implement asset copying from backup to active folder
- [ ] 5.3.2: Add restoration progress tracking and reporting
- [ ] 5.3.3: Create restoration verification and validation
- [ ] 5.3.4: Add error handling for failed restorations

## Phase 6: Reporting & Output

### 6.1 Report Generation
- [ ] 6.1.1: Create detailed missing asset report template
- [ ] 6.1.2: Implement canvas-grouped asset reporting (no workspace grouping)
- [ ] 6.1.3: Add asset metadata inclusion (widget names, hashes, widget_type)
- [ ] 6.1.4: Create restoration summary reporting

### 6.2 CSV Export
- [ ] 6.2.1: Generate CSV file with missing asset hashes
- [ ] 6.2.2: Add restoration tracking CSV with status
- [ ] 6.2.3: Create backup search results CSV
- [ ] 6.2.4: Implement CSV validation and formatting

### 6.3 Error Reporting
- [ ] 6.3.1: Create comprehensive error logging and reporting
- [ ] 6.3.2: Add assets-not-found-in-backup reporting
- [ ] 6.3.3: Implement restoration failure tracking
- [ ] 6.3.4: Create troubleshooting guide generation

## Phase 7: Performance & Optimization

### 7.1 Parallel Processing
- [ ] 7.1.1: Optimize API call concurrency and rate limiting
- [ ] 7.1.2: Implement efficient filesystem scanning
- [ ] 7.1.3: Add concurrent backup searching
- [ ] 7.1.4: Optimize memory usage for large datasets

### 7.2 Error Handling
- [ ] 7.2.1: Implement comprehensive error recovery
- [ ] 7.2.2: Add graceful degradation for partial failures
- [ ] 7.2.3: Create retry mechanisms for transient errors
- [ ] 7.2.4: Add operation resumption capabilities

### 7.3 Performance Monitoring
- [ ] 7.3.1: Add performance metrics and timing
- [ ] 7.3.2: Implement progress reporting for long operations
- [ ] 7.3.3: Create performance optimization recommendations
- [ ] 7.3.4: Add resource usage monitoring

## Phase 8: Testing & Validation

### 8.1 Unit Testing
- [ ] 8.1.1: Create unit tests for core functionality
- [ ] 8.1.2: Add API integration tests with mock server
- [ ] 8.1.3: Implement filesystem operation tests
- [ ] 8.1.4: Create backup search and restoration tests

### 8.2 Integration Testing
- [ ] 8.2.1: Test with real Canvus Server 3.3.0
- [ ] 8.2.2: Validate with large-scale test data
- [ ] 8.2.3: Test Windows compatibility and permissions
- [ ] 8.2.4: Verify backup restoration accuracy

### 8.3 Performance Testing
- [ ] 8.3.1: Test with thousands of assets and canvases
- [ ] 8.3.2: Validate memory usage under load
- [ ] 8.3.3: Test concurrent operation limits
- [ ] 8.3.4: Verify error handling under stress

## Phase 9: Windows Deployment

### 9.1 Build Configuration
- [ ] 9.1.1: Configure Windows cross-compilation
- [ ] 9.1.2: Create Windows-specific build scripts
- [ ] 9.1.3: Add Windows executable optimization
- [ ] 9.1.4: Test Windows compatibility

### 9.2 Deployment Package
- [ ] 9.2.1: Create deployment documentation
- [ ] 9.2.2: Add configuration examples and templates
- [ ] 9.2.3: Create user guide and troubleshooting
- [ ] 9.2.4: Add installation and setup instructions

### 9.3 Validation
- [ ] 9.3.1: Test on Windows 11 and Windows Server
- [ ] 9.3.2: Validate admin privilege requirements
- [ ] 9.3.3: Test with real KPMG environment (if available)
- [ ] 9.3.4: Verify all functionality works in production

## Phase 10: Documentation & Handover

### 10.1 User Documentation
- [ ] 10.1.1: Create comprehensive user manual
- [ ] 10.1.2: Add troubleshooting guide
- [ ] 10.1.3: Create configuration examples
- [ ] 10.1.4: Add FAQ and common issues

### 10.2 Technical Documentation
- [ ] 10.2.1: Document API integration details
- [ ] 10.2.2: Add architecture and design documentation
- [ ] 10.2.3: Create maintenance and update procedures
- [ ] 10.2.4: Add performance tuning guidelines

### 10.3 Handover
- [ ] 10.3.1: Create deployment checklist
- [ ] 10.3.2: Add monitoring and maintenance procedures
- [ ] 10.3.3: Create support and escalation procedures
- [ ] 10.3.4: Final validation and sign-off

---

## Task Dependencies

### Critical Path
1. **Foundation** (1.1-1.2) → **CLI Framework** (2.1) → **API Integration** (3.1-3.2)
2. **API Integration** → **Asset Discovery** (3.3) → **Filesystem Operations** (4.1-4.2)
3. **Dual Asset Storage Investigation** (4.4) → **Filesystem Operations** (4.1-4.2) → **Backup Search** (5.1-5.2)
4. **Backup Search** → **Asset Restoration** (5.3) → **Reporting** (6.1-6.3) → **Testing** (8.1-8.3)

### Parallel Tracks
- **Configuration & Logging** (2.2-2.3) can run parallel with CLI development
- **Performance Optimization** (7.1-7.3) can be done incrementally throughout development
- **Documentation** (10.1-10.2) can be written alongside development

## Success Criteria

Each phase is considered complete when:
- [ ] All tasks in the phase are completed
- [ ] Code passes all tests and quality checks
- [ ] Documentation is updated and accurate
- [ ] Performance meets requirements
- [ ] Ready for next phase or deployment

## Timeline Estimate

- **Phase 1-2**: 2-3 days (Foundation & CLI)
- **Phase 3**: 3-4 days (API Integration)
- **Phase 4**: 2-3 days (Filesystem Operations)
- **Phase 5**: 3-4 days (Backup & Recovery)
- **Phase 6**: 2-3 days (Reporting)
- **Phase 7**: 2-3 days (Performance)
- **Phase 8**: 3-4 days (Testing)
- **Phase 9**: 2-3 days (Windows Deployment)
- **Phase 10**: 1-2 days (Documentation)

**Total Estimated Time**: 20-30 days
**Target Completion**: September 9, 2025
