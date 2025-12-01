# Canvus Server DB Solver - Product Requirements Document

## Problem Statement

KPMG has a Canvus Server 3.3.0 deployment with thousands of workspaces containing tens of thousands of assets. Many widgets fail to load in the Canvus Client due to missing asset files on the server's filesystem. While the database entries exist with correct hash references, the actual asset files are missing from the assets folder.

The current manual workaround of re-uploading files to make them reappear is entirely impractical given the scale (thousands of canvases, tens of thousands of assets, several terabytes of data).

## Goals & Success Metrics

### Primary Goals
- **Automated Asset Recovery**: Automatically identify and restore missing Canvus assets from backup locations
- **Comprehensive Reporting**: Generate detailed reports showing missing assets by canvas with accompanying CSV for restoration
- **Zero Database Impact**: Operate without modifying the Canvus database to avoid corruption risks

### Success Metrics
- **Asset Recovery Rate**: Successfully restore 95%+ of missing assets from available backups
- **Processing Time**: Complete asset discovery and restoration within reasonable timeframe for several terabytes of data
- **Accuracy**: 100% accuracy in identifying missing vs. present assets
- **Safety**: Zero database corruption or data loss incidents

## User Personas & Use Cases

### Primary User: KPMG System Administrator
- **Technical Level**: Intermediate to advanced
- **Environment**: Windows 11/Windows Server with admin privileges
- **Pain Points**: Manual asset recovery is impossible at scale
- **Goals**: Automated, reliable asset restoration with comprehensive reporting

### Use Cases
1. **Asset Discovery**: Identify all missing assets across all Canvus workspaces
2. **Hash Lookup**: Query PostgreSQL database to find hash values for assets without hash information
3. **Backup Search**: Locate missing assets in multiple backup generations (newest first)
4. **Asset Restoration**: Copy missing assets from backup to active assets folder
5. **Reporting**: Generate detailed reports for audit and verification purposes

## Functional Requirements

### Core Features
1. **Canvus API Integration**
   - Connect to Canvus Server 3.3.0 via username/password authentication
   - Query all canvases directly (no workspace enumeration needed)
   - Extract media asset hashes and metadata from widget JSON responses (video, image, PDF only)

2. **Database Hash Lookup** (NEW)
   - Connect to PostgreSQL database using configuration from `mt-canvus-server.ini`
   - Query `asset_files` table by `original_filename` to retrieve public and private hash values
   - Process assets that don't have hash values from API discovery
   - Support case-insensitive filename matching

3. **Filesystem Analysis**
   - Scan active assets folder for existing files
   - Search for hash values in assets folder using first 2 characters as subfolder
   - Compare API asset hashes against filesystem contents (hash-only matching)
   - Identify missing assets by hash value (ignore file extensions for matching)

4. **Backup Search & Recovery**
   - Search multiple backup folders recursively (newest to oldest)
   - Locate missing assets in backup locations
   - Copy assets from backup to active assets folder
   - Handle file permission issues (admin privileges required)

5. **Parallel Processing**
   - Simultaneous filesystem scanning and API querying for performance
   - Concurrent canvas processing with rate limiting (100/sec → 50/sec → 25/sec)
   - Concurrent backup searching across multiple backup locations

6. **Comprehensive Reporting**
   - Detailed report showing missing assets by canvas:
     ```
     Missing Files
     CanvasName - CanvasID
       WidgetName (Type: Pdf) - Hash: 347b3c308971
       WidgetName (Type: Image) - Hash: a1b2c3d4e5f6
     ```
   - CSV export with `{hash}.{ext}` filenames for restoration tracking
   - Hash lookup report showing database query results, assets folder search, and backup locations
   - Error reporting for assets found in DB but not in any backup

### User Interface
- **CLI Interface**: Command-line tool with interactive prompts
- **Configuration**: User-defined paths for assets folder and backup root folder
- **Progress Indicators**: Real-time progress reporting during long operations
- **Verbose Logging**: Optional detailed logging for audit purposes

## Non-Functional Requirements

### Performance
- **Scalability**: Handle thousands of workspaces and tens of thousands of assets
- **Efficiency**: Parallel processing for optimal performance
- **Memory Management**: Efficient handling of large datasets without memory issues

### Reliability
- **Error Handling**: Graceful handling of network issues, file system errors, and API failures
- **Data Integrity**: Verify file integrity during copy operations
- **Recovery**: Ability to resume interrupted operations

### Security
- **Authentication**: Secure credential handling for Canvus API
- **File Permissions**: Proper handling of Windows file permissions
- **Audit Trail**: Comprehensive logging for compliance requirements

### Compatibility
- **Platform**: Windows 11 and Windows Server
- **Architecture**: Standalone executable (no installer required)
- **Dependencies**: Minimal external dependencies

## Constraints & Assumptions

### Constraints
- **Database Read-Only**: Must not modify Canvus database to avoid corruption
- **API-Only Access**: Use Canvus API for data retrieval (no direct database access)
- **Windows Environment**: Target platform is Windows with admin privileges
- **Backup Structure**: Backup folders are timestamped with nightly retention (4+ generations)

### Assumptions
- **File Naming Convention**: Assets are stored as `{hash_value}.{filetype}`
- **Hash Uniqueness**: Hash collision risk is negligible for asset matching
- **Backup Integrity**: Backup files are intact and accessible
- **Network Connectivity**: Stable connection to Canvus Server during operation
- **Admin Privileges**: Tool will run with administrator privileges for file operations

## Out of Scope

### Explicitly Excluded
- **Database Modifications**: No direct database access or modifications
- **GUI Interface**: CLI-only implementation (GUI may be added later if time permits)
- **Real-time Monitoring**: No continuous monitoring or alerting capabilities
- **Asset Validation**: No validation of asset content integrity beyond file existence
- **Cross-Platform Support**: Windows-only implementation
- **Automated Scheduling**: No built-in scheduling or automation features

### Future Considerations
- **GUI Interface**: Simple graphical interface for non-technical users
- **Database Direct Access**: Read-only database access for improved performance (with safeguards)
- **Asset Content Validation**: Hash verification of restored assets
- **Automated Scheduling**: Integration with Windows Task Scheduler

## Risk Assessment

### High Risk
- **Database Corruption**: Mitigated by API-only access approach
- **File Permission Issues**: Mitigated by requiring admin privileges
- **Large Scale Performance**: Mitigated by parallel processing design

### Medium Risk
- **Network Connectivity**: API calls may fail during long operations
- **Backup Accessibility**: Backup locations may be unavailable or corrupted
- **Memory Usage**: Large datasets may cause memory issues

### Low Risk
- **File System Errors**: Standard error handling for file operations
- **User Input Validation**: Standard input validation and error messages

## Success Criteria

The project will be considered successful when:
1. **Functional**: Tool successfully identifies and restores missing assets from backups
2. **Performance**: Processes thousands of assets within reasonable timeframe
3. **Reliability**: Operates without data corruption or system instability
4. **Usability**: KPMG administrators can successfully use the tool with minimal training
5. **Reporting**: Generates comprehensive reports for audit and verification purposes

## Timeline

- **Target Completion**: September 9, 2025
- **Priority**: High - Critical business need for asset recovery
- **Approach**: Rapid development focusing on core functionality first
