package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JourneyRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.Journey, error)
	Create(journey *models.Journey) error
	Update(journey *models.Journey) error
	Delete(id uint) error
	GetByID(id uint) (*models.Journey, error)
}

type journeyRepository struct {
	db *gorm.DB
}

func NewJourneyRepository(db *gorm.DB) JourneyRepository {
	return &journeyRepository{db: db}
}

func (r *journeyRepository) GetByProjectID(projectID uuid.UUID) ([]models.Journey, error) {
	var journeys []models.Journey
	err := r.db.Where("project_id = ?", projectID).Order("\"order\" ASC, date ASC").Find(&journeys).Error
	return journeys, err
}

func (r *journeyRepository) Create(journey *models.Journey) error {
	return r.db.Create(journey).Error
}

func (r *journeyRepository) Update(journey *models.Journey) error {
	return r.db.Save(journey).Error
}

func (r *journeyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Journey{}, id).Error
}

func (r *journeyRepository) GetByID(id uint) (*models.Journey, error) {
	var journey models.Journey
	err := r.db.First(&journey, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &journey, nil
}
