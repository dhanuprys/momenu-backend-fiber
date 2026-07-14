package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"gorm.io/gorm"
)

type MusicRepository interface {
	ListCategories() ([]models.MusicCategory, error)
	ListMusicsByCategory(categoryID uint) ([]models.Music, error)
	ListAllMusics() ([]models.Music, error)
	GetMusicByID(id uint) (*models.Music, error)
}

type musicRepository struct {
	db *gorm.DB
}

func NewMusicRepository(db *gorm.DB) MusicRepository {
	return &musicRepository{db: db}
}

func (r *musicRepository) ListCategories() ([]models.MusicCategory, error) {
	var categories []models.MusicCategory
	err := r.db.Order("\"order\" asc").Find(&categories).Error
	return categories, err
}

func (r *musicRepository) ListMusicsByCategory(categoryID uint) ([]models.Music, error) {
	var musics []models.Music
	err := r.db.Where("category_id = ?", categoryID).Order("\"order\" asc").Find(&musics).Error
	return musics, err
}

func (r *musicRepository) ListAllMusics() ([]models.Music, error) {
	var musics []models.Music
	err := r.db.Preload("Category").Order("category_id asc, \"order\" asc").Find(&musics).Error
	return musics, err
}

func (r *musicRepository) GetMusicByID(id uint) (*models.Music, error) {
	var music models.Music
	err := r.db.Preload("Category").First(&music, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &music, nil
}
