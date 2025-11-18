package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaypaulb/kpmg-db-solver/internal/logging"
)

// HashSearchResult represents the result of searching for a hash in the assets folder
type HashSearchResult struct {
	Found     bool
	FilePath  string
	Hash      string
	Extension string
	Size      int64
}

// SearchHashInAssetsFolder searches for a hash in the assets folder
// The hash is expected to be in a subfolder based on the first 2 characters
// Format: {assets_folder}/{first2chars}/{hash}.{ext}
func SearchHashInAssetsFolder(assetsPath string, hash string) (*HashSearchResult, error) {
	logger := logging.GetLogger()

	if hash == "" {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash cannot be empty")
	}

	if len(hash) < 2 {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash must be at least 2 characters")
	}

	// Get first 2 characters of hash for subfolder
	subfolder := hash[:2]
	subfolderPath := filepath.Join(assetsPath, subfolder)

	logger.Verbose("Searching for hash '%s' in subfolder: %s", hash, subfolderPath)

	// Check if subfolder exists
	if _, err := os.Stat(subfolderPath); os.IsNotExist(err) {
		logger.Verbose("Subfolder does not exist: %s", subfolderPath)
		return &HashSearchResult{Found: false}, nil
	}

	// Search for files matching the hash (with any extension)
	entries, err := os.ReadDir(subfolderPath)
	if err != nil {
		return &HashSearchResult{Found: false}, fmt.Errorf("failed to read subfolder: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := filepath.Ext(filename)
		fileHash := strings.TrimSuffix(filename, ext)

		// Check if this file matches our hash
		if fileHash == hash {
			filePath := filepath.Join(subfolderPath, filename)
			info, err := entry.Info()
			if err != nil {
				logger.Verbose("Failed to get file info for %s: %v", filePath, err)
				continue
			}

			logger.Verbose("Found hash '%s' at: %s", hash, filePath)
			return &HashSearchResult{
				Found:     true,
				FilePath:  filePath,
				Hash:      hash,
				Extension: ext,
				Size:      info.Size(),
			}, nil
		}
	}

	logger.Verbose("Hash '%s' not found in subfolder: %s", hash, subfolderPath)
	return &HashSearchResult{Found: false}, nil
}

// SearchHashInAssetsFolderRecursive searches for a hash recursively in the assets folder
// This is a fallback if the first 2 chars subfolder structure doesn't match
func SearchHashInAssetsFolderRecursive(assetsPath string, hash string) (*HashSearchResult, error) {
	logger := logging.GetLogger()

	if hash == "" {
		return &HashSearchResult{Found: false}, fmt.Errorf("hash cannot be empty")
	}

	logger.Verbose("Recursively searching for hash '%s' in: %s", hash, assetsPath)

	var foundResult *HashSearchResult
	var searchErr error

	err := filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		ext := filepath.Ext(filename)
		fileHash := strings.TrimSuffix(filename, ext)

		if fileHash == hash {
			logger.Verbose("Found hash '%s' at: %s", hash, path)
			foundResult = &HashSearchResult{
				Found:     true,
				FilePath:  path,
				Hash:      hash,
				Extension: ext,
				Size:      info.Size(),
			}
			return filepath.SkipAll // Stop searching
		}

		return nil
	})

	if err != nil {
		searchErr = err
	}

	if foundResult != nil {
		return foundResult, nil
	}

	if searchErr != nil {
		return &HashSearchResult{Found: false}, searchErr
	}

	logger.Verbose("Hash '%s' not found recursively in: %s", hash, assetsPath)
	return &HashSearchResult{Found: false}, nil
}

