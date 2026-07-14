package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
)

type ThemeService interface {
	GetAllThemes(eventType string) ([]models.Theme, error)
	GetThemeByID(id string) (*models.Theme, error)
}

type themeService struct {
	repo repository.ThemeRepository
}

func NewThemeService(repo repository.ThemeRepository) ThemeService {
	return &themeService{repo: repo}
}

func (s *themeService) GetAllThemes(eventType string) ([]models.Theme, error) {
	return s.repo.GetAllThemes(eventType)
}

func (s *themeService) GetThemeByID(id string) (*models.Theme, error) {
	theme, err := s.repo.GetThemeByID(id)
	if err != nil {
		return nil, err
	}
	if theme == nil {
		return nil, errors.New("theme not found")
	}
	return theme, nil
}
