package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/jaypaulb/kpmg-db-solver/internal/commands"
	"github.com/jaypaulb/kpmg-db-solver/internal/config"
	"github.com/jaypaulb/kpmg-db-solver/internal/logging"
)

var (
	version   string = "dev"
	buildTime string = "unknown"
	goVersion string = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "kpmg-db-solver",
	Short: "KPMG DB Solver - Canvus Asset Recovery Tool",
	Long: `KPMG DB Solver is a tool for identifying and restoring missing Canvus assets.

This tool connects to a Canvus Server, identifies missing asset files,
and restores them from backup locations. It's designed to handle large-scale
deployments with thousands of canvases and tens of thousands of assets.

Key features:
- Automated asset discovery via Canvus API
- Missing asset detection through filesystem comparison
- Backup search and recovery from multiple locations
- Comprehensive reporting with detailed and CSV outputs
- Parallel processing for optimal performance`,
	Version: fmt.Sprintf("%s (built %s with %s)", version, buildTime, goVersion),
}


var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover missing assets from Canvus Server",
	Long:  `Scan the Canvus Server and identify missing asset files by comparing API data with filesystem contents.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDiscoverCommand()
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore missing assets from backup locations",
	Long:  `Search backup locations for missing assets and restore them to the active assets folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		runRestoreCommand()
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate detailed reports of missing assets",
	Long:  `Generate comprehensive reports showing missing assets grouped by canvas with detailed information.`,
	Run: func(cmd *cobra.Command, args []string) {
		runReportCommand()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run complete workflow (discover, search, restore, report)",
	Long:  `Execute the complete KPMG DB Solver workflow: discover missing assets, search backups, restore found assets, and generate reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		runRunCommand()
	},
}

var lookupHashCmd = &cobra.Command{
	Use:   "lookup-hash",
	Short: "Lookup hash values for assets without hash using database",
	Long:  `Process assets that don't have hash values by querying the PostgreSQL database for hash information, then searching for files in assets folder and backup locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		runLookupHashCommand()
	},
}

var (
	dryRunFlag  bool
	iniPathFlag string
)

func init() {
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(lookupHashCmd)
	
	lookupHashCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Run in dry-run mode (no files will be restored)")
	lookupHashCmd.Flags().StringVar(&iniPathFlag, "ini-path", "", "Path to mt-canvus-server.ini file (default: auto-detect)")
}

// Command implementations

func runDiscoverCommand() {
	fmt.Println("🔍 Asset Discovery")
	fmt.Println("==================")
	fmt.Println()

	// Load or prompt for configuration
	cfg, err := loadOrPromptConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create and execute discover command
	discoverCmd := commands.NewDiscoverCommand(cfg)
	err = discoverCmd.Execute()
	if err != nil {
		fmt.Printf("❌ Discovery failed: %v\n", err)
		os.Exit(1)
	}
}

func runRestoreCommand() {
	fmt.Println("🔄 Asset Restoration")
	fmt.Println("====================")
	fmt.Println()

	// Load or prompt for configuration
	cfg, err := loadOrPromptConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📁 Assets Folder: %s\n", cfg.Paths.AssetsFolder)
	fmt.Printf("📁 Backup Root: %s\n", cfg.Paths.BackupRootFolder)
	fmt.Println()

	fmt.Println("🚧 Asset restoration implementation in progress...")
	fmt.Println("This will:")
	fmt.Println("  1. Read missing asset list")
	fmt.Println("  2. Search backup locations (newest first)")
	fmt.Println("  3. Copy missing assets to active folder")
	fmt.Println("  4. Verify restoration success")
	fmt.Println("  5. Generate restoration report")
}

func runReportCommand() {
	fmt.Println("📊 Report Generation")
	fmt.Println("====================")
	fmt.Println()

	// Load or prompt for configuration
	cfg, err := loadOrPromptConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📁 Output Folder: %s\n", cfg.Paths.OutputFolder)
	fmt.Println()

	fmt.Println("🚧 Report generation implementation in progress...")
	fmt.Println("This will:")
	fmt.Println("  1. Read asset discovery results")
	fmt.Println("  2. Generate detailed canvas-grouped report")
	fmt.Println("  3. Create CSV export for tracking")
	fmt.Println("  4. Include restoration status")
	fmt.Println("  5. Output error summaries")
}

func runRunCommand() {
	fmt.Println("🚀 Complete Workflow")
	fmt.Println("===================")

	// Load or prompt for configuration
	cfg, err := loadOrPromptConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create and execute run command
	runCmd := commands.NewRunCommand(cfg)
	err = runCmd.Execute(nil, nil)
	if err != nil {
		fmt.Printf("❌ Workflow failed: %v\n", err)
		os.Exit(1)
	}
}

func runLookupHashCommand() {
	fmt.Println("🔍 Hash Lookup for Assets Without Hash")
	fmt.Println("======================================")
	fmt.Println()

	// Load or prompt for configuration
	cfg, err := loadOrPromptConfig()
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create and execute lookup-hash command
	lookupCmd := commands.NewLookupHashCommand(cfg)
	lookupCmd.SetDryRun(dryRunFlag)
	if iniPathFlag != "" {
		lookupCmd.SetINIPath(iniPathFlag)
	}

	err = lookupCmd.Execute(lookupHashCmd, nil)
	if err != nil {
		fmt.Printf("❌ Hash lookup failed: %v\n", err)
		os.Exit(1)
	}
}

func loadOrPromptConfig() (*config.Config, error) {
	// Try to load from config file first
	cfg, err := config.LoadConfig("")
	if err == nil {
		// Check if we actually loaded a meaningful config (not just defaults)
		if cfg.CanvusServer.Username != "" {
			fmt.Println("✅ Loaded configuration from file")
			// Initialize logging with loaded config
			if err := initLogging(cfg); err != nil {
				fmt.Printf("⚠️  Warning: Failed to initialize logging: %v\n", err)
			}
			return cfg, nil
		}
	}

	fmt.Println("📝 No configuration file found, prompting for settings...")
	fmt.Println()

	// Prompt for configuration
	prompts := config.NewInteractivePrompts()
	cfg, err = prompts.PromptForConfig()
	if err != nil {
		return nil, err
	}

	// Initialize logging with prompted config
	if err := initLogging(cfg); err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize logging: %v\n", err)
	}

	// Ask if user wants to save configuration
	if err := prompts.PromptForSaveConfig(cfg); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save configuration: %v\n", err)
	}

	return cfg, nil
}

// initLogging initializes the logging system based on configuration
func initLogging(cfg *config.Config) error {
	level := logging.ParseLogLevel(cfg.Logging.Level)

	var logFile string
	if cfg.Logging.LogToFile {
		logFile = cfg.Logging.LogFile
	}

	return logging.InitLogger(level, cfg.Logging.Verbose, logFile)
}
