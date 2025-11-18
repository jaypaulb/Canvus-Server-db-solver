package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jaypaulb/kpmg-db-solver/internal/backup"
	"github.com/jaypaulb/kpmg-db-solver/internal/canvus"
	"github.com/jaypaulb/kpmg-db-solver/internal/config"
	"github.com/jaypaulb/kpmg-db-solver/internal/database"
	"github.com/jaypaulb/kpmg-db-solver/internal/filesystem"
	"github.com/jaypaulb/kpmg-db-solver/internal/logging"
	canvussdk "canvus-go-api/canvus"
)

// LookupHashCommand handles the lookup-hash command for processing items without hash
type LookupHashCommand struct {
	config  *config.Config
	dryRun  bool
	iniPath string
}

// NewLookupHashCommand creates a new lookup-hash command
func NewLookupHashCommand(cfg *config.Config) *LookupHashCommand {
	return &LookupHashCommand{
		config: cfg,
	}
}

// Execute runs the lookup-hash command
func (cmd *LookupHashCommand) Execute(cobraCmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()

	logger.Info("🔍 Starting Hash Lookup for Assets Without Hash")
	logger.Info("============================================================")

	if cmd.dryRun {
		logger.Info("🔍 DRY-RUN MODE: No files will be restored")
	}

	// Step 1: Discover assets (including those without hash)
	logger.Info("")
	logger.Info("📡 Step 1: Connecting to Canvus Server and discovering assets...")

	ctx := context.Background()
	var session *canvussdk.Session
	if cmd.config.CanvusServer.InsecureTLS {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL(), canvussdk.WithInsecureTLS())
	} else {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL())
	}

	logger.Info("🔐 Authenticating with Canvus Server...")
	err := session.Login(ctx, cmd.config.CanvusServer.Username, cmd.config.CanvusServer.Password)
	if err != nil {
		logger.Error("Authentication failed: %v", err)
		return fmt.Errorf("authentication failed: %w", err)
	}
	defer session.Logout(ctx)

	// Discover assets
	discoveryResult, err := canvus.DiscoverAllAssets(session, cmd.config.Performance.MaxConcurrentAPI)
	if err != nil {
		logger.Error("Asset discovery failed: %v", err)
		return fmt.Errorf("asset discovery failed: %w", err)
	}

	logger.Info("📈 Found %d assets without hash values", len(discoveryResult.AssetsWithoutHash))

	if len(discoveryResult.AssetsWithoutHash) == 0 {
		logger.Info("✅ No assets without hash found. Nothing to process.")
		return nil
	}

	// Step 2: Load database configuration
	logger.Info("")
	logger.Info("💾 Step 2: Loading database configuration...")

	iniPath := cmd.iniPath
	if iniPath == "" {
		var err error
		iniPath, err = database.FindINIFile()
		if err != nil {
			logger.Warn("Could not find INI file automatically: %v", err)
			iniPath = `C:\ProgramData\MultiTaction\canvus\mt-canvus-server.ini`
			logger.Info("Using default path: %s", iniPath)
		}
	}

	dbConfig, err := database.LoadDBConfigFromINI(iniPath)
	if err != nil {
		logger.Error("Failed to load database configuration: %v", err)
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	// Step 3: Connect to database
	logger.Info("")
	logger.Info("🔌 Step 3: Connecting to PostgreSQL database...")

	dbClient, err := database.NewClient(dbConfig)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbClient.Close()

	// Step 4: Process each asset without hash
	logger.Info("")
	logger.Info("🔍 Step 4: Processing assets without hash...")

	results := make([]HashLookupResult, 0)
	assetsFolder := cmd.config.Paths.AssetsFolder
	backupRootFolder := cmd.config.Paths.BackupRootFolder

	for i, asset := range discoveryResult.AssetsWithoutHash {
		if asset.OriginalFilename == "" {
			logger.Verbose("Skipping asset %d: no original filename", i+1)
			continue
		}

		logger.Info("Processing %d/%d: %s (Canvas: %s)", i+1, len(discoveryResult.AssetsWithoutHash),
			asset.OriginalFilename, asset.CanvasName)

		result := cmd.processAssetWithoutHash(dbClient, asset, assetsFolder, backupRootFolder)
		results = append(results, result)
	}

	// Step 5: Generate report
	logger.Info("")
	logger.Info("📋 Step 5: Generating report...")

	err = cmd.generateReport(results)
	if err != nil {
		logger.Error("Failed to generate report: %v", err)
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Step 6: Summary
	logger.Info("")
	logger.Info("📊 Step 6: Summary")
	logger.Info("========================================")
	logger.Info("📈 Total assets without hash: %d", len(discoveryResult.AssetsWithoutHash))
	logger.Info("🔍 Processed: %d", len(results))

	foundInDB := 0
	foundInAssets := 0
	foundInBackup := 0
	notFound := 0

	for _, result := range results {
		if result.FoundInDatabase {
			foundInDB++
		}
		if result.FoundInAssetsFolder {
			foundInAssets++
		}
		if result.FoundInBackup {
			foundInBackup++
		}
		if !result.FoundInDatabase && !result.FoundInAssetsFolder && !result.FoundInBackup {
			notFound++
		}
	}

	logger.Info("✅ Found in database: %d", foundInDB)
	logger.Info("✅ Found in assets folder: %d", foundInAssets)
	logger.Info("✅ Found in backup: %d", foundInBackup)
	logger.Info("❌ Not found: %d", notFound)

	return nil
}

// HashLookupResult represents the result of looking up a hash for an asset without hash
type HashLookupResult struct {
	Asset              canvus.AssetInfo
	FoundInDatabase    bool
	PublicHash         string
	PrivateHash        string
	FoundInAssetsFolder bool
	AssetsFolderPath   string
	FoundInBackup      bool
	BackupPath         string
	Error              string
}

// processAssetWithoutHash processes a single asset without hash
func (cmd *LookupHashCommand) processAssetWithoutHash(
	dbClient *database.Client,
	asset canvus.AssetInfo,
	assetsFolder string,
	backupRootFolder string,
) HashLookupResult {
	logger := logging.GetLogger()
	result := HashLookupResult{
		Asset: asset,
	}

	// Step 1: Query database for hash values
	logger.Verbose("Querying database for filename: %s", asset.OriginalFilename)

	dbRecords, err := dbClient.FindAssetByOriginalFilename(asset.OriginalFilename)
	if err != nil {
		// Try case-insensitive search
		dbRecords, err = dbClient.FindAssetByOriginalFilenameCaseInsensitive(asset.OriginalFilename)
		if err != nil {
			result.Error = fmt.Sprintf("Database query failed: %v", err)
			logger.Verbose("Database query failed: %v", err)
			return result
		}
	}

	if len(dbRecords) == 0 {
		result.Error = "No records found in database"
		logger.Verbose("No records found in database for: %s", asset.OriginalFilename)
		return result
	}

	// Use first record (if multiple, take the first one)
	record := dbRecords[0]
	result.FoundInDatabase = true
	result.PublicHash = record.PublicHash
	result.PrivateHash = record.PrivateHash

	logger.Verbose("Found in database: PublicHash=%s, PrivateHash=%s", record.PublicHash, record.PrivateHash)

	// Step 2: Search for private hash in assets folder
	if record.PrivateHash != "" {
		logger.Verbose("Searching for private hash in assets folder: %s", record.PrivateHash)

		hashResult, err := filesystem.SearchHashInAssetsFolder(assetsFolder, record.PrivateHash)
		if err != nil {
			logger.Verbose("Error searching assets folder: %v", err)
		} else if hashResult.Found {
			result.FoundInAssetsFolder = true
			result.AssetsFolderPath = hashResult.FilePath
			logger.Verbose("Found in assets folder: %s", hashResult.FilePath)
			return result // Found, no need to check backup
		}

		// If not found, try recursive search as fallback
		hashResult, err = filesystem.SearchHashInAssetsFolderRecursive(assetsFolder, record.PrivateHash)
		if err != nil {
			logger.Verbose("Error in recursive search: %v", err)
		} else if hashResult.Found {
			result.FoundInAssetsFolder = true
			result.AssetsFolderPath = hashResult.FilePath
			logger.Verbose("Found in assets folder (recursive): %s", hashResult.FilePath)
			return result
		}

		logger.Verbose("Not found in assets folder, searching backup...")

		// Step 3: Search in backup folder
		searcher := backup.NewSearcher(backupRootFolder)
		backupResult, err := searcher.SearchForAssets([]string{record.PrivateHash})
		if err != nil {
			logger.Verbose("Error searching backup: %v", err)
			result.Error = fmt.Sprintf("Backup search failed: %v", err)
		} else if backupFiles, found := backupResult.FoundFiles[record.PrivateHash]; found && len(backupFiles) > 0 {
			result.FoundInBackup = true
			result.BackupPath = backupFiles[0].Path // Newest file
			logger.Verbose("Found in backup: %s", backupFiles[0].Path)

			// Offer to restore if not in dry-run mode
			if !cmd.dryRun {
				logger.Info("💾 Found in backup: %s", backupFiles[0].Path)
				logger.Info("   Would restore to: %s", cmd.getTargetPath(assetsFolder, record.PrivateHash, backupFiles[0]))
				// TODO: Implement actual restore logic
			}
		} else {
			logger.Verbose("Not found in backup folder")
		}
	}

	return result
}

// getTargetPath determines the target path for restoring a file
func (cmd *LookupHashCommand) getTargetPath(assetsFolder string, hash string, backupFile backup.BackupFile) string {
	// Use first 2 chars of hash for subfolder
	if len(hash) >= 2 {
		subfolder := hash[:2]
		return filepath.Join(assetsFolder, subfolder, backupFile.Hash+backupFile.Extension)
	}
	return filepath.Join(assetsFolder, backupFile.Hash+backupFile.Extension)
}

// generateReport generates a report of the hash lookup results
func (cmd *LookupHashCommand) generateReport(results []HashLookupResult) error {
	logger := logging.GetLogger()

	reportPath := filepath.Join(cmd.config.Paths.OutputFolder, "hash_lookup_report.txt")
	content := "KPMG DB Solver - Hash Lookup Report (Assets Without Hash)\n"
	content += strings.Repeat("=", 60) + "\n"
	content += fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("Mode: %s\n", map[bool]string{true: "DRY-RUN", false: "LIVE"}[cmd.dryRun])
	content += fmt.Sprintf("Total Processed: %d\n\n", len(results))

	for i, result := range results {
		content += fmt.Sprintf("Asset %d:\n", i+1)
		content += fmt.Sprintf("  Canvas: %s (ID: %s)\n", result.Asset.CanvasName, result.Asset.CanvasID)
		content += fmt.Sprintf("  Widget: %s (ID: %s, Type: %s)\n", result.Asset.WidgetName, result.Asset.WidgetID, result.Asset.WidgetType)
		content += fmt.Sprintf("  Original Filename: %s\n", result.Asset.OriginalFilename)

		if result.FoundInDatabase {
			content += fmt.Sprintf("  ✅ Found in Database:\n")
			content += fmt.Sprintf("     Public Hash: %s\n", result.PublicHash)
			content += fmt.Sprintf("     Private Hash: %s\n", result.PrivateHash)

			if result.FoundInAssetsFolder {
				content += fmt.Sprintf("  ✅ Found in Assets Folder: %s\n", result.AssetsFolderPath)
			} else {
				content += fmt.Sprintf("  ❌ Not Found in Assets Folder\n")

				if result.FoundInBackup {
					content += fmt.Sprintf("  ✅ Found in Backup: %s\n", result.BackupPath)
					if !cmd.dryRun {
						// Get backup file info for target path
						searcher := backup.NewSearcher(cmd.config.Paths.BackupRootFolder)
						backupResult, _ := searcher.SearchForAssets([]string{result.PrivateHash})
						if backupFiles, found := backupResult.FoundFiles[result.PrivateHash]; found && len(backupFiles) > 0 {
							content += fmt.Sprintf("  💾 Would restore to: %s\n", cmd.getTargetPath(cmd.config.Paths.AssetsFolder, result.PrivateHash, backupFiles[0]))
						}
					}
				} else {
					content += fmt.Sprintf("  ❌ Not Found in Backup\n")
				}
			}
		} else {
			content += fmt.Sprintf("  ❌ Not Found in Database\n")
			if result.Error != "" {
				content += fmt.Sprintf("  Error: %s\n", result.Error)
			}
		}

		content += "\n"
	}

	err := os.WriteFile(reportPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	logger.Info("📄 Report saved to: %s", reportPath)
	return nil
}

// SetDryRun sets the dry-run flag
func (cmd *LookupHashCommand) SetDryRun(dryRun bool) {
	cmd.dryRun = dryRun
}

// SetINIPath sets the INI file path
func (cmd *LookupHashCommand) SetINIPath(iniPath string) {
	cmd.iniPath = iniPath
}

