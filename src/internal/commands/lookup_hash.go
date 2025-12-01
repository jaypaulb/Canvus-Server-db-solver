package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/jaypaulb/canvus-server-db-solver/internal/backup"
	"github.com/jaypaulb/canvus-server-db-solver/internal/canvus"
	"github.com/jaypaulb/canvus-server-db-solver/internal/config"
	"github.com/jaypaulb/canvus-server-db-solver/internal/database"
	"github.com/jaypaulb/canvus-server-db-solver/internal/filesystem"
	"github.com/jaypaulb/canvus-server-db-solver/internal/logging"
	canvussdk "canvus-go-api/canvus"
)

// LookupHashCommand handles the lookup-hash command for processing items without hash
type LookupHashCommand struct {
	config       *config.Config
	dryRun       bool
	iniPath      string
	skipArchived bool
	lowMemory    bool // Low memory mode - skip catalog building, search on demand
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

	// Step 1: Validate database configuration and connection BEFORE making API calls
	logger.Info("")
	logger.Info("✅ Step 1: Validating database configuration and connection...")

	iniPath := cmd.iniPath
	if iniPath == "" {
		var err error
		iniPath, err = database.FindINIFile()
		if err != nil {
			logger.Warn("Could not find INI file automatically: %v", err)
			iniPath = `C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`
			logger.Info("Using default path: %s", iniPath)
		}
	}

	logger.Info("📄 Loading database configuration from: %s", iniPath)
	dbConfig, err := database.LoadDBConfigFromINI(iniPath)
	if err != nil {
		logger.Error("Failed to load database configuration: %v", err)
		logger.Error("Please ensure the INI file exists and contains the required database settings:")
		logger.Error("  - host (or hostname, server)")
		logger.Error("  - database (or dbname, name)")
		logger.Error("  - username (or user)")
		logger.Error("  - password (or pass) [optional]")
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	logger.Info("🔌 Testing database connection...")
	dbClient, err := database.NewClient(dbConfig)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		logger.Error("Database connection details:")
		logger.Error("  Host: %s", dbConfig.Host)
		logger.Error("  Port: %s", dbConfig.Port)
		logger.Error("  Database: %s", dbConfig.Database)
		logger.Error("  Username: %s", dbConfig.Username)
		logger.Error("Please verify that:")
		logger.Error("  1. The database server is running and accessible")
		logger.Error("  2. The credentials are correct")
		logger.Error("  3. The database exists")
		logger.Error("  4. Network connectivity is available")
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("✅ Database connection successful!")

	// Keep connection open for later use
	defer dbClient.Close()

	// Step 2: Validate Canvus Server configuration
	logger.Info("")
	logger.Info("✅ Step 2: Validating Canvus Server configuration...")

	if cmd.config.CanvusServer.Username == "" || cmd.config.CanvusServer.Password == "" {
		logger.Error("Canvus Server credentials not configured")
		return fmt.Errorf("canvus Server credentials not configured in config.yaml")
	}

	ctx := context.Background()
	var session *canvussdk.Session
	if cmd.config.CanvusServer.InsecureTLS {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL(), canvussdk.WithInsecureTLS())
	} else {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL())
	}

	logger.Info("🔐 Testing Canvus Server authentication...")
	err = session.Login(ctx, cmd.config.CanvusServer.Username, cmd.config.CanvusServer.Password)
	if err != nil {
		logger.Error("Authentication failed: %v", err)
		logger.Error("Please verify that:")
		logger.Error("  1. The Canvus Server URL is correct: %s", cmd.config.GetCanvusAPIURL())
		logger.Error("  2. The username and password are correct")
		logger.Error("  3. The Canvus Server is running and accessible")
		return fmt.Errorf("authentication failed: %w", err)
	}
	logger.Info("✅ Canvus Server authentication successful!")
	defer session.Logout(ctx)

	// Step 3: Validate file system paths
	logger.Info("")
	logger.Info("✅ Step 3: Validating file system paths...")

	assetsFolder := cmd.config.Paths.AssetsFolder
	backupRootFolder := cmd.config.Paths.BackupRootFolder

	if assetsFolder == "" {
		logger.Error("Assets folder not configured in config.yaml")
		return fmt.Errorf("assets folder not configured")
	}
	if _, err := os.Stat(assetsFolder); os.IsNotExist(err) {
		logger.Error("Assets folder does not exist: %s", assetsFolder)
		return fmt.Errorf("assets folder does not exist: %s", assetsFolder)
	}
	logger.Info("✅ Assets folder accessible: %s", assetsFolder)

	if backupRootFolder == "" {
		logger.Error("Backup root folder not configured in config.yaml")
		return fmt.Errorf("backup root folder not configured")
	}
	if _, err := os.Stat(backupRootFolder); os.IsNotExist(err) {
		logger.Error("Backup root folder does not exist: %s", backupRootFolder)
		return fmt.Errorf("backup root folder does not exist: %s", backupRootFolder)
	}
	logger.Info("✅ Backup root folder accessible: %s", backupRootFolder)

	// Step 4: Discover assets (including those without hash)
	logger.Info("")
	logger.Info("📡 Step 4: Discovering assets from Canvus Server...")
	logger.Info("⏳ This may take some time for large deployments...")

	// Discover assets with options
	discoveryOptions := canvus.DiscoveryOptions{
		SkipArchived: cmd.skipArchived,
	}
	discoveryResult, err := canvus.DiscoverAllAssetsWithOptions(session, cmd.config.Performance.MaxConcurrentAPI, discoveryOptions)
	if err != nil {
		logger.Error("Asset discovery failed: %v", err)
		return fmt.Errorf("asset discovery failed: %w", err)
	}

	logger.Info("📈 Found %d assets without hash values", len(discoveryResult.AssetsWithoutHash))

	if len(discoveryResult.AssetsWithoutHash) == 0 {
		logger.Info("✅ No assets without hash found. Nothing to process.")
		return nil
	}

	// Step 5: Build hash catalogs for assets and backup folders (skip in low-memory mode)
	var assetsCatalog *filesystem.HashCatalog
	var backupCatalog *filesystem.HashCatalog

	if cmd.lowMemory {
		logger.Info("")
		logger.Info("📂 Step 5: LOW MEMORY MODE - Skipping catalog pre-build")
		logger.Info("⚠️  Files will be searched on-demand (slower but uses less RAM)")
	} else {
		logger.Info("")
		logger.Info("📂 Step 5: Building hash catalogs for fast lookups...")
		logger.Info("⏳ Scanning filesystem once instead of %d times...", len(discoveryResult.AssetsWithoutHash))

		// Build assets folder catalog
		assetsCatalog = filesystem.NewHashCatalog()
		err = assetsCatalog.BuildCatalogFromFolder(assetsFolder)
		if err != nil {
			logger.Warn("Failed to build assets catalog: %v", err)
			logger.Warn("Continuing without assets catalog...")
		}

		// Build backup folder catalog
		backupCatalog = filesystem.NewHashCatalog()
		err = backupCatalog.BuildCatalogFromFolder(backupRootFolder)
		if err != nil {
			logger.Warn("Failed to build backup catalog: %v", err)
			logger.Warn("Continuing without backup catalog...")
		}

		logger.Info("✅ Hash catalogs ready:")
		logger.Info("   📁 Assets folder: %d unique hashes", assetsCatalog.Size())
		logger.Info("   💾 Backup folder: %d unique hashes", backupCatalog.Size())
	}

	// Step 6: Batch load all database records
	logger.Info("")
	logger.Info("🗄️  Step 6: Loading database records into memory...")

	dbCatalog := make(map[string]database.AssetFileRecord)
	for _, asset := range discoveryResult.AssetsWithoutHash {
		if asset.OriginalFilename == "" {
			continue
		}

		// Try to find in database if not already loaded
		if _, exists := dbCatalog[asset.OriginalFilename]; !exists {
			records, err := dbClient.FindAssetByOriginalFilename(asset.OriginalFilename)
			if err != nil || len(records) == 0 {
				// Try case-insensitive
				records, err = dbClient.FindAssetByOriginalFilenameCaseInsensitive(asset.OriginalFilename)
			}
			if err == nil && len(records) > 0 {
				dbCatalog[asset.OriginalFilename] = records[0]
			}
		}
	}

	logger.Info("✅ Database catalog ready: %d records loaded", len(dbCatalog))

	// Step 7: Deduplicate assets by Canvas+Filename combination
	logger.Info("")
	logger.Info("🔍 Step 7: Deduplicating assets by canvas and filename...")

	// Group assets by (CanvasName, OriginalFilename) tuple
	type AssetKey struct {
		CanvasName       string
		OriginalFilename string
	}

	uniqueAssets := make(map[AssetKey][]canvus.AssetInfo)
	skippedCount := 0

	for _, asset := range discoveryResult.AssetsWithoutHash {
		if asset.OriginalFilename == "" {
			skippedCount++
			continue
		}

		key := AssetKey{
			CanvasName:       asset.CanvasName,
			OriginalFilename: asset.OriginalFilename,
		}
		uniqueAssets[key] = append(uniqueAssets[key], asset)
	}

	// Create list of unique assets (one per group)
	uniqueAssetsList := make([]canvus.AssetInfo, 0, len(uniqueAssets))
	assetGroups := make(map[AssetKey][]canvus.AssetInfo)

	for key, group := range uniqueAssets {
		uniqueAssetsList = append(uniqueAssetsList, group[0]) // Use first instance as representative
		assetGroups[key] = group
	}

	logger.Info("📊 Deduplicated: %d unique files from %d total widget instances", len(uniqueAssetsList), len(discoveryResult.AssetsWithoutHash))
	logger.Info("   Skipped: %d assets without filenames", skippedCount)
	logger.Info("")
	logger.Info("⚡ Processing %d unique files using %d workers...", len(uniqueAssetsList), cmd.config.Performance.MaxConcurrentAPI)

	// Process unique assets in parallel
	uniqueResults := cmd.processAssetsParallel(uniqueAssetsList, dbCatalog, assetsCatalog, backupCatalog)

	// Expand results back to all instances
	results := make([]HashLookupResult, 0, len(discoveryResult.AssetsWithoutHash))
	for _, uniqueAsset := range uniqueAssetsList {
		key := AssetKey{
			CanvasName:       uniqueAsset.CanvasName,
			OriginalFilename: uniqueAsset.OriginalFilename,
		}

		// Find the result for this unique asset
		var uniqueResult HashLookupResult
		for _, r := range uniqueResults {
			if r.Asset.CanvasName == uniqueAsset.CanvasName && r.Asset.OriginalFilename == uniqueAsset.OriginalFilename {
				uniqueResult = r
				break
			}
		}

		// Replicate result for all instances in this group
		for _, instance := range assetGroups[key] {
			instanceResult := uniqueResult
			instanceResult.Asset = instance // Preserve original widget IDs
			results = append(results, instanceResult)
		}
	}

	// Step 8: Generate report
	logger.Info("")
	logger.Info("📋 Step 8: Generating report...")

	err = cmd.generateReport(results)
	if err != nil {
		logger.Error("Failed to generate report: %v", err)
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Step 9: Summary with unique file counts
	logger.Info("")
	logger.Info("📊 Step 9: Summary")
	logger.Info("========================================")

	// Count unique files and total instances
	uniqueFoundInDB := make(map[AssetKey]bool)
	uniqueFoundInAssets := make(map[AssetKey]bool)
	uniqueFoundInBackup := make(map[AssetKey]bool)
	uniqueNotFound := make(map[AssetKey]bool)

	foundInDBInstances := 0
	foundInAssetsInstances := 0
	foundInBackupInstances := 0
	notFoundInstances := 0

	for _, result := range results {
		key := AssetKey{
			CanvasName:       result.Asset.CanvasName,
			OriginalFilename: result.Asset.OriginalFilename,
		}

		if result.FoundInDatabase {
			uniqueFoundInDB[key] = true
			foundInDBInstances++
		}
		if result.FoundInAssetsFolder {
			uniqueFoundInAssets[key] = true
			foundInAssetsInstances++
		}
		if result.FoundInBackup {
			uniqueFoundInBackup[key] = true
			foundInBackupInstances++
		}
		if !result.FoundInDatabase && !result.FoundInAssetsFolder && !result.FoundInBackup {
			uniqueNotFound[key] = true
			notFoundInstances++
		}
	}

	logger.Info("📈 Total: %d widget instances across %d unique files", len(results), len(uniqueAssetsList))
	logger.Info("✅ Found in database: %d unique files", len(uniqueFoundInDB))
	logger.Info("✅ Found in assets folder: %d unique files (%d instances)", len(uniqueFoundInAssets), foundInAssetsInstances)
	logger.Info("✅ Found in backup: %d unique files (%d instances)", len(uniqueFoundInBackup), foundInBackupInstances)
	logger.Info("❌ Not found: %d unique files (%d instances)", len(uniqueNotFound), notFoundInstances)

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

// processAssetsParallel processes multiple assets in parallel using a worker pool
func (cmd *LookupHashCommand) processAssetsParallel(
	assets []canvus.AssetInfo,
	dbCatalog map[string]database.AssetFileRecord,
	assetsCatalog *filesystem.HashCatalog,
	backupCatalog *filesystem.HashCatalog,
) []HashLookupResult {
	logger := logging.GetLogger()
	numWorkers := cmd.config.Performance.MaxConcurrentAPI
	if numWorkers <= 0 {
		numWorkers = 4 // Default to 4 workers
	}

	results := make([]HashLookupResult, len(assets))

	// Create work channel and wait group
	workChan := make(chan int, len(assets))
	var wg sync.WaitGroup
	var progressMutex sync.Mutex
	processedCount := 0

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workChan {
				asset := assets[i]

				// Update progress counter
				progressMutex.Lock()
				processedCount++
				currentProgress := processedCount
				progressMutex.Unlock()

				// Log progress every 100 files or first file
				if currentProgress%100 == 0 || currentProgress == 1 {
					logger.Info("")
					logger.Info("═══ Progress: %d/%d unique files processed ═══", currentProgress, len(assets))
					logger.Info("")
				}

				result := cmd.processAssetWithCatalogs(asset, dbCatalog, assetsCatalog, backupCatalog)
				results[i] = result
			}
		}()
	}

	// Send work to workers
	for i := range assets {
		workChan <- i
	}
	close(workChan)

	// Wait for all workers to finish
	wg.Wait()

	logger.Info("✅ Completed processing %d assets", len(assets))
	return results
}

// processAssetWithCatalogs processes a single asset using pre-built hash catalogs (or on-demand search in low-memory mode)
// When catalogs are provided, uses O(1) hash lookups. When nil (low-memory mode), searches filesystem on-demand.
func (cmd *LookupHashCommand) processAssetWithCatalogs(
	asset canvus.AssetInfo,
	dbCatalog map[string]database.AssetFileRecord,
	assetsCatalog *filesystem.HashCatalog,
	backupCatalog *filesystem.HashCatalog,
) HashLookupResult {
	logger := logging.GetLogger()
	result := HashLookupResult{
		Asset: asset,
	}

	// Step 1: Lookup in database catalog (O(1))
	dbRecord, foundInDB := dbCatalog[asset.OriginalFilename]
	if !foundInDB {
		result.Error = "Not found in database"
		logger.Info("Database: '%s' on canvas '%s' → Not Found", asset.OriginalFilename, asset.CanvasName)
		return result
	}

	result.FoundInDatabase = true
	result.PublicHash = dbRecord.PublicHash
	result.PrivateHash = dbRecord.PrivateHash

	// Get last 4 chars of public hash for logging
	hashSuffix := "N/A"
	if len(dbRecord.PublicHash) >= 4 {
		hashSuffix = dbRecord.PublicHash[len(dbRecord.PublicHash)-4:]
	}

	logger.Info("Database: '%s' on canvas '%s' → hash=[...%s]", asset.OriginalFilename, asset.CanvasName, hashSuffix)

	// Step 2: Lookup in assets folder
	if dbRecord.PrivateHash != "" {
		var foundInAssets bool
		var assetsPath string

		if assetsCatalog != nil {
			// Fast path: use pre-built catalog (O(1))
			if entry, found := assetsCatalog.Lookup(dbRecord.PrivateHash); found {
				foundInAssets = true
				assetsPath = entry.FilePath
			}
		} else {
			// Low-memory mode: search on-demand
			hashResult, err := filesystem.SearchHashInAssetsFolder(cmd.config.Paths.AssetsFolder, dbRecord.PrivateHash)
			if err == nil && hashResult.Found {
				foundInAssets = true
				assetsPath = hashResult.FilePath
			}
		}

		if foundInAssets {
			result.FoundInAssetsFolder = true
			result.AssetsFolderPath = assetsPath
			logger.Info("Assets folder: '%s' hash=[...%s] in '%s' → Found",
				asset.OriginalFilename, hashSuffix, cmd.config.Paths.AssetsFolder)
			return result // Found, no need to check backup
		}

		logger.Info("Assets folder: '%s' hash=[...%s] in '%s' → Not Found",
			asset.OriginalFilename, hashSuffix, cmd.config.Paths.AssetsFolder)

		// Step 3: Lookup in backup folder
		var foundInBackup bool
		var backupPath string
		var backupExt string

		if backupCatalog != nil {
			// Fast path: use pre-built catalog (O(1))
			if entry, found := backupCatalog.Lookup(dbRecord.PrivateHash); found {
				foundInBackup = true
				backupPath = entry.FilePath
				backupExt = entry.Extension
			}
		} else {
			// Low-memory mode: search on-demand using backup searcher
			searcher := backup.NewSearcher(cmd.config.Paths.BackupRootFolder)
			backupResult, err := searcher.SearchForAssets([]string{dbRecord.PrivateHash})
			if err == nil {
				if files, found := backupResult.FoundFiles[dbRecord.PrivateHash]; found && len(files) > 0 {
					foundInBackup = true
					backupPath = files[0].Path
					backupExt = files[0].Extension
				}
			}
		}

		if foundInBackup {
			result.FoundInBackup = true
			result.BackupPath = backupPath

			// Calculate target path
			targetPath := cmd.getTargetPathFromHash(dbRecord.PrivateHash, backupExt)

			if cmd.dryRun {
				logger.Info("Backup folder: '%s' hash=[...%s] in '%s' → Found at '%s' | Restoring: Skipped (Dry-Run) → Target: '%s'",
					asset.OriginalFilename, hashSuffix, cmd.config.Paths.BackupRootFolder,
					backupPath, targetPath)
			} else {
				logger.Info("Backup folder: '%s' hash=[...%s] in '%s' → Found at '%s' | Restoring: Active → Target: '%s'",
					asset.OriginalFilename, hashSuffix, cmd.config.Paths.BackupRootFolder,
					backupPath, targetPath)
				// TODO: Implement actual restore logic
			}
		} else {
			logger.Info("Backup folder: '%s' hash=[...%s] in '%s' → Not Found",
				asset.OriginalFilename, hashSuffix, cmd.config.Paths.BackupRootFolder)
		}
	}

	return result
}

// getTargetPathFromHash calculates the target path for restoring a file based on its hash
func (cmd *LookupHashCommand) getTargetPathFromHash(hash string, extension string) string {
	if len(hash) >= 2 {
		subfolder := hash[:2]
		return filepath.Join(cmd.config.Paths.AssetsFolder, subfolder, hash+extension)
	}
	return filepath.Join(cmd.config.Paths.AssetsFolder, hash+extension)
}

// processAssetWithoutHash processes a single asset without hash
// DEPRECATED: Use processAssetWithCatalogs for much better performance
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

	// Save report in current working directory (where exe is being run from)
	reportPath := "hash_lookup_report.txt"
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

// SetSkipArchived sets the skip-archived flag
func (cmd *LookupHashCommand) SetSkipArchived(skipArchived bool) {
	cmd.skipArchived = skipArchived
}

// SetLowMemory sets the low-memory mode flag
func (cmd *LookupHashCommand) SetLowMemory(lowMemory bool) {
	cmd.lowMemory = lowMemory
}

