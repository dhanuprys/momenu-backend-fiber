package repository

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRecordRepository interface {
	Create(record *models.FileRecord) error
	GetByURL(url string) (*models.FileRecord, error)
	GetByProjectID(projectID uuid.UUID) ([]models.FileRecord, error)
	GetUnoptimizedImages() ([]models.FileRecord, error)
	GetOrphanCandidates(olderThan time.Time) ([]models.FileRecord, error)
	MarkOptimized(id uint, newURL, newFilePath, newContentType string, newSize int64) error
	UpdateURL(oldURL, newURL string) error
	Delete(id uint) error
	HardDelete(id uint) error
	GetProjectDiskUsage(projectID uuid.UUID) (int64, error)
	GetDiskUsageStats() (totalSize, totalFiles, optimizedCount, spaceSaved int64, err error)
}

type fileRecordRepository struct {
	db *gorm.DB
}

func NewFileRecordRepository(db *gorm.DB) FileRecordRepository {
	return &fileRecordRepository{db: db}
}

func (r *fileRecordRepository) Create(record *models.FileRecord) error {
	return r.db.Create(record).Error
}

func (r *fileRecordRepository) GetByURL(url string) (*models.FileRecord, error) {
	var record models.FileRecord
	err := r.db.Where("url = ?", url).First(&record).Error
	return &record, err
}

func (r *fileRecordRepository) GetByProjectID(projectID uuid.UUID) ([]models.FileRecord, error) {
	var records []models.FileRecord
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Find(&records).Error
	return records, err
}

func (r *fileRecordRepository) GetUnoptimizedImages() ([]models.FileRecord, error) {
	var records []models.FileRecord
	err := r.db.Where("is_optimized = ? AND media_type = ? AND content_type NOT IN ?", false, "image", []string{"image/webp", "image/gif"}).
		Order("created_at asc").
		Limit(10).
		Find(&records).Error
	return records, err
}

func (r *fileRecordRepository) GetOrphanCandidates(olderThan time.Time) ([]models.FileRecord, error) {
	var records []models.FileRecord
	err := r.db.Where("created_at < ?", olderThan).Find(&records).Error
	return records, err
}

func (r *fileRecordRepository) MarkOptimized(id uint, newURL, newFilePath, newContentType string, newSize int64) error {
	return r.db.Model(&models.FileRecord{}).Where("id = ?", id).Updates(map[string]interface{}{
		"url":            newURL,
		"file_path":      newFilePath,
		"content_type":   newContentType,
		"optimized_size": newSize,
		"is_optimized":   true,
	}).Error
}

func (r *fileRecordRepository) UpdateURL(oldURL, newURL string) error {
	return r.db.Model(&models.FileRecord{}).Where("url = ?", oldURL).Update("url", newURL).Error
}

func (r *fileRecordRepository) Delete(id uint) error {
	return r.db.Delete(&models.FileRecord{}, id).Error
}

func (r *fileRecordRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.FileRecord{}, id).Error
}

func (r *fileRecordRepository) GetProjectDiskUsage(projectID uuid.UUID) (int64, error) {
	var totalSize int64
	err := r.db.Model(&models.FileRecord{}).Where("project_id = ?", projectID).Select("COALESCE(SUM(COALESCE(optimized_size, size)), 0)").Scan(&totalSize).Error
	return totalSize, err
}

func (r *fileRecordRepository) GetDiskUsageStats() (totalSize, totalFiles, optimizedCount, spaceSaved int64, err error) {
	type StatsResult struct {
		TotalSize      int64
		TotalFiles     int64
		OptimizedCount int64
		SpaceSaved     int64
	}

	var result StatsResult
	err = r.db.Model(&models.FileRecord{}).
		Select("COALESCE(SUM(COALESCE(optimized_size, size)), 0) as total_size, " +
			"COUNT(id) as total_files, " +
			"SUM(CASE WHEN is_optimized = true THEN 1 ELSE 0 END) as optimized_count, " +
			"COALESCE(SUM(CASE WHEN is_optimized = true THEN size - COALESCE(optimized_size, size) ELSE 0 END), 0) as space_saved").
		Scan(&result).Error

	if err != nil {
		return 0, 0, 0, 0, err
	}

	return result.TotalSize, result.TotalFiles, result.OptimizedCount, result.SpaceSaved, nil
}
