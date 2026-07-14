package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type GuestbookService interface {
	GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.Guestbook, int64, error)
	Create(projectID uuid.UUID, name string, message string) (*models.Guestbook, error)
	Delete(id uint, projectID uuid.UUID) error
}

type guestbookService struct {
	repo        repository.GuestbookRepository
	projectRepo repository.ProjectRepository
}

func NewGuestbookService(repo repository.GuestbookRepository, projectRepo repository.ProjectRepository) GuestbookService {
	return &guestbookService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (s *guestbookService) GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.Guestbook, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetByProjectID(projectID, page, limit)
}

func (s *guestbookService) Create(projectID uuid.UUID, name string, message string) (*models.Guestbook, error) {
	// Check feature toggle
	toggle, err := s.projectRepo.GetFeatureToggleByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if toggle != nil && !toggle.ShowWishes {
		return nil, errors.New("guestbook feature is disabled for this project")
	}

	entry := &models.Guestbook{
		ProjectID: projectID,
		Name:      name,
		Message:   message,
	}

	if err := s.repo.Create(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *guestbookService) Delete(id uint, projectID uuid.UUID) error {
	entry, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if entry == nil || entry.ProjectID != projectID {
		return errors.New("guestbook entry not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
