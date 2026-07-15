package worker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var optimizerTicker *time.Ticker

// InitImageOptimizer starts the background image optimization worker.
func InitImageOptimizer(db *gorm.DB, fileRepo repository.FileRecordRepository, logger *zap.Logger) {
	optimizerTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for range optimizerTicker.C {
			processUnoptimizedImages(db, fileRepo, logger)
		}
	}()
	logger.Info("Image optimizer worker started")
}

// StopImageOptimizer stops the optimization worker.
func StopImageOptimizer() {
	if optimizerTicker != nil {
		optimizerTicker.Stop()
	}
}

func processUnoptimizedImages(db *gorm.DB, fileRepo repository.FileRecordRepository, logger *zap.Logger) {
	logger.Info("Photo optimizer worker started cycle")
	records, err := fileRepo.GetUnoptimizedImages()
	if err != nil {
		logger.Error("Failed to fetch unoptimized images", zap.Error(err))
		return
	}

	if len(records) == 0 {
		logger.Info("No unoptimized images found in this cycle")
		return
	}

	logger.Info("Found unoptimized images to process", zap.Int("count", len(records)))

	for _, record := range records {
		optimizeSingleImage(db, fileRepo, record, logger)
	}
	
	logger.Info("Photo optimizer worker completed cycle")
}

func optimizeSingleImage(db *gorm.DB, fileRepo repository.FileRecordRepository, record models.FileRecord, logger *zap.Logger) {
	inputPath := record.FilePath
	logger.Info("Starting optimization for image", zap.String("path", inputPath))

	// Ensure the file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		logger.Warn("File not found for optimization", zap.String("path", inputPath))
		// We might want to mark it as optimized or delete the record if file is missing
		return
	}

	// Prepare output paths
	ext := filepath.Ext(inputPath)
	outPath := strings.TrimSuffix(inputPath, ext) + ".webp"

	// Compute new URL (assuming URL matches file path structure)
	outURL := strings.TrimSuffix(record.URL, ext) + ".webp"

	// Run cwebp with memory limits and timeout
	qualityStr := strconv.FormatInt(config.AppConfig.ImageOptimizationQuality, 10)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "cwebp", "-q", qualityStr, "-low_memory", inputPath, "-o", outPath)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("cwebp timed out", zap.String("input", inputPath), zap.Error(ctx.Err()))
		} else {
			logger.Error("cwebp failed", zap.String("input", inputPath), zap.Error(err))
		}
		return
	}

	// Check new file size
	fileInfo, err := os.Stat(outPath)
	if err != nil {
		logger.Error("Failed to stat optimized file", zap.String("output", outPath), zap.Error(err))
		return
	}
	newSize := fileInfo.Size()

	// Update DB in a transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Update FileRecord
		if err := tx.Exec(`
			UPDATE file_records 
			SET is_optimized = true, url = ?, file_path = ?, content_type = ?, optimized_size = ?
			WHERE id = ?`,
			outURL, outPath, "image/webp", newSize, record.ID).Error; err != nil {
			return err
		}

		// 2. Update media mappings
		if err := tx.Exec(`UPDATE media_mappings SET url = ? WHERE url = ?`, outURL, record.URL).Error; err != nil {
			return err
		}

		// 3. Update project sharing thumbnail
		if err := tx.Exec(`UPDATE projects SET sharing_thumbnail = ? WHERE sharing_thumbnail = ?`, outURL, record.URL).Error; err != nil {
			return err
		}

		// 4. Update gift registry QR code
		if err := tx.Exec(`UPDATE gift_registries SET qr_code_image = ? WHERE qr_code_image = ?`, outURL, record.URL).Error; err != nil {
			return err
		}

		// 5. Update JSON payload (replace exact URL string in JSON text)
		// This is a naive but effective way for JSONB in postgres.
		// For MySQL, REPLACE on JSON might be tricky, but we can do a text replacement.
		if err := tx.Exec(`
			UPDATE projects 
			SET payload = REPLACE(payload::text, ?, ?)::jsonb 
			WHERE payload::text LIKE '%' || ? || '%'`,
			record.URL, outURL, record.URL).Error; err != nil {
			// If dialect doesn't support this (e.g. SQLite in tests), we might need a fallback or ignore payload for now
			// Let's log and ignore if it fails, or maybe just log.
			logger.Warn("Failed to update payload JSON (might be unsupported in dialect)", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to update DB after optimization", zap.Error(err))
		// Cleanup the newly created webp file since DB update failed
		_ = os.Remove(outPath)
		return
	}

	// DB updated successfully, delete the original file
	if err := os.Remove(inputPath); err != nil {
		logger.Warn("Failed to delete original file after optimization", zap.String("path", inputPath), zap.Error(err))
	}

	savings := record.Size - newSize
	logger.Info("Image optimized successfully",
		zap.String("original", inputPath),
		zap.String("optimized", outPath),
		zap.Int64("saved_bytes", savings))
}
