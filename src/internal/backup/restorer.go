package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jaypaulb/canvus-server-db-solver/internal/logging"
)

// Restorer handles copying backup files to the assets folder
type Restorer struct {
	assetsFolder string
	logger       *logging.Logger
}

// NewRestorer creates a new backup restorer
func NewRestorer(assetsFolder string) *Restorer {
	return &Restorer{
		assetsFolder: assetsFolder,
		logger:       logging.GetLogger(),
	}
}

// RestoreResult contains the results of a restoration operation
type RestoreResult struct {
	RestoredFiles []string // List of successfully restored files
	FailedFiles   []string // List of files that failed to restore
	TotalBytes    int64    // Total bytes restored
	Errors        []string // List of error messages
}

// RestoreAssets copies backup files to the assets folder, preserving folder structure
func (r *Restorer) RestoreAssets(searchResult *SearchResult) (*RestoreResult, error) {
	result := &RestoreResult{
		RestoredFiles: make([]string, 0),
		FailedFiles:   make([]string, 0),
		TotalBytes:    0,
		Errors:        make([]string, 0),
	}

	if len(searchResult.FoundFiles) == 0 {
		r.logger.Info("No backup files to restore")
		return result, nil
	}

	r.logger.Info("🔄 Restoring %d assets to: %s (preserving folder structure)", len(searchResult.FoundFiles), r.assetsFolder)

	// Ensure assets folder exists
	if err := os.MkdirAll(r.assetsFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create assets folder: %w", err)
	}

	// Restore each found asset
	for hash, backupFiles := range searchResult.FoundFiles {
		if len(backupFiles) == 0 {
			continue
		}

		// Use the newest backup file (first in the sorted list)
		backupFile := backupFiles[0]

		err := r.restoreSingleFile(backupFile, result)
		if err != nil {
			r.logger.Error("Failed to restore %s: %v", hash, err)
			result.FailedFiles = append(result.FailedFiles, hash)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", hash, err))
		} else {
			targetPath := r.getAssetPath(backupFile.RelativePath)
			r.logger.Verbose("✅ Restored: %s -> %s", backupFile.Path, targetPath)
		}
	}

	r.logger.Info("✅ Restoration completed:")
	r.logger.Info("   ✅ Files restored: %d", len(result.RestoredFiles))
	r.logger.Info("   ❌ Files failed: %d", len(result.FailedFiles))
	r.logger.Info("   📊 Total bytes: %d", result.TotalBytes)

	return result, nil
}

// restoreSingleFile copies a single backup file to the assets folder, preserving folder structure
func (r *Restorer) restoreSingleFile(backupFile BackupFile, result *RestoreResult) error {
	// Determine the target path (preserving relative folder structure)
	targetPath := r.getAssetPath(backupFile.RelativePath)

	// Check if target file already exists
	if r.pathExists(targetPath) {
		r.logger.Verbose("Asset already exists, skipping: %s", targetPath)
		result.RestoredFiles = append(result.RestoredFiles, backupFile.Hash)
		return nil
	}

	// Ensure the target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	// Copy the file
	err := r.copyFile(backupFile.Path, targetPath)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Update result
	result.RestoredFiles = append(result.RestoredFiles, backupFile.Hash)
	result.TotalBytes += backupFile.Size

	return nil
}

// getAssetPath returns the full path for an asset file, preserving folder structure
func (r *Restorer) getAssetPath(relativePath string) string {
	return filepath.Join(r.assetsFolder, relativePath)
}

// copyFile copies a file from source to destination
func (r *Restorer) copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure the file is written to disk
	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// pathExists checks if a path exists
func (r *Restorer) pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
