package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LiveStreamRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.LiveStream, error)
	Create(stream *models.LiveStream) error
	Update(stream *models.LiveStream) error
	Delete(id uint) error
	GetByID(id uint) (*models.LiveStream, error)
}

type liveStreamRepository struct {
	db *gorm.DB
}

func NewLiveStreamRepository(db *gorm.DB) LiveStreamRepository {
	return &liveStreamRepository{db: db}
}

func (r *liveStreamRepository) GetByProjectID(projectID uuid.UUID) ([]models.LiveStream, error) {
	var streams []models.LiveStream
	err := r.db.Where("project_id = ?", projectID).Find(&streams).Error
	return streams, err
}

func (r *liveStreamRepository) Create(stream *models.LiveStream) error {
	return r.db.Create(stream).Error
}

func (r *liveStreamRepository) Update(stream *models.LiveStream) error {
	return r.db.Save(stream).Error
}

func (r *liveStreamRepository) Delete(id uint) error {
	return r.db.Delete(&models.LiveStream{}, id).Error
}

func (r *liveStreamRepository) GetByID(id uint) (*models.LiveStream, error) {
	var stream models.LiveStream
	err := r.db.First(&stream, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stream, nil
}
