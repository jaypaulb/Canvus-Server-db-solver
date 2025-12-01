package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/ini.v1"
	"github.com/jaypaulb/canvus-server-db-solver/internal/logging"
)

// DBConfig represents database configuration from mt-canvus-server.ini
type DBConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

// getDefaultINIPath returns the default INI path based on the operating system
func getDefaultINIPath() string {
	if runtime.GOOS == "linux" {
		return "/etc/MultiTaction/canvus/mt-canvus-server.ini"
	}
	// Windows default
	return `C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`
}

// LoadDBConfigFromINI loads database configuration from mt-canvus-server.ini
func LoadDBConfigFromINI(iniPath string) (*DBConfig, error) {
	logger := logging.GetLogger()

	// Default path if not provided
	if iniPath == "" {
		iniPath = getDefaultINIPath()
	}

	// Check if file exists
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("INI file not found: %s", iniPath)
	}

	logger.Verbose("Loading database configuration from: %s", iniPath)

	// Load INI file
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load INI file: %w", err)
	}

	// Try to find database section (common names: [database], [postgresql], [db])
	dbConfig := &DBConfig{
		SSLMode: "disable", // Default SSL mode
	}

	// Try different section names (including 'sql' for Canvus INI format)
	sections := []string{"sql", "database", "postgresql", "db", "Database", "PostgreSQL", "DB", "SQL"}
	var dbSection *ini.Section

	for _, sectionName := range sections {
		if section := cfg.Section(sectionName); section != nil {
			dbSection = section
			logger.Verbose("Found database section: [%s]", sectionName)
			break
		}
	}

	// If no database section found, try to find any section with database-related keys
	if dbSection == nil {
		for _, section := range cfg.Sections() {
			sectionName := section.Name()
			if section.HasKey("host") || section.HasKey("database") || section.HasKey("dbname") {
				dbSection = section
				logger.Verbose("Found database configuration in section: [%s]", sectionName)
				break
			}
		}
	}

	if dbSection == nil {
		return nil, fmt.Errorf("no database configuration section found in INI file: %s", iniPath)
	}

	// Read database configuration (try various key names)
	if dbSection.HasKey("host") {
		dbConfig.Host = dbSection.Key("host").String()
	} else if dbSection.HasKey("hostname") {
		dbConfig.Host = dbSection.Key("hostname").String()
	} else if dbSection.HasKey("server") {
		dbConfig.Host = dbSection.Key("server").String()
	}

	if dbSection.HasKey("port") {
		dbConfig.Port = dbSection.Key("port").String()
	} else {
		dbConfig.Port = "5432" // Default PostgreSQL port
	}

	if dbSection.HasKey("database") {
		dbConfig.Database = dbSection.Key("database").String()
	} else if dbSection.HasKey("databasename") {
		dbConfig.Database = dbSection.Key("databasename").String()
	} else if dbSection.HasKey("dbname") {
		dbConfig.Database = dbSection.Key("dbname").String()
	} else if dbSection.HasKey("name") {
		dbConfig.Database = dbSection.Key("name").String()
	}

	if dbSection.HasKey("username") {
		dbConfig.Username = dbSection.Key("username").String()
	} else if dbSection.HasKey("user") {
		dbConfig.Username = dbSection.Key("user").String()
	}

	if dbSection.HasKey("password") {
		dbConfig.Password = dbSection.Key("password").String()
	} else if dbSection.HasKey("pass") {
		dbConfig.Password = dbSection.Key("pass").String()
	}

	if dbSection.HasKey("sslmode") {
		dbConfig.SSLMode = dbSection.Key("sslmode").String()
	} else if dbSection.HasKey("ssl_mode") {
		dbConfig.SSLMode = dbSection.Key("ssl_mode").String()
	}

	// Validate required fields and set defaults
	if dbConfig.Host == "" {
		// Default to localhost if not specified (common for Canvus INI files)
		dbConfig.Host = "localhost"
		logger.Verbose("Database host not specified in INI, defaulting to: localhost")
	}
	if dbConfig.Database == "" {
		return nil, fmt.Errorf("database name not found in INI file: %s", iniPath)
	}
	if dbConfig.Username == "" {
		return nil, fmt.Errorf("database username not found in INI file: %s", iniPath)
	}

	logger.Verbose("Database configuration loaded: Host=%s, Port=%s, Database=%s, Username=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Username)

	return dbConfig, nil
}

// GetConnectionString returns a PostgreSQL connection string
func (c *DBConfig) GetConnectionString() string {
	parts := []string{
		fmt.Sprintf("host=%s", c.Host),
		fmt.Sprintf("port=%s", c.Port),
		fmt.Sprintf("dbname=%s", c.Database),
		fmt.Sprintf("user=%s", c.Username),
		fmt.Sprintf("sslmode=%s", c.SSLMode),
	}

	if c.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.Password))
	}

	return strings.Join(parts, " ")
}

// FindINIFile attempts to find the mt-canvus-server.ini file in common locations
func FindINIFile() (string, error) {
	var commonPaths []string

	if runtime.GOOS == "linux" {
		commonPaths = []string{
			"/etc/MultiTaction/canvus/mt-canvus-server.ini",
			"/etc/multitaction/canvus/mt-canvus-server.ini",
			"/opt/MultiTaction/canvus/mt-canvus-server.ini",
			"./mt-canvus-server.ini",
			"mt-canvus-server.ini",
		}
	} else {
		// Windows paths
		commonPaths = []string{
			`C:\ProgramData\MultiTaction\Canvus\mt-canvus-server.ini`,
			`C:\Program Files\MultiTaction\Canvus\mt-canvus-server.ini`,
			`.\mt-canvus-server.ini`,
			`mt-canvus-server.ini`,
		}
	}

	for _, path := range commonPaths {
		// Expand any environment variables
		expandedPath := os.ExpandEnv(path)
		
		// Try absolute path first
		if _, err := os.Stat(expandedPath); err == nil {
			absPath, err := filepath.Abs(expandedPath)
			if err == nil {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("mt-canvus-server.ini not found in common locations")
}

