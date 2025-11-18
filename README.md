# KPMG DB Solver (Non-Admin Version)

A Go-based tool for identifying missing Canvus assets and locating them in backup folders.

## Overview

KPMG DB Solver addresses the critical issue of missing asset files in Canvus Server deployments. When asset files are missing from the filesystem but still referenced in the database, widgets fail to load. This non-admin version identifies missing assets and reports their backup locations without requiring elevated privileges.

## Features

- **Automated Asset Discovery**: Queries Canvus API to identify all referenced assets
- **Missing Asset Detection**: Compares API data with filesystem to find missing files
- **Hash Lookup for Assets Without Hash**: Queries PostgreSQL database to find hash values for assets missing hash information
- **Backup Search & Location Reporting**: Searches multiple backup locations and reports file locations
- **Comprehensive Reporting**: Generates detailed reports and CSV exports with backup information
- **Parallel Processing**: Efficient handling of thousands of assets and canvases
- **No Admin Privileges Required**: Read-only access to system directories (restoration requires admin)
- **Windows Deployment**: Standalone executable for Windows 11/Server

## Quick Start

### Prerequisites

- Windows 11 or Windows Server
- Read access to Canvus Server directories (no admin privileges required)
- Access to Canvus Server 3.3.0
- Backup locations with asset files

### Installation

1. Download the latest release
2. Extract to desired location
3. Run normally (no administrator privileges required)

### Usage

```bash
# Run the complete workflow
kpmg-db-solver.exe run

# Discover missing assets
kpmg-db-solver.exe discover

# Lookup hash values for assets without hash (NEW)
kpmg-db-solver.exe lookup-hash --dry-run

# Lookup hash values and restore from backup
kpmg-db-solver.exe lookup-hash

# With custom INI path
kpmg-db-solver.exe lookup-hash --ini-path "C:\path\to\mt-canvus-server.ini"
```

The tool uses interactive prompts for configuration. Key settings:
- **Canvus Server URL**: Full URL to your Canvus Server API
- **Username and Password**: Canvus Server credentials
- **Assets Folder**: Path to the active Canvus assets directory
- **Backup Root Folder**: Root directory containing backup folders

## Configuration

The tool uses interactive prompts for configuration. Key settings:

- **Canvus Server URL**: Full URL to your Canvus Server API
- **Assets Folder**: Path to the active Canvus assets directory (default: `C:\ProgramData\MultiTaction\canvus\assets`)
- **Backup Root Folder**: Root directory containing backup folders (default: `C:\ProgramData\MultiTaction\canvus\backups`)
- **Database Configuration**: Automatically reads from `C:\ProgramData\MultiTaction\canvus\mt-canvus-server.ini`
- **Verbose Logging**: Optional detailed logging for troubleshooting

## Output

The tool generates:

1. **Detailed Report**: Missing assets grouped by canvas with widget information and backup locations
2. **CSV Export**: Comprehensive list of missing assets with backup status and file locations
3. **Hash Lookup Report**: For assets without hash values, shows database lookup results, assets folder search results, and backup locations
4. **Backup Location Report**: All backup file locations for assets that can be restored

## Limitations

This non-admin version provides read-only access and cannot restore assets. For asset restoration, you would need:
- Administrator privileges
- The full version of the tool
- Write access to the Canvus assets directory

## Architecture

- **Go 1.21+**: High-performance, concurrent processing
- **Canvus SDK**: Proven API integration with Canvus Server 3.3.0
- **PostgreSQL Integration**: Direct database access for hash lookup (read-only)
- **Parallel Processing**: Simultaneous API calls and filesystem operations
- **Modular Design**: Clean separation of concerns for maintainability

## New Feature: Hash Lookup for Assets Without Hash

The `lookup-hash` command processes assets that don't have hash values by:

1. **Database Query**: Connects to PostgreSQL database (from `mt-canvus-server.ini`) and queries the `asset_files` table by `original_filename` to retrieve public and private hash values
2. **Assets Folder Search**: Searches for the private hash in the assets folder using the first 2 characters as a subfolder (e.g., `assets/ab/abcdef123456.jpg`)
3. **Backup Search**: If not found in assets folder, searches backup locations for the hash
4. **Restoration**: Offers to restore files from backup if found (requires admin privileges)

### Hash Lookup Usage

```bash
# Dry-run mode (recommended first) - shows what would be done without restoring
kpmg-db-solver.exe lookup-hash --dry-run

# Live mode - will offer to restore files from backup
kpmg-db-solver.exe lookup-hash

# Custom INI file path
kpmg-db-solver.exe lookup-hash --ini-path "C:\custom\path\mt-canvus-server.ini"
```

The hash lookup report shows:
- Assets found in database with their hash values
- Whether files exist in the assets folder
- Whether files exist in backup locations
- Restoration paths for files found in backup

## Documentation

- [Product Requirements](docs/PRD.md)
- [Technical Architecture](docs/TECH_STACK.md)
- [Development Tasks](docs/TASKS.md)

## Support

For issues, questions, or contributions, please refer to the project documentation or contact the development team.

## License

This project is proprietary software developed for KPMG internal use.
