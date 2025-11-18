package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/jaypaulb/kpmg-db-solver/internal/logging"
)

// AssetFileRecord represents a record from the asset_files table
type AssetFileRecord struct {
	PublicHash  string
	PrivateHash string
	OriginalFilename string
}

// Client handles database operations
type Client struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewClient creates a new database client
func NewClient(config *DBConfig) (*Client, error) {
	logger := logging.GetLogger()

	connStr := config.GetConnectionString()
	logger.Verbose("Connecting to PostgreSQL database: %s@%s:%s/%s",
		config.Username, config.Host, config.Port, config.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("✅ Successfully connected to PostgreSQL database")

	return &Client{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// FindAssetByOriginalFilename searches the asset_files table for records matching the original filename
func (c *Client) FindAssetByOriginalFilename(originalFilename string) ([]AssetFileRecord, error) {
	if originalFilename == "" {
		return nil, fmt.Errorf("original filename cannot be empty")
	}

	// Try different possible column names for original_filename
	queries := []string{
		`SELECT public_hash, private_hash, original_filename FROM asset_files WHERE original_filename = $1`,
		`SELECT public_hash, private_hash, original_filename FROM asset_files WHERE original_filename = $1 COLLATE "C"`,
		`SELECT public_hash, private_hash, filename FROM asset_files WHERE filename = $1`,
		`SELECT public_hash, private_hash, name FROM asset_files WHERE name = $1`,
	}

	var records []AssetFileRecord
	var lastErr error

	for _, query := range queries {
		c.logger.Verbose("Trying query: %s", query)
		
		rows, err := c.db.Query(query, originalFilename)
		if err != nil {
			// If table doesn't exist or column doesn't exist, try next query
			if strings.Contains(err.Error(), "does not exist") || 
			   strings.Contains(err.Error(), "column") {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("query failed: %w", err)
		}

		// Try to read results
		found := false
		for rows.Next() {
			var record AssetFileRecord
			var pubHash, privHash, filename sql.NullString

			err := rows.Scan(&pubHash, &privHash, &filename)
			if err != nil {
				rows.Close()
				lastErr = err
				break
			}

			if pubHash.Valid {
				record.PublicHash = pubHash.String
			}
			if privHash.Valid {
				record.PrivateHash = privHash.String
			}
			if filename.Valid {
				record.OriginalFilename = filename.String
			}

			records = append(records, record)
			found = true
		}
		rows.Close()

		if found {
			c.logger.Verbose("Found %d records for filename: %s", len(records), originalFilename)
			return records, nil
		}
	}

	if len(records) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("no records found and query failed: %w", lastErr)
		}
		return nil, fmt.Errorf("no records found for filename: %s", originalFilename)
	}

	return records, nil
}

// FindAssetByOriginalFilenameCaseInsensitive searches case-insensitively
func (c *Client) FindAssetByOriginalFilenameCaseInsensitive(originalFilename string) ([]AssetFileRecord, error) {
	if originalFilename == "" {
		return nil, fmt.Errorf("original filename cannot be empty")
	}

	// Try case-insensitive search
	queries := []string{
		`SELECT public_hash, private_hash, original_filename FROM asset_files WHERE LOWER(original_filename) = LOWER($1)`,
		`SELECT public_hash, private_hash, filename FROM asset_files WHERE LOWER(filename) = LOWER($1)`,
		`SELECT public_hash, private_hash, name FROM asset_files WHERE LOWER(name) = LOWER($1)`,
	}

	var records []AssetFileRecord
	var lastErr error

	for _, query := range queries {
		c.logger.Verbose("Trying case-insensitive query: %s", query)
		
		rows, err := c.db.Query(query, originalFilename)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") || 
			   strings.Contains(err.Error(), "column") {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("query failed: %w", err)
		}

		found := false
		for rows.Next() {
			var record AssetFileRecord
			var pubHash, privHash, filename sql.NullString

			err := rows.Scan(&pubHash, &privHash, &filename)
			if err != nil {
				rows.Close()
				lastErr = err
				break
			}

			if pubHash.Valid {
				record.PublicHash = pubHash.String
			}
			if privHash.Valid {
				record.PrivateHash = privHash.String
			}
			if filename.Valid {
				record.OriginalFilename = filename.String
			}

			records = append(records, record)
			found = true
		}
		rows.Close()

		if found {
			c.logger.Verbose("Found %d records (case-insensitive) for filename: %s", len(records), originalFilename)
			return records, nil
		}
	}

	if len(records) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("no records found and query failed: %w", lastErr)
		}
		return nil, fmt.Errorf("no records found for filename: %s", originalFilename)
	}

	return records, nil
}

