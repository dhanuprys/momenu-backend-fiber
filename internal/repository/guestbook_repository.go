package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuestbookRepository interface {
	GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.Guestbook, int64, error)
	Create(entry *models.Guestbook) error
	Delete(id uint) error
	GetByID(id uint) (*models.Guestbook, error)
}

type guestbookRepository struct {
	db *gorm.DB
}

func NewGuestbookRepository(db *gorm.DB) GuestbookRepository {
	return &guestbookRepository{db: db}
}

func (r *guestbookRepository) GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.Guestbook, int64, error) {
	var entries []models.Guestbook
	var total int64

	query := r.db.Model(&models.Guestbook{}).Where("project_id = ?", projectID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error
	return entries, total, err
}

func (r *guestbookRepository) Create(entry *models.Guestbook) error {
	return r.db.Create(entry).Error
}

func (r *guestbookRepository) Delete(id uint) error {
	return r.db.Delete(&models.Guestbook{}, id).Error
}

func (r *guestbookRepository) GetByID(id uint) (*models.Guestbook, error) {
	var entry models.Guestbook
	err := r.db.First(&entry, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}
