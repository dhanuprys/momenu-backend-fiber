package service

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
)

type MusicService interface {
	ListCategories() ([]models.MusicCategory, error)
	ListMusicsByCategory(categoryID uint) ([]models.Music, error)
	ListAllMusics() ([]models.Music, error)
}

type musicService struct {
	musicRepo repository.MusicRepository
}

func NewMusicService(musicRepo repository.MusicRepository) MusicService {
	return &musicService{musicRepo: musicRepo}
}

func (s *musicService) ListCategories() ([]models.MusicCategory, error) {
	return s.musicRepo.ListCategories()
}

func (s *musicService) ListMusicsByCategory(categoryID uint) ([]models.Music, error) {
	return s.musicRepo.ListMusicsByCategory(categoryID)
}

func (s *musicService) ListAllMusics() ([]models.Music, error) {
	return s.musicRepo.ListAllMusics()
}
