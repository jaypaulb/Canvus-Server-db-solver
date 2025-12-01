package canvus

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jaypaulb/canvus-server-db-solver/internal/logging"
	canvussdk "canvus-go-api/canvus"
)

// AssetInfo represents information about a discovered asset
type AssetInfo struct {
	Hash             string `json:"hash"`
	WidgetType       string `json:"widget_type"`
	OriginalFilename string `json:"original_filename"`
	CanvasID         string `json:"canvas_id"`
	CanvasName       string `json:"canvas_name"`
	WidgetID         string `json:"widget_id"`
	WidgetName       string `json:"widget_name"`
}

// DiscoveryResult represents the result of asset discovery
type DiscoveryResult struct {
	Assets           []AssetInfo `json:"assets"`
	AssetsWithoutHash []AssetInfo `json:"assets_without_hash"` // Assets that don't have a hash value
	Canvases         []canvussdk.Canvas `json:"canvases"`
	StartTime        time.Time   `json:"start_time"`
	EndTime          time.Time   `json:"end_time"`
	Duration         time.Duration `json:"duration"`
	Errors           []string    `json:"errors"`
	ServerValidation *ServerValidationResult `json:"server_validation,omitempty"`
}

// ServerValidationResult represents the result of server-side asset validation
type ServerValidationResult struct {
	TotalAssets     int `json:"total_assets"`
	ExistingAssets  int `json:"existing_assets"`
	MissingAssets   int `json:"missing_assets"`
	ValidationErrors []string `json:"validation_errors"`
}

// RateLimiter controls the rate of API requests
type RateLimiter struct {
	requests chan struct{}
	rate     time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rate := time.Second / time.Duration(requestsPerSecond)
	rl := &RateLimiter{
		requests: make(chan struct{}, requestsPerSecond),
		rate:     rate,
	}

	// Start the rate limiter goroutine
	go rl.run()

	return rl
}

// run manages the rate limiting
func (rl *RateLimiter) run() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.requests <- struct{}{}:
		default:
			// Channel is full, skip this tick
		}
	}
}

// Wait blocks until a request slot is available
func (rl *RateLimiter) Wait() {
	<-rl.requests
}

// DiscoveryOptions contains options for asset discovery
type DiscoveryOptions struct {
	SkipArchived bool
}

// DiscoverAllAssets discovers all media assets across all canvases using the existing SDK
func DiscoverAllAssets(session *canvussdk.Session, requestsPerSecond int) (*DiscoveryResult, error) {
	return DiscoverAllAssetsWithOptions(session, requestsPerSecond, DiscoveryOptions{})
}

// DiscoverAllAssetsWithOptions discovers all media assets with custom options
func DiscoverAllAssetsWithOptions(session *canvussdk.Session, requestsPerSecond int, options DiscoveryOptions) (*DiscoveryResult, error) {
	startTime := time.Now()
	result := &DiscoveryResult{
		StartTime: startTime,
		Assets:    make([]AssetInfo, 0),
		AssetsWithoutHash: make([]AssetInfo, 0),
		Canvases:  make([]canvussdk.Canvas, 0),
		Errors:    make([]string, 0),
	}

	ctx := context.Background()
	logger := logging.GetLogger()

	// ═══════════════════════════════════════════════════════════════════════
	// STEP 1: Fetch all canvases from Canvus Server API
	// ═══════════════════════════════════════════════════════════════════════
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("STEP 1: Fetching canvas list from Canvus Server API...")
	logger.Info("═══════════════════════════════════════════════════════════════")

	allCanvases, err := session.ListCanvases(ctx, nil)
	if err != nil {
		logger.Error("STEP 1 FAILED: Could not fetch canvases: %v", err)
		return nil, fmt.Errorf("failed to get canvases: %w", err)
	}

	logger.Info("STEP 1 COMPLETE:")
	logger.Info("   ✅ Total canvases returned by API: %d", len(allCanvases))

	// ═══════════════════════════════════════════════════════════════════════
	// STEP 2: Filter out archived canvases (InTrash=true)
	// ═══════════════════════════════════════════════════════════════════════
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("STEP 2: Filtering archived canvases (InTrash=true)...")
	logger.Info("═══════════════════════════════════════════════════════════════")

	var canvases []canvussdk.Canvas
	archivedCount := 0
	for _, canvas := range allCanvases {
		if canvas.InTrash {
			archivedCount++
			continue
		}
		canvases = append(canvases, canvas)
	}

	logger.Info("STEP 2 COMPLETE:")
	logger.Info("   ✅ Active canvases (InTrash=false): %d", len(canvases))
	logger.Info("   ⏭️  Archived canvases skipped (InTrash=true): %d", archivedCount)

	result.Canvases = canvases

	// ═══════════════════════════════════════════════════════════════════════
	// STEP 3: Process each active canvas to extract media assets
	// ═══════════════════════════════════════════════════════════════════════
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("STEP 3: Processing %d active canvases to extract media assets...", len(canvases))
	logger.Info("═══════════════════════════════════════════════════════════════")

	// Create rate limiter
	rateLimiter := NewRateLimiter(requestsPerSecond)

	// Process canvases in parallel with rate limiting
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 10) // Limit concurrent requests

	// Counters for metrics
	var processedCount int
	var successCount int
	var archivedByServerCount int
	var otherErrorCount int
	totalCanvases := len(canvases)

	for _, canvas := range canvases {
		wg.Add(1)
		go func(canvas canvussdk.Canvas) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			rateLimiter.Wait() // Rate limit

			// Extract media assets from widgets
			widgetAssets, widgetAssetsNoHash, widgetErr := extractMediaAssetsWithError(ctx, session, canvas)

			// Extract media assets from canvas background
			backgroundAssets, backgroundAssetsNoHash := extractBackgroundAssets(ctx, session, canvas)

			mu.Lock()
			processedCount++

			if widgetErr != nil {
				errStr := widgetErr.Error()
				if strings.Contains(errStr, "has been archived") || strings.Contains(errStr, "archived") {
					archivedByServerCount++
				} else {
					otherErrorCount++
				}
			} else {
				successCount++
				result.Assets = append(result.Assets, widgetAssets...)
				result.Assets = append(result.Assets, backgroundAssets...)
				result.AssetsWithoutHash = append(result.AssetsWithoutHash, widgetAssetsNoHash...)
				result.AssetsWithoutHash = append(result.AssetsWithoutHash, backgroundAssetsNoHash...)
			}

			// Progress reporting every 100 canvases
			if processedCount%100 == 0 || processedCount == totalCanvases {
				logger.Info("   Progress: %d/%d canvases processed...", processedCount, totalCanvases)
			}
			mu.Unlock()
		}(canvas)
	}

	wg.Wait()

	logger.Info("STEP 3 COMPLETE:")
	logger.Info("   ✅ Successfully processed: %d canvases", successCount)
	logger.Info("   ⏭️  Skipped (server reports archived): %d canvases", archivedByServerCount)
	logger.Info("   ⚠️  Skipped (other errors): %d canvases", otherErrorCount)
	logger.Info("   📦 Assets with hash found: %d", len(result.Assets))
	logger.Info("   📦 Assets without hash found: %d", len(result.AssetsWithoutHash))

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// ═══════════════════════════════════════════════════════════════════════
	// STEP 4: Asset validation summary
	// ═══════════════════════════════════════════════════════════════════════
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("STEP 4: Asset validation summary...")
	logger.Info("═══════════════════════════════════════════════════════════════")

	validationResult, err := validateAssetsOnServer(ctx, session, result.Assets)
	if err != nil {
		logger.Warn("Asset validation failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Asset validation failed: %v", err))
	} else {
		result.ServerValidation = validationResult
	}

	logger.Info("STEP 4 COMPLETE:")
	logger.Info("   ✅ Total unique assets: %d", validationResult.TotalAssets)
	logger.Info("   ⏱️  Discovery duration: %v", result.Duration)

	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("DISCOVERY COMPLETE")
	logger.Info("═══════════════════════════════════════════════════════════════")

	return result, nil
}

// extractMediaAssetsWithError extracts media assets from widgets and returns any error encountered
// Returns assets with hash, assets without hash, and error (if any)
func extractMediaAssetsWithError(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas) ([]AssetInfo, []AssetInfo, error) {
	var assets []AssetInfo
	var assetsWithoutHash []AssetInfo
	logger := logging.GetLogger()

	// Get all widgets for this canvas
	logger.Verbose("Getting widgets for canvas '%s' (ID: %s)", canvas.Name, canvas.ID)
	widgets, err := session.ListWidgets(ctx, canvas.ID, nil)
	if err != nil {
		return assets, assetsWithoutHash, err
	}

	logger.Verbose("Found %d widgets in canvas '%s' (ID: %s)", len(widgets), canvas.Name, canvas.ID)

	// Log the raw widget response in verbose mode
	if len(widgets) > 0 {
		logger.Verbose("Widget response for canvas '%s':", canvas.Name)
		for i, widget := range widgets {
			logger.Verbose("  Widget %d: ID=%s, Type=%s", i+1, widget.ID, widget.WidgetType)
		}
	}

	// Process each widget and extract media assets
	mediaCount := 0
	mediaCountNoHash := 0
	for _, widget := range widgets {
		// Get the specific widget details to check for hash field
		asset, assetNoHash := extractAssetFromWidget(ctx, session, canvas, widget)
		if asset != nil {
			assets = append(assets, *asset)
			mediaCount++
			logger.Verbose("Found media asset: %s (%s) - Hash: %s", asset.WidgetName, asset.WidgetType, asset.Hash)
		}
		if assetNoHash != nil {
			assetsWithoutHash = append(assetsWithoutHash, *assetNoHash)
			mediaCountNoHash++
			logger.Verbose("Found media asset without hash: %s (%s) - Filename: %s", assetNoHash.WidgetName, assetNoHash.WidgetType, assetNoHash.OriginalFilename)
		}
	}

	logger.Verbose("Extracted %d media assets (with hash) and %d media assets (without hash) from canvas '%s' (ID: %s)",
		mediaCount, mediaCountNoHash, canvas.Name, canvas.ID)
	return assets, assetsWithoutHash, nil
}

// extractMediaAssets extracts media assets from widgets by calling the generic ListWidgets endpoint
// Returns assets with hash and assets without hash separately (legacy wrapper)
func extractMediaAssets(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas) ([]AssetInfo, []AssetInfo) {
	assets, assetsWithoutHash, _ := extractMediaAssetsWithError(ctx, session, canvas)
	return assets, assetsWithoutHash
}

// extractBackgroundAssets extracts media assets from canvas background images
// Returns assets with hash and assets without hash separately
func extractBackgroundAssets(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas) ([]AssetInfo, []AssetInfo) {
	var assets []AssetInfo
	var assetsWithoutHash []AssetInfo
	logger := logging.GetLogger()

	// Get canvas background
	logger.Verbose("Getting background for canvas '%s' (ID: %s)", canvas.Name, canvas.ID)
	background, err := session.GetCanvasBackground(ctx, canvas.ID)
	if err != nil {
		logger.Verbose("Failed to get background for canvas '%s' (ID: %s): %v", canvas.Name, canvas.ID, err)
		return assets, assetsWithoutHash
	}

	// Check if background has an image with a hash
	if background.Image != nil && background.Image.Hash != "" {
		logger.Verbose("Found background image with hash: %s for canvas '%s'", background.Image.Hash, canvas.Name)

		asset := AssetInfo{
			Hash:             background.Image.Hash,
			WidgetType:       "CanvasBackground",
			OriginalFilename: "", // Background images don't have original filenames
			CanvasID:         canvas.ID,
			CanvasName:       canvas.Name,
			WidgetID:         "background", // Special ID for background
			WidgetName:       "Canvas Background",
		}

		assets = append(assets, asset)
		logger.Verbose("Found background asset: Canvas Background (CanvasBackground) - Hash: %s", background.Image.Hash)
	} else {
		logger.Verbose("No background image found for canvas '%s' (ID: %s)", canvas.Name, canvas.ID)
	}

	return assets, assetsWithoutHash
}

// validateAssetsOnServer is disabled to prevent OOM crashes
// The original implementation downloaded full asset files which caused memory exhaustion
// Asset validation will be done during the file search phase instead
func validateAssetsOnServer(ctx context.Context, session *canvussdk.Session, assets []AssetInfo) (*ServerValidationResult, error) {
	logger := logging.GetLogger()

	result := &ServerValidationResult{
		TotalAssets:     len(assets),
		ExistingAssets:  0,
		MissingAssets:   0,
		ValidationErrors: make([]string, 0),
	}

	if len(assets) == 0 {
		return result, nil
	}

	// Get unique assets by hash to avoid duplicate counting
	uniqueAssets := make(map[string]AssetInfo)
	for _, asset := range assets {
		if asset.Hash != "" {
			uniqueAssets[asset.Hash] = asset
		}
	}

	// Skip the expensive server validation - just count unique assets
	// This prevents OOM crashes from downloading thousands of large asset files
	logger.Info("⏭️  Skipping server-side asset validation (disabled to prevent OOM)")
	logger.Info("📊 Found %d unique assets with hashes across %d total asset references", len(uniqueAssets), len(assets))

	result.TotalAssets = len(uniqueAssets)
	// Mark all as "existing" since we're not validating - actual validation happens during file search
	result.ExistingAssets = len(uniqueAssets)
	result.MissingAssets = 0

	return result, nil
}

// extractAssetFromWidget extracts asset information from a widget
// Returns asset with hash (if hash exists) and asset without hash (if no hash but has filename)
// One or both may be nil
func extractAssetFromWidget(ctx context.Context, session *canvussdk.Session, canvas canvussdk.Canvas, widget canvussdk.Widget) (*AssetInfo, *AssetInfo) {
	logger := logging.GetLogger()

	// Get the specific widget details based on type
	var widgetDetails interface{}
	var err error

	logger.Verbose("Getting details for widget ID=%s, Type=%s in canvas '%s'", widget.ID, widget.WidgetType, canvas.Name)

	switch widget.WidgetType {
	case "Image":
		widgetDetails, err = session.GetImage(ctx, canvas.ID, widget.ID)
	case "Pdf":
		widgetDetails, err = session.GetPDF(ctx, canvas.ID, widget.ID)
	case "Video":
		widgetDetails, err = session.GetVideo(ctx, canvas.ID, widget.ID)
	default:
		logger.Verbose("Skipping non-media widget type: %s", widget.WidgetType)
		return nil, nil // Not a media widget type
	}

	if err != nil {
		logger.Verbose("Failed to get widget details for ID=%s, Type=%s: %v", widget.ID, widget.WidgetType, err)
		return nil, nil
	}

	// Extract hash and other fields using reflection
	hash := ""
	filename := ""
	name := ""

	if widgetValue := reflect.ValueOf(widgetDetails); widgetValue.IsValid() && !widgetValue.IsNil() {
		// Get hash field - if it exists and is not empty, this is a media asset
		if hashField := widgetValue.Elem().FieldByName("Hash"); hashField.IsValid() && hashField.CanInterface() {
			if hashStr, ok := hashField.Interface().(string); ok && hashStr != "" {
				hash = hashStr
				logger.Verbose("Found hash for widget ID=%s: %s", widget.ID, hash)
			}
		}

		// Get filename field
		if filenameField := widgetValue.Elem().FieldByName("OriginalFilename"); filenameField.IsValid() && filenameField.CanInterface() {
			if filenameStr, ok := filenameField.Interface().(string); ok {
				filename = filenameStr
				logger.Verbose("Found filename for widget ID=%s: %s", widget.ID, filename)
			}
		}

		// Get name field (could be Title, Name, etc.)
		if nameField := widgetValue.Elem().FieldByName("Title"); nameField.IsValid() && nameField.CanInterface() {
			if nameStr, ok := nameField.Interface().(string); ok {
				name = nameStr
				logger.Verbose("Found title for widget ID=%s: %s", widget.ID, name)
			}
		} else if nameField := widgetValue.Elem().FieldByName("Name"); nameField.IsValid() && nameField.CanInterface() {
			if nameStr, ok := nameField.Interface().(string); ok {
				name = nameStr
				logger.Verbose("Found name for widget ID=%s: %s", widget.ID, name)
			}
		}
	}

	var assetWithHash *AssetInfo
	var assetWithoutHash *AssetInfo

	// If hash exists, return asset with hash
	if hash != "" {
		assetWithHash = &AssetInfo{
			Hash:             hash,
			WidgetType:       widget.WidgetType,
			OriginalFilename: filename,
			CanvasID:         canvas.ID,
			CanvasName:       canvas.Name,
			WidgetID:         widget.ID,
			WidgetName:       name,
		}
		logger.Verbose("Found asset with hash for widget ID=%s: %s", widget.ID, hash)
	} else if filename != "" {
		// If no hash but has filename, return asset without hash (for database lookup)
		assetWithoutHash = &AssetInfo{
			Hash:             "", // No hash
			WidgetType:       widget.WidgetType,
			OriginalFilename: filename,
			CanvasID:         canvas.ID,
			CanvasName:       canvas.Name,
			WidgetID:         widget.ID,
			WidgetName:       name,
		}
		logger.Verbose("Found asset without hash for widget ID=%s, Type=%s - Filename: %s", widget.ID, widget.WidgetType, filename)
	} else {
		logger.Verbose("No hash or filename found for widget ID=%s, Type=%s - not a media asset", widget.ID, widget.WidgetType)
	}

	return assetWithHash, assetWithoutHash
}


// GetUniqueAssets returns unique assets (deduplicated by hash)
func (result *DiscoveryResult) GetUniqueAssets() []AssetInfo {
	hashMap := make(map[string]AssetInfo)

	for _, asset := range result.Assets {
		if _, exists := hashMap[asset.Hash]; !exists {
			hashMap[asset.Hash] = asset
		}
	}

	uniqueAssets := make([]AssetInfo, 0, len(hashMap))
	for _, asset := range hashMap {
		uniqueAssets = append(uniqueAssets, asset)
	}

	return uniqueAssets
}

// GetAssetsByCanvas groups assets by canvas
func (result *DiscoveryResult) GetAssetsByCanvas() map[string][]AssetInfo {
	canvasMap := make(map[string][]AssetInfo)

	for _, asset := range result.Assets {
		canvasMap[asset.CanvasName] = append(canvasMap[asset.CanvasName], asset)
	}

	return canvasMap
}
