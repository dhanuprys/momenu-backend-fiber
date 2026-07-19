package worker

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var orphanCleanerTicker *time.Ticker

// InitOrphanCleaner starts the background orphan file cleaner worker.
func InitOrphanCleaner(db *gorm.DB, fileRepo repository.FileRecordRepository, logger *zap.Logger) {
	orphanCleanerTicker = time.NewTicker(1 * time.Hour)
	go func() {
		for range orphanCleanerTicker.C {
			processOrphanFiles(db, fileRepo, logger)
		}
	}()
	logger.Info("Orphan cleaner worker started")
}

// StopOrphanCleaner stops the orphan cleaner worker.
func StopOrphanCleaner() {
	if orphanCleanerTicker != nil {
		orphanCleanerTicker.Stop()
	}
}

func processOrphanFiles(db *gorm.DB, fileRepo repository.FileRecordRepository, logger *zap.Logger) {
	// Files older than 24 hours
	olderThan := time.Now().Add(-24 * time.Hour)
	
	records, err := fileRepo.GetOrphanCandidates(olderThan)
	if err != nil {
		logger.Error("Failed to fetch orphan candidates", zap.Error(err))
		return
	}

	for _, record := range records {
		if !isReferenced(db, record.URL, logger) {
			logger.Info("Deleting orphan file", zap.String("url", record.URL), zap.Int64("size", record.Size))
			
			// Soft delete record first
			if err := fileRepo.Delete(record.ID); err != nil {
				logger.Error("Failed to soft delete orphan record", zap.Uint("id", record.ID), zap.Error(err))
				continue
			}

			// Delete physical file
			if err := storage.DeleteFile(record.URL); err != nil {
				logger.Warn("Failed to delete physical file", zap.String("url", record.URL), zap.Error(err))
			}

			// Hard delete record
			if err := fileRepo.HardDelete(record.ID); err != nil {
				logger.Warn("Failed to hard delete orphan record", zap.Uint("id", record.ID), zap.Error(err))
			}
		}
	}
}

func isReferenced(db *gorm.DB, url string, logger *zap.Logger) bool {
	var count int64

	// 1. Check media_mappings
	db.Model(&models.MediaMapping{}).Where("url = ?", url).Count(&count)
	if count > 0 {
		return true
	}

	// 2. Check projects sharing_thumbnail
	db.Model(&models.Project{}).Where("sharing_thumbnail = ?", url).Count(&count)
	if count > 0 {
		return true
	}

	// 3. Check gift_registries qr_code_image
	db.Model(&models.GiftRegistry{}).Where("qr_code_image = ?", url).Count(&count)
	if count > 0 {
		return true
	}

	// 4. Check musics file_path or cover_image
	db.Model(&models.Music{}).Where("file_path = ? OR cover_image = ?", url, url).Count(&count)
	if count > 0 {
		return true
	}

	// 5. Check projects payload (JSON)
	db.Model(&models.Project{}).Where("payload::text LIKE ?", "%"+url+"%").Count(&count)
	if count > 0 {
		return true
	}

	return false
}
