package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaypaulb/kpmg-db-solver/internal/logging"
)

// HashCatalogEntry represents a file in the hash catalog
type HashCatalogEntry struct {
	Hash         string
	FilePath     string
	Extension    string
	Size         int64
	ModifiedTime time.Time
}

// HashCatalog is a map of hash -> file location for fast lookups
type HashCatalog struct {
	entries map[string]HashCatalogEntry
	logger  *logging.Logger
}

// NewHashCatalog creates a new empty hash catalog
func NewHashCatalog() *HashCatalog {
	return &HashCatalog{
		entries: make(map[string]HashCatalogEntry),
		logger:  logging.GetLogger(),
	}
}

// BuildCatalogFromFolder recursively scans a folder and builds a hash catalog
// Each file is indexed by its filename (without extension) as the hash
func (c *HashCatalog) BuildCatalogFromFolder(folderPath string) error {
	c.logger.Info("📂 Building hash catalog from: %s", folderPath)
	startTime := time.Now()
	fileCount := 0

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.logger.Verbose("Error accessing %s: %v", path, err)
			return nil // Continue on error
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Extract hash from filename (filename without extension)
		filename := info.Name()
		ext := filepath.Ext(filename)
		hash := strings.TrimSuffix(filename, ext)

		// Skip empty hashes
		if hash == "" {
			return nil
		}

		// Add to catalog (or update if newer)
		if existing, exists := c.entries[hash]; exists {
			// Keep the newer file
			if info.ModTime().After(existing.ModifiedTime) {
				c.entries[hash] = HashCatalogEntry{
					Hash:         hash,
					FilePath:     path,
					Extension:    ext,
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
				}
			}
		} else {
			c.entries[hash] = HashCatalogEntry{
				Hash:         hash,
				FilePath:     path,
				Extension:    ext,
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
			}
			fileCount++
		}

		return nil
	})

	duration := time.Since(startTime)
	c.logger.Info("✅ Catalog built: %d unique hashes indexed in %v", len(c.entries), duration)

	return err
}

// Lookup searches for a hash in the catalog
func (c *HashCatalog) Lookup(hash string) (HashCatalogEntry, bool) {
	entry, found := c.entries[hash]
	return entry, found
}

// Size returns the number of entries in the catalog
func (c *HashCatalog) Size() int {
	return len(c.entries)
}

// Merge merges another catalog into this one
// If there are duplicates, keeps the entry with the newer modification time
func (c *HashCatalog) Merge(other *HashCatalog) {
	for hash, entry := range other.entries {
		if existing, exists := c.entries[hash]; exists {
			// Keep the newer file
			if entry.ModifiedTime.After(existing.ModifiedTime) {
				c.entries[hash] = entry
			}
		} else {
			c.entries[hash] = entry
		}
	}
}
