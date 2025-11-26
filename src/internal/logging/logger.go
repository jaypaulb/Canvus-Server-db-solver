package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	fileLogger *log.Logger
	verbose    bool
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, verbose bool, logFile string) (*Logger, error) {
	logger := &Logger{
		level:   level,
		verbose: verbose,
	}

	// Set up console logger
	logger.logger = log.New(os.Stdout, "", 0)

	// Set up file logger if log file is specified
	if logFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Create file logger with timestamp
		logger.fileLogger = log.New(file, "", log.LstdFlags)
	}

	return logger, nil
}

// log writes a message to both console and file if appropriate
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// Check if we should log this level
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] %s: ", timestamp, level.String())

	// Log to console (always for ERROR, WARN, and INFO; only if verbose for DEBUG)
	if level >= INFO || l.verbose {
		l.logger.Printf("%s%s", prefix, message)
	}

	// Log to file if file logger is available
	if l.fileLogger != nil {
		l.fileLogger.Printf("%s%s", prefix, message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Verbose logs a verbose message (only shown in verbose mode)
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.verbose {
		l.log(INFO, "VERBOSE: "+format, args...)
	}
}

// SetVerbose sets the verbose mode
func (l *Logger) SetVerbose(verbose bool) {
	l.verbose = verbose
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// ParseLogLevel parses a log level string
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger
func InitLogger(level LogLevel, verbose bool, logFile string) error {
	var err error
	globalLogger, err = NewLogger(level, verbose, logFile)
	return err
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Create a default logger if none exists
		globalLogger, _ = NewLogger(INFO, false, "")
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Verbose(format string, args ...interface{}) {
	GetLogger().Verbose(format, args...)
}
