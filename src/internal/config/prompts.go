package config

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// InteractivePrompts handles user input for configuration
type InteractivePrompts struct {
	reader *bufio.Reader
}

// NewInteractivePrompts creates a new interactive prompts instance
func NewInteractivePrompts() *InteractivePrompts {
	p := &InteractivePrompts{
		reader: bufio.NewReader(os.Stdin),
	}

	// Set up signal handler to restore terminal state on interrupt
	p.setupSignalHandler()

	return p
}

// setupSignalHandler sets up a signal handler to restore terminal state on interrupt
func (p *InteractivePrompts) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		// Restore terminal state
		if term.IsTerminal(int(syscall.Stdin)) {
			fmt.Print("\n")
		}
		os.Exit(1)
	}()
}

// PromptForConfig prompts the user for all required configuration
func (p *InteractivePrompts) PromptForConfig() (*Config, error) {
	fmt.Println("🔧 Canvus Server DB Solver Configuration")
	fmt.Println("=========================================")
	fmt.Println()

	config := DefaultConfig()

	// Canvus Server configuration
	fmt.Println("📡 Canvus Server Configuration")
	fmt.Println("-------------------------------")
	fmt.Println("ℹ️  Server URL: https://localhost:443 (local access only)")

	config.CanvusServer.Username = p.promptString("Username", "", true)
	config.CanvusServer.Password = p.promptPassword("Password")

	timeout := p.promptInt("API Timeout (seconds)", 30)
	config.CanvusServer.Timeout = timeout

	fmt.Println()

	// Paths configuration
	fmt.Println("📁 Paths Configuration")
	fmt.Println("----------------------")

	config.Paths.AssetsFolder = p.promptPath("Assets Folder Path", "", true)
	config.Paths.BackupRootFolder = p.promptPath("Backup Root Folder Path", "", true)
	config.Paths.OutputFolder = p.promptPath("Output Folder Path", "./output", false)

	fmt.Println()

	// Logging configuration
	fmt.Println("📝 Logging Configuration")
	fmt.Println("------------------------")

	verbose := p.promptBool("Enable verbose logging", false)
	config.Logging.Verbose = verbose

	if verbose {
		config.Logging.Level = "debug"
	}

	logToFile := p.promptBool("Log to file", false)
	config.Logging.LogToFile = logToFile

	if logToFile {
		config.Logging.LogFile = p.promptString("Log file path", "canvus-server-db-solver.log", false)
	}

	fmt.Println()

	// Performance configuration
	fmt.Println("⚡ Performance Configuration")
	fmt.Println("-----------------------------")

	maxAPI := p.promptInt("Max concurrent API calls", 10)
	config.Performance.MaxConcurrentAPI = maxAPI

	maxFiles := p.promptInt("Max concurrent file operations", 20)
	config.Performance.MaxConcurrentFiles = maxFiles

	fmt.Println()

	// Validate configuration
	fmt.Println("✅ Validating configuration...")
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	fmt.Println("✅ Configuration validated successfully!")
	fmt.Println()

	return config, nil
}

// promptString prompts for a string input
func (p *InteractivePrompts) promptString(prompt, defaultValue string, required bool) string {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		input, err := p.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			if defaultValue != "" {
				return defaultValue
			}
			if required {
				fmt.Println("❌ This field is required. Please enter a value.")
				continue
			}
			return ""
		}

		return input
	}
}

// promptPassword prompts for a password input (hidden)
func (p *InteractivePrompts) promptPassword(prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)

		// Check if stdin is a terminal
		if term.IsTerminal(int(syscall.Stdin)) {
			// Interactive mode - read password without echoing
			password, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("\nError reading password: %v\n", err)
				// Ensure terminal is in a good state
				fmt.Print("\n")
				continue
			}
			fmt.Println() // New line after hidden input

			if len(password) == 0 {
				fmt.Println("❌ Password is required. Please enter a value.")
				continue
			}

			return string(password)
		} else {
			// Non-interactive mode - read from stdin normally
			password, err := p.reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading password: %v\n", err)
				continue
			}

			password = strings.TrimSpace(password)
			if password == "" {
				fmt.Println("❌ Password is required. Please enter a value.")
				continue
			}

			return password
		}
	}
}

// promptInt prompts for an integer input
func (p *InteractivePrompts) promptInt(prompt string, defaultValue int) int {
	for {
		input := p.promptString(prompt, fmt.Sprintf("%d", defaultValue), false)

		if input == "" {
			return defaultValue
		}

		var value int
		if _, err := fmt.Sscanf(input, "%d", &value); err != nil {
			fmt.Println("❌ Please enter a valid integer.")
			continue
		}

		if value < 1 {
			fmt.Println("❌ Value must be at least 1.")
			continue
		}

		return value
	}
}

// promptBool prompts for a yes/no input
func (p *InteractivePrompts) promptBool(prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		input := p.promptString(prompt+" (y/n)", defaultStr, false)

		if input == "" {
			return defaultValue
		}

		input = strings.ToLower(input)
		switch input {
		case "y", "yes", "true", "1":
			return true
		case "n", "no", "false", "0":
			return false
		default:
			fmt.Println("❌ Please enter 'y' for yes or 'n' for no.")
			continue
		}
	}
}

// promptPath prompts for a file path input
func (p *InteractivePrompts) promptPath(prompt, defaultValue string, mustExist bool) string {
	for {
		path := p.promptString(prompt, defaultValue, true)

		// Expand tilde to home directory
		if strings.HasPrefix(path, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("❌ Error getting home directory: %v\n", err)
				continue
			}
			path = filepath.Join(home, path[1:])
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("❌ Error converting to absolute path: %v\n", err)
			continue
		}

		// Check if path exists if required
		if mustExist {
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				fmt.Printf("❌ Path does not exist: %s\n", absPath)
				continue
			}
		}

		return absPath
	}
}

// PromptForSaveConfig asks if the user wants to save the configuration
func (p *InteractivePrompts) PromptForSaveConfig(config *Config) error {
	save := p.promptBool("Save configuration to file", true)

	if !save {
		return nil
	}

	configFile := p.promptString("Configuration file path", "config.yaml", false)

	if err := config.SaveConfig(configFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Configuration saved to: %s\n", configFile)
	return nil
}

// PromptForConfirmation asks for user confirmation before proceeding
func (p *InteractivePrompts) PromptForConfirmation(message string) bool {
	return p.promptBool(message, false)
}
