# Canvus Server DB Solver - Testing Guide

## Quick Start

1. **Copy the executable** to your Windows machine
2. **Create a config file** by copying `config-sample.yaml` to `config.yaml`
3. **Update the configuration** with your actual values:
   - Canvus Server credentials
   - Assets folder path (defaults to `C:\ProgramData\MultiTaction\canvus\assets`)
   - Backup root folder path
   - TLS settings (`insecure_tls: true` for self-signed certificates)

## Testing the Discovery Command

```bash
# Test with configuration file
canvus-server-db-solver.exe discover

# The tool will:
# 1. Connect to Canvus Server at https://localhost:443
# 2. Authenticate with your credentials
# 3. List all canvases
# 4. Extract media assets (Image, PDF, Video) from each canvas
# 5. Scan the assets folder for existing files
# 6. Generate reports for missing assets
```

## Expected Output

```
🔍 Asset Discovery
==================

✅ Loaded configuration from file
🔍 Starting asset discovery...
📡 Connecting to Canvus Server: https://localhost:443
📁 Scanning assets folder: C:\ProgramData\MultiTaction\canvus\assets
🔐 Authenticating with Canvus Server...
✅ Authenticated successfully
📊 Discovering assets from Canvus API...
📈 Found X canvases with Y total media assets
🔗 Unique assets (deduplicated): Z
💾 Scanning assets folder...
📂 Found A files in assets folder (B MB total)
❌ Missing assets: C
📋 Generating reports...
📄 Detailed report saved to: missing_assets_report.txt
📊 CSV report saved to: missing_assets.csv

============================================================
📊 DISCOVERY SUMMARY
============================================================
⏱️  Discovery Duration: Xs
📈 Total Canvases: X
🎯 Total Media Assets: Y
🔗 Unique Assets: Z
💾 Files in Assets Folder: A
💽 Total Assets Size: B MB
❌ Missing Assets: C
============================================================
```

## Generated Reports

- **`missing_assets_report.txt`** - Detailed report grouped by canvas
- **`missing_assets.csv`** - CSV file for restoration tracking

## Troubleshooting

### Authentication Issues
- Verify Canvus Server is running on https://localhost:443
- Check username/password credentials
- Ensure user has access to canvases
- For self-signed certificates, set `insecure_tls: true` in config

### Path Issues
- Verify assets folder path exists (default: `C:\ProgramData\MultiTaction\canvus\assets`)
- Check backup root folder path exists (default: `C:\ProgramData\MultiTaction\canvus\backups`)
- Ensure proper Windows path format (C:\\path\\to\\folder)

### API Issues
- Check network connectivity to localhost:443
- Verify Canvus Server version 3.3.0
- Check server logs for any errors

## Current Implementation Status

✅ **Working Features:**
- Canvus Server authentication
- Canvas listing
- Media asset extraction (Image, PDF, Video)
- Hash-based asset detection
- Filesystem scanning
- Report generation (detailed + CSV)
- Parallel processing with rate limiting

🚧 **Pending Features:**
- Backup search and restoration
- Windows-specific file operations
- Error handling improvements

## Test Data Requirements

For comprehensive testing, you'll need:
- Canvus Server 3.3.0 running locally
- Multiple canvases with various media assets
- Assets folder with some missing files
- Backup folders with asset files

## Performance Notes

- Rate limiting: 100 requests/sec → 50/sec → 25/sec if needed
- Parallel processing: 10 concurrent canvas operations
- Memory efficient: Processes canvases in batches
- Progress reporting: Real-time status updates
