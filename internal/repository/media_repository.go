package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.MediaMapping, error)
	Create(media *models.MediaMapping) error
	Update(media *models.MediaMapping) error
	Delete(id uint) error
	GetByID(id uint) (*models.MediaMapping, error)
	CountByProjectAndBucket(projectID uuid.UUID, bucket string) (int64, error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) GetByProjectID(projectID uuid.UUID) ([]models.MediaMapping, error) {
	var medias []models.MediaMapping
	err := r.db.Where("project_id = ?", projectID).Find(&medias).Error
	return medias, err
}

func (r *mediaRepository) Create(media *models.MediaMapping) error {
	return r.db.Create(media).Error
}

func (r *mediaRepository) Update(media *models.MediaMapping) error {
	return r.db.Save(media).Error
}

func (r *mediaRepository) Delete(id uint) error {
	return r.db.Delete(&models.MediaMapping{}, id).Error
}

func (r *mediaRepository) GetByID(id uint) (*models.MediaMapping, error) {
	var media models.MediaMapping
	err := r.db.First(&media, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) CountByProjectAndBucket(projectID uuid.UUID, bucket string) (int64, error) {
	var count int64
	err := r.db.Model(&models.MediaMapping{}).
		Where("project_id = ? AND bucket = ?", projectID, bucket).
		Count(&count).Error
	return count, err
}
