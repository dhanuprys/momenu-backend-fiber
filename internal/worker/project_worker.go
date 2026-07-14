package worker

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var updateChan = make(chan uuid.UUID, 5000)

// InitWorker starts the background goroutine to process project update increments.
func InitWorker() {
	go func() {
		logger.Info("Starting Project Update Counter Worker...")
		for projectID := range updateChan {
			err := database.DB.Model(&models.Project{}).
				Where("id = ?", projectID).
				UpdateColumn("update_count", gorm.Expr("update_count + ?", 1)).Error

			if err != nil {
				logger.Error("Failed to update project update_count", zap.Error(err))
			}
		}
	}()

	// Initialize File Repo for storage workers
	fileRepo := repository.NewFileRecordRepository(database.DB)

	// Start Storage Workers
	InitImageOptimizer(database.DB, fileRepo, logger.Log)
	InitOrphanCleaner(database.DB, fileRepo, logger.Log)
}

// IncrementUpdateCount queues a project to have its update counter incremented.
func IncrementUpdateCount(projectID uuid.UUID) {
	select {
	case updateChan <- projectID:
		// successfully queued
	default:
		// channel is full, drop it to avoid blocking HTTP threads
		logger.Error("Update counter channel is full, dropping increment for project: " + projectID.String())
	}
}
