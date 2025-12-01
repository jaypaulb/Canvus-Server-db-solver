# Canvus Server DB Solver

A Go-based tool for identifying missing Canvus assets and locating them in backup folders.

## Overview

Canvus Server DB Solver addresses the critical issue of missing asset files in Canvus Server deployments. When asset files are missing from the filesystem but still referenced in the database, widgets fail to load. This tool identifies missing assets and reports their backup locations.

## Features

- **Automated Asset Discovery**: Queries Canvus API to identify all referenced assets
- **Missing Asset Detection**: Compares API data with filesystem to find missing files
- **Hash Lookup for Assets Without Hash**: Queries PostgreSQL database to find hash values for assets missing hash information
- **Backup Search & Location Reporting**: Searches multiple backup locations and reports file locations
- **Comprehensive Reporting**: Generates detailed reports and CSV exports with backup information
- **Parallel Processing**: Efficient handling of thousands of assets and canvases
- **Cross-Platform**: Standalone executables for Windows, Linux, and macOS

## Quick Start

### Prerequisites

- Windows 11/Server, Linux, or macOS
- Read access to Canvus Server directories
- Access to Canvus Server 3.3.0+
- Backup locations with asset files

### Installation

1. Download the latest release for your platform
2. Extract to desired location
3. Run the executable

### Usage

```bash
# Run the complete workflow
canvus-server-db-solver run

# Discover missing assets
canvus-server-db-solver discover

# Lookup hash values for assets without hash
canvus-server-db-solver lookup-hash --dry-run

# Lookup hash values and restore from backup
canvus-server-db-solver lookup-hash

# With custom INI path
canvus-server-db-solver lookup-hash --ini-path "/path/to/mt-canvus-server.ini"
```

The tool uses interactive prompts for configuration. Key settings:
- **Canvus Server URL**: Full URL to your Canvus Server API
- **Username and Password**: Canvus Server credentials
- **Assets Folder**: Path to the active Canvus assets directory
- **Backup Root Folder**: Root directory containing backup folders

## Configuration

The tool uses interactive prompts for configuration. Key settings:

- **Canvus Server URL**: Full URL to your Canvus Server API
- **Assets Folder**: Path to the active Canvus assets directory
  - Windows: `C:\ProgramData\MultiTaction\Canvus\assets`
  - Linux: `/var/lib/mt-canvus-server/assets`
- **Backup Root Folder**: Root directory containing backup folders
  - Windows: `C:\ProgramData\MultiTaction\Canvus\backups`
  - Linux: `/var/lib/mt-canvus-server/backups`
- **Database Configuration**: Automatically reads from `mt-canvus-server.ini`
- **Verbose Logging**: Optional detailed logging for troubleshooting

## Output

The tool generates:

1. **Detailed Report**: Missing assets grouped by canvas with widget information and backup locations
2. **CSV Export**: Comprehensive list of missing assets with backup status and file locations
3. **Hash Lookup Report**: For assets without hash values, shows database lookup results, assets folder search results, and backup locations
4. **Backup Location Report**: All backup file locations for assets that can be restored

## Architecture

- **Go 1.21+**: High-performance, concurrent processing
- **Canvus SDK**: Proven API integration with Canvus Server 3.3.0+
- **PostgreSQL Integration**: Direct database access for hash lookup (read-only)
- **Parallel Processing**: Simultaneous API calls and filesystem operations
- **Modular Design**: Clean separation of concerns for maintainability

## Hash Lookup Feature

The `lookup-hash` command processes assets that don't have hash values by:

1. **Database Query**: Connects to PostgreSQL database (from `mt-canvus-server.ini`) and queries the `asset_files` table by `original_filename` to retrieve public and private hash values
2. **Assets Folder Search**: Searches for the private hash in the assets folder using the first 2 characters as a subfolder (e.g., `assets/ab/abcdef123456.jpg`)
3. **Backup Search**: If not found in assets folder, searches backup locations for the hash
4. **Restoration**: Offers to restore files from backup if found

### Hash Lookup Usage

```bash
# Dry-run mode (recommended first) - shows what would be done without restoring
canvus-server-db-solver lookup-hash --dry-run

# Live mode - will offer to restore files from backup
canvus-server-db-solver lookup-hash

# Custom INI file path
canvus-server-db-solver lookup-hash --ini-path "/path/to/mt-canvus-server.ini"

# Low memory mode for systems with limited RAM
canvus-server-db-solver lookup-hash --low-memory
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

This project is open source software for Canvus Server asset management.
