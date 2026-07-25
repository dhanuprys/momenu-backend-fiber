package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type JourneyService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.Journey, error)
	Create(projectID uuid.UUID, title string, date string, content string, order int) (*models.Journey, error)
	Update(id uint, projectID uuid.UUID, title string, date string, content string, order int) (*models.Journey, error)
	Delete(id uint, projectID uuid.UUID) error
}

type journeyService struct {
	repo repository.JourneyRepository
}

func NewJourneyService(repo repository.JourneyRepository) JourneyService {
	return &journeyService{repo: repo}
}

func (s *journeyService) GetByProjectID(projectID uuid.UUID) ([]models.Journey, error) {
	return s.repo.GetByProjectID(projectID)
}

func (s *journeyService) Create(projectID uuid.UUID, title string, date string, content string, order int) (*models.Journey, error) {
	journey := &models.Journey{
		ProjectID: projectID,
		Title:     title,
		Date:      date,
		Content:   content,
		Order:     order,
	}

	if err := s.repo.Create(journey); err != nil {
		return nil, err
	}
	return journey, nil
}

func (s *journeyService) Update(id uint, projectID uuid.UUID, title string, date string, content string, order int) (*models.Journey, error) {
	journey, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if journey == nil || journey.ProjectID != projectID {
		return nil, errors.New("journey not found or does not belong to project")
	}

	journey.Title = title
	journey.Date = date
	journey.Content = content
	journey.Order = order

	if err := s.repo.Update(journey); err != nil {
		return nil, err
	}
	return journey, nil
}

func (s *journeyService) Delete(id uint, projectID uuid.UUID) error {
	journey, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if journey == nil || journey.ProjectID != projectID {
		return errors.New("journey not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
