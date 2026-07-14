package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GiftRegistryRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.GiftRegistry, error)
	Create(registry *models.GiftRegistry) error
	Update(registry *models.GiftRegistry) error
	Delete(id uint) error
	GetByID(id uint) (*models.GiftRegistry, error)
}

type giftRegistryRepository struct {
	db *gorm.DB
}

func NewGiftRegistryRepository(db *gorm.DB) GiftRegistryRepository {
	return &giftRegistryRepository{db: db}
}

func (r *giftRegistryRepository) GetByProjectID(projectID uuid.UUID) ([]models.GiftRegistry, error) {
	var registries []models.GiftRegistry
	err := r.db.Where("project_id = ?", projectID).Find(&registries).Error
	return registries, err
}

func (r *giftRegistryRepository) Create(registry *models.GiftRegistry) error {
	return r.db.Create(registry).Error
}

func (r *giftRegistryRepository) Update(registry *models.GiftRegistry) error {
	return r.db.Save(registry).Error
}

func (r *giftRegistryRepository) Delete(id uint) error {
	return r.db.Delete(&models.GiftRegistry{}, id).Error
}

func (r *giftRegistryRepository) GetByID(id uint) (*models.GiftRegistry, error) {
	var registry models.GiftRegistry
	err := r.db.First(&registry, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &registry, nil
}
