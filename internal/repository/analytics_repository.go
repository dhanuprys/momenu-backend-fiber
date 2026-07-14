package repository

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	CreateVisit(visit *models.ProjectVisit) error
	GetTotalVisits(projectID uuid.UUID) (int64, error)
	GetUniqueGuests(projectID uuid.UUID) (int64, error)
	GetRecentVisits(projectID uuid.UUID, limit int) ([]models.ProjectVisit, error)
	GetVisitsOverTime(projectID uuid.UUID, days int) ([]DailyVisit, error)
	GetVisitsBySource(projectID uuid.UUID) ([]SourceStats, error)
	GetVisitsByDevice(projectID uuid.UUID) ([]DeviceStats, error)
}

type DailyVisit struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type SourceStats struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

type DeviceStats struct {
	DeviceType string `json:"device_type"`
	Count      int64  `json:"count"`
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) CreateVisit(visit *models.ProjectVisit) error {
	return r.db.Create(visit).Error
}

func (r *analyticsRepository) GetTotalVisits(projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProjectVisit{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

func (r *analyticsRepository) GetUniqueGuests(projectID uuid.UUID) (int64, error) {
	var count int64
	// Count distinct IP addresses as unique visits, or distinct guest names if present
	err := r.db.Model(&models.ProjectVisit{}).
		Where("project_id = ?", projectID).
		Group("ip_address").
		Count(&count).Error
	return count, err
}

func (r *analyticsRepository) GetRecentVisits(projectID uuid.UUID, limit int) ([]models.ProjectVisit, error) {
	visits := make([]models.ProjectVisit, 0)
	err := r.db.Where("project_id = ?", projectID).
		Order("created_at desc").
		Limit(limit).
		Find(&visits).Error
	return visits, err
}

func (r *analyticsRepository) GetVisitsOverTime(projectID uuid.UUID, days int) ([]DailyVisit, error) {
	var results []DailyVisit
	timeAgo := time.Now().AddDate(0, 0, -days)
	err := r.db.Model(&models.ProjectVisit{}).
		Select("DATE(created_at) as date, count(id) as count").
		Where("project_id = ? AND created_at >= ?", projectID, timeAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (r *analyticsRepository) GetVisitsBySource(projectID uuid.UUID) ([]SourceStats, error) {
	var results []SourceStats
	err := r.db.Model(&models.ProjectVisit{}).
		Select("source, count(id) as count").
		Where("project_id = ?", projectID).
		Group("source").
		Order("count DESC").
		Scan(&results).Error
	return results, err
}

func (r *analyticsRepository) GetVisitsByDevice(projectID uuid.UUID) ([]DeviceStats, error) {
	var results []DeviceStats
	err := r.db.Model(&models.ProjectVisit{}).
		Select("device_type, count(id) as count").
		Where("project_id = ?", projectID).
		Group("device_type").
		Order("count DESC").
		Scan(&results).Error
	return results, err
}
