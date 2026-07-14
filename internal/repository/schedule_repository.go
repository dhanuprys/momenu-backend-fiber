package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScheduleRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.Schedule, error)
	Create(schedule *models.Schedule) error
	Update(schedule *models.Schedule) error
	Delete(id uint) error
	GetByID(id uint) (*models.Schedule, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) GetByProjectID(projectID uuid.UUID) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.db.Where("project_id = ?", projectID).Find(&schedules).Error
	return schedules, err
}

func (r *scheduleRepository) Create(schedule *models.Schedule) error {
	return r.db.Create(schedule).Error
}

func (r *scheduleRepository) Update(schedule *models.Schedule) error {
	return r.db.Save(schedule).Error
}

func (r *scheduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Schedule{}, id).Error
}

func (r *scheduleRepository) GetByID(id uint) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.db.First(&schedule, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}
