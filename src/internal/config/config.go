package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	CanvusServer CanvusServerConfig `mapstructure:"canvus_server"`
	Paths        PathsConfig        `mapstructure:"paths"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Performance  PerformanceConfig  `mapstructure:"performance"`
}

// CanvusServerConfig contains Canvus Server connection settings
type CanvusServerConfig struct {
	URL              string `mapstructure:"url"`                // Always localhost for local file access
	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	Timeout          int    `mapstructure:"timeout"`            // seconds
	InsecureTLS      bool   `mapstructure:"insecuretls"`        // Skip TLS certificate verification
}

// PathsConfig contains file system paths
type PathsConfig struct {
	AssetsFolder     string `mapstructure:"assetsfolder"`
	BackupRootFolder string `mapstructure:"backuprootfolder"`
	OutputFolder     string `mapstructure:"outputfolder"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Verbose    bool   `mapstructure:"verbose"`
	LogToFile  bool   `mapstructure:"logtofile"`
	LogFile    string `mapstructure:"logfile"`
}

// PerformanceConfig contains performance tuning settings
type PerformanceConfig struct {
	MaxConcurrentAPI    int `mapstructure:"maxconcurrentapi"`
	MaxConcurrentFiles  int `mapstructure:"maxconcurrentfiles"`
	APIRequestTimeout   int `mapstructure:"apirequesttimeout"`   // seconds
	FileOperationTimeout int `mapstructure:"fileoperationtimeout"` // seconds
}

// getDefaultPaths returns OS-appropriate default paths
func getDefaultPaths() PathsConfig {
	if runtime.GOOS == "linux" {
		return PathsConfig{
			AssetsFolder:     "/var/lib/mt-canvus-server/assets",
			BackupRootFolder: "/var/lib/mt-canvus-server/backups",
			OutputFolder:     "./reports",
		}
	}
	// Windows defaults
	return PathsConfig{
		AssetsFolder:     `C:\ProgramData\MultiTaction\canvus\assets`,
		BackupRootFolder: `C:\ProgramData\MultiTaction\canvus\backups`,
		OutputFolder:     "./reports",
	}
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		CanvusServer: CanvusServerConfig{
			URL:         "https://localhost:443", // Default localhost for local file access
			Username:    "",
			Password:    "",
			Timeout:     30,
			InsecureTLS: true, // Default to true for self-signed certificates
		},
		Paths: getDefaultPaths(),
		Logging: LoggingConfig{
			Level:      "info",
			Verbose:    true,  // Enable verbose mode by default to show API responses
			LogToFile:  true,  // Enable log file by default
			LogFile:    "kpmg-db-solver.log",
		},
		Performance: PerformanceConfig{
			MaxConcurrentAPI:     10,
			MaxConcurrentFiles:   20,
			APIRequestTimeout:    30,
			FileOperationTimeout: 60,
		},
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	config := DefaultConfig()

	// Reset viper to avoid issues with multiple calls
	viper.Reset()

	// Set up viper
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.kpmg-db-solver")

	// Set environment variable prefix
	viper.SetEnvPrefix("KPMG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load config file if specified
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	} else {
		// Config file found, use it
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Unmarshal into config struct using exact key matching
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Preserve defaults for empty fields
	config.preserveDefaults()

	return config, nil
}

// preserveDefaults ensures that default values are preserved when config fields are empty
func (c *Config) preserveDefaults() {
	defaults := DefaultConfig()

	// Preserve default assets folder if empty
	if c.Paths.AssetsFolder == "" {
		c.Paths.AssetsFolder = defaults.Paths.AssetsFolder
	}

	// Preserve default backup root folder if empty
	if c.Paths.BackupRootFolder == "" {
		c.Paths.BackupRootFolder = defaults.Paths.BackupRootFolder
	}

	// Preserve default output folder if empty
	if c.Paths.OutputFolder == "" {
		c.Paths.OutputFolder = defaults.Paths.OutputFolder
	}

	// Preserve default Canvus Server URL if empty
	if c.CanvusServer.URL == "" {
		c.CanvusServer.URL = defaults.CanvusServer.URL
	}

	// Preserve default timeout if zero
	if c.CanvusServer.Timeout == 0 {
		c.CanvusServer.Timeout = defaults.CanvusServer.Timeout
	}

	// Preserve default logging level if empty
	if c.Logging.Level == "" {
		c.Logging.Level = defaults.Logging.Level
	}

	// Preserve default log file if empty
	if c.Logging.LogFile == "" {
		c.Logging.LogFile = defaults.Logging.LogFile
	}

	// Preserve default performance settings if zero
	if c.Performance.MaxConcurrentAPI == 0 {
		c.Performance.MaxConcurrentAPI = defaults.Performance.MaxConcurrentAPI
	}
	if c.Performance.MaxConcurrentFiles == 0 {
		c.Performance.MaxConcurrentFiles = defaults.Performance.MaxConcurrentFiles
	}
	if c.Performance.APIRequestTimeout == 0 {
		c.Performance.APIRequestTimeout = defaults.Performance.APIRequestTimeout
	}
	if c.Performance.FileOperationTimeout == 0 {
		c.Performance.FileOperationTimeout = defaults.Performance.FileOperationTimeout
	}
}

// ValidateConfig validates the configuration
func (c *Config) Validate() error {
	// Validate Canvus Server settings
	// URL is always localhost for local file access, so we just ensure it's set
	if c.CanvusServer.URL == "" {
		c.CanvusServer.URL = "https://localhost:443" // Default to localhost with HTTPS
	}
	if c.CanvusServer.Username == "" {
		return fmt.Errorf("canvus server username is required")
	}
	if c.CanvusServer.Password == "" {
		return fmt.Errorf("canvus server password is required")
	}

	// Validate paths
	if c.Paths.AssetsFolder == "" {
		return fmt.Errorf("assets folder path is required")
	}
	if c.Paths.BackupRootFolder == "" {
		return fmt.Errorf("backup root folder path is required")
	}

	// Validate that paths exist (read-only access required)
	if !pathExists(c.Paths.AssetsFolder) {
		return fmt.Errorf("assets folder does not exist or is not accessible: %s", c.Paths.AssetsFolder)
	}
	if !pathExists(c.Paths.BackupRootFolder) {
		return fmt.Errorf("backup root folder does not exist or is not accessible: %s", c.Paths.BackupRootFolder)
	}

	// Create output folder if it doesn't exist
	if err := os.MkdirAll(c.Paths.OutputFolder, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}

	// Validate logging level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("invalid logging level: %s (must be one of: %s)",
			c.Logging.Level, strings.Join(validLevels, ", "))
	}

	// Validate performance settings
	if c.Performance.MaxConcurrentAPI < 1 {
		return fmt.Errorf("max concurrent API calls must be at least 1")
	}
	if c.Performance.MaxConcurrentFiles < 1 {
		return fmt.Errorf("max concurrent file operations must be at least 1")
	}

	return nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set viper values
	viper.Set("canvus_server", c.CanvusServer)
	viper.Set("paths", c.Paths)
	viper.Set("logging", c.Logging)
	viper.Set("performance", c.Performance)

	// Write to file
	return viper.WriteConfigAs(filename)
}

// GetCanvusAPIURL returns the full API URL
func (c *Config) GetCanvusAPIURL() string {
	url := strings.TrimSuffix(c.CanvusServer.URL, "/")
	if !strings.HasSuffix(url, "/api/v1") {
		url += "/api/v1"
	}
	return url
}

// Helper functions

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
