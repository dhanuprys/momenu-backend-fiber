package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DressCodeRepository interface {
	GetByProjectID(projectID uuid.UUID) ([]models.DressCode, error)
	Create(dressCode *models.DressCode) error
	Update(dressCode *models.DressCode) error
	Delete(id uint) error
	GetByID(id uint) (*models.DressCode, error)
}

type dressCodeRepository struct {
	db *gorm.DB
}

func NewDressCodeRepository(db *gorm.DB) DressCodeRepository {
	return &dressCodeRepository{db: db}
}

func (r *dressCodeRepository) GetByProjectID(projectID uuid.UUID) ([]models.DressCode, error) {
	var dressCodes []models.DressCode
	err := r.db.Where("project_id = ?", projectID).Find(&dressCodes).Error
	return dressCodes, err
}

func (r *dressCodeRepository) Create(dressCode *models.DressCode) error {
	return r.db.Create(dressCode).Error
}

func (r *dressCodeRepository) Update(dressCode *models.DressCode) error {
	return r.db.Save(dressCode).Error
}

func (r *dressCodeRepository) Delete(id uint) error {
	return r.db.Delete(&models.DressCode{}, id).Error
}

func (r *dressCodeRepository) GetByID(id uint) (*models.DressCode, error) {
	var dressCode models.DressCode
	err := r.db.First(&dressCode, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dressCode, nil
}
