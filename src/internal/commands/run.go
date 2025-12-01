package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaypaulb/canvus-server-db-solver/internal/backup"
	"github.com/jaypaulb/canvus-server-db-solver/internal/canvus"
	"github.com/jaypaulb/canvus-server-db-solver/internal/config"
	"github.com/jaypaulb/canvus-server-db-solver/internal/filesystem"
	"github.com/jaypaulb/canvus-server-db-solver/internal/logging"
	canvussdk "canvus-go-api/canvus"
)

// RunCommand handles the run command (complete workflow)
type RunCommand struct {
	config *config.Config
}

// NewRunCommand creates a new run command
func NewRunCommand(cfg *config.Config) *RunCommand {
	return &RunCommand{
		config: cfg,
	}
}

// Execute runs the complete workflow sequentially
func (cmd *RunCommand) Execute(cobraCmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()

	logger.Info("🚀 Starting KPMG DB Solver - Complete Workflow")
	logger.Info("============================================================")

	// Step 1: Authentication and Asset Discovery
	logger.Info("📡 Step 1: Connecting to Canvus Server and discovering assets...")
	logger.Info("📡 Connecting to Canvus Server: %s", cmd.config.CanvusServer.URL)
	logger.Info("📁 Scanning assets folder: %s", cmd.config.Paths.AssetsFolder)

	// Create Canvus session
	ctx := context.Background()
	var session *canvussdk.Session
	if cmd.config.CanvusServer.InsecureTLS {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL(), canvussdk.WithInsecureTLS())
	} else {
		session = canvussdk.NewSession(cmd.config.GetCanvusAPIURL())
	}

	// Authenticate
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

	logger.Info("📈 Found %d canvases with %d total media assets",
		len(discoveryResult.Canvases), len(discoveryResult.Assets))

	// Get unique assets
	uniqueAssets := discoveryResult.GetUniqueAssets()
	logger.Info("🔗 Unique assets (deduplicated): %d", len(uniqueAssets))

	// Step 2: Filesystem Scanning
	logger.Info("")
	logger.Info("💾 Step 2: Scanning local assets folder...")

	// Extract asset hashes for filesystem comparison
	assetHashes := make([]string, len(uniqueAssets))
	for i, asset := range uniqueAssets {
		assetHashes[i] = asset.Hash
	}

	// Scan filesystem
	scanResult, err := filesystem.ScanAssetsFolder(cmd.config.Paths.AssetsFolder)
	if err != nil {
		logger.Error("Filesystem scan failed: %v", err)
		return fmt.Errorf("filesystem scan failed: %w", err)
	}

	logger.Info("📂 Found %d files in assets folder (%.2f MB total)",
		len(scanResult.Files), float64(scanResult.TotalSize)/(1024*1024))

	// Find missing assets
	missingAssets := filesystem.FindMissingAssets(assetHashes, scanResult)
	logger.Info("❌ Missing assets: %d", len(missingAssets))

	if len(missingAssets) == 0 {
		logger.Info("")
		logger.Info("✅ No missing assets found! All assets are present.")
		logger.Info("🎉 Workflow completed successfully!")
		return nil
	}

	// Step 3: Backup Search
	logger.Info("")
	logger.Info("🔍 Step 3: Searching for missing assets in backup folder...")

	searcher := backup.NewSearcher(cmd.config.Paths.BackupRootFolder)
	backupSearchResult, err := searcher.SearchForAssets(missingAssets)
	if err != nil {
		logger.Error("Backup search failed: %v", err)
		return fmt.Errorf("backup search failed: %w", err)
	}

	// Sort backup files by modification time (newest first)
	searcher.SortBackupFiles(backupSearchResult)

	// Step 4: Asset Discovery Summary (restoration disabled in non-admin version)
	if len(backupSearchResult.FoundFiles) > 0 {
		logger.Info("")
		logger.Info("💾 Step 4: Asset Discovery Summary...")
		logger.Info("💾 Found %d missing assets in backup folder", len(backupSearchResult.FoundFiles))
		logger.Info("📋 Asset locations will be included in the detailed report")
		logger.Info("⚠️  Note: Asset restoration requires administrator privileges and is disabled in this version")
	} else {
		logger.Info("")
		logger.Info("❌ Step 4: No backup assets found")
	}

	// Step 5: Report Generation
	logger.Info("")
	logger.Info("📋 Step 5: Generating reports...")

	// Create missing assets map for quick lookup
	missingMap := make(map[string]bool)
	for _, hash := range missingAssets {
		missingMap[hash] = true
	}

	// Filter unique assets to only include missing ones
	var missingAssetInfos []canvus.AssetInfo
	for _, asset := range uniqueAssets {
		if missingMap[asset.Hash] {
			missingAssetInfos = append(missingAssetInfos, asset)
		}
	}

	// Generate reports
	discoverCmd := NewDiscoverCommand(cmd.config)
	err = discoverCmd.generateReports(discoveryResult, missingAssets, missingAssetInfos, backupSearchResult)
	if err != nil {
		logger.Error("Report generation failed: %v", err)
		return fmt.Errorf("report generation failed: %w", err)
	}

	// Step 6: Summary
	logger.Info("")
	logger.Info("📊 Step 6: Workflow Summary")
	logger.Info("========================================")
	logger.Info("📈 Total canvases: %d", len(discoveryResult.Canvases))
	logger.Info("🔗 Total unique assets: %d", len(uniqueAssets))
	logger.Info("📂 Local assets found: %d", len(scanResult.Files))
	logger.Info("❌ Missing assets: %d", len(missingAssets))

	if backupSearchResult != nil {
		logger.Info("💾 Assets found in backup: %d", len(backupSearchResult.FoundFiles))
		logger.Info("❌ Assets still missing: %d", len(backupSearchResult.MissingHashes))
	}

	if discoveryResult.ServerValidation != nil {
		logger.Info("🔍 Server validation:")
		logger.Info("   ✅ Assets exist on server: %d", discoveryResult.ServerValidation.ExistingAssets)
		logger.Info("   ❌ Assets missing from server: %d", discoveryResult.ServerValidation.MissingAssets)
	}

	logger.Info("")
	logger.Info("🎉 Complete workflow finished successfully!")
	logger.Info("📄 Reports generated: missing_assets_report.txt, missing_assets.csv")

	return nil
}

// shouldAutoRestore determines if assets should be auto-restored without user prompt
func (cmd *RunCommand) shouldAutoRestore() bool {
	// Check for environment variable or config setting
	// For now, always prompt - this can be enhanced later
	return false
}
