package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"gorm.io/gorm"
)

type ThemeRepository interface {
	GetAllThemes(eventType string) ([]models.Theme, error)
	GetThemeByID(id string) (*models.Theme, error)
}

type themeRepository struct {
	db *gorm.DB
}

func NewThemeRepository(db *gorm.DB) ThemeRepository {
	return &themeRepository{db: db}
}

func (r *themeRepository) GetAllThemes(eventType string) ([]models.Theme, error) {
	var themes []models.Theme
	query := r.db.Model(&models.Theme{})
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	err := query.Find(&themes).Error
	return themes, err
}

func (r *themeRepository) GetThemeByID(id string) (*models.Theme, error) {
	var theme models.Theme
	err := r.db.Where("id = ?", id).First(&theme).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &theme, nil
}
