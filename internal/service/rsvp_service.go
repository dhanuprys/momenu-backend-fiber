package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type RSVPService interface {
	GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.RSVP, int64, error)
	GetAllByProjectID(projectID uuid.UUID) ([]models.RSVP, error)
	GetGuestByName(projectID uuid.UUID, name string) (*models.RSVP, error)
	Upsert(projectID uuid.UUID, name string, attending bool, guestCount int) (*models.RSVP, error)
	OwnerUpsert(projectID uuid.UUID, name string, specialMessage string, whatsapp *string) (*models.RSVP, error)
	Update(projectID uuid.UUID, id uint, name string, specialMessage string, whatsapp *string) (*models.RSVP, error)
	MarkAsOpened(projectID uuid.UUID, name string) error
	Delete(projectID uuid.UUID, id uint) error
	GetStatsByProjectID(projectID uuid.UUID) (map[string]interface{}, error)
}

type rsvpService struct {
	repo        repository.RSVPRepository
	projectRepo repository.ProjectRepository
}

func NewRSVPService(repo repository.RSVPRepository, projectRepo repository.ProjectRepository) RSVPService {
	return &rsvpService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (s *rsvpService) GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.RSVP, int64, error) {
	return s.repo.GetByProjectID(projectID, page, limit)
}

func (s *rsvpService) GetAllByProjectID(projectID uuid.UUID) ([]models.RSVP, error) {
	return s.repo.GetAllByProjectID(projectID)
}

func (s *rsvpService) GetGuestByName(projectID uuid.UUID, name string) (*models.RSVP, error) {
	return s.repo.GetByName(projectID, name)
}

func (s *rsvpService) Upsert(projectID uuid.UUID, name string, attending bool, guestCount int) (*models.RSVP, error) {
	// Check feature toggle
	toggle, err := s.projectRepo.GetFeatureToggleByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if toggle != nil && !toggle.ShowRSVP {
		return nil, errors.New("RSVP feature is disabled for this project")
	}

	// Validate guest count
	if attending && guestCount <= 0 {
		return nil, errors.New("guest count must be greater than 0 if attending")
	}
	if !attending {
		guestCount = 0
	}

	// Ensure case-insensitive upsert updates the existing row if it exists
	existingGuest, err := s.GetGuestByName(projectID, name)
	if err == nil && existingGuest != nil {
		name = existingGuest.Name // Use exact casing from DB to trigger ON CONFLICT
	} else if toggle != nil && toggle.RequireRegisteredGuest {
		return nil, errors.New("hanya tamu terdaftar yang dapat melakukan RSVP")
	}

	rsvp := &models.RSVP{
		ProjectID:   projectID,
		Name:        name,
		Attending:   attending,
		GuestCount:  guestCount,
		IsResponded: true, // Public upsert implies response
	}

	if err := s.repo.Upsert(rsvp); err != nil {
		return nil, err
	}

	return rsvp, nil
}

func (s *rsvpService) OwnerUpsert(projectID uuid.UUID, name string, specialMessage string, whatsapp *string) (*models.RSVP, error) {
	rsvp := &models.RSVP{
		ProjectID:      projectID,
		Name:           name,
		SpecialMessage: specialMessage,
		Whatsapp:       whatsapp,
		IsResponded:    false, // Default to false when owner creates it
	}

	if err := s.repo.OwnerUpsert(rsvp); err != nil {
		return nil, err
	}

	return rsvp, nil
}

func (s *rsvpService) Update(projectID uuid.UUID, id uint, name string, specialMessage string, whatsapp *string) (*models.RSVP, error) {
	rsvp := &models.RSVP{
		Name:           name,
		SpecialMessage: specialMessage,
		Whatsapp:       whatsapp,
	}

	if err := s.repo.Update(projectID, id, rsvp); err != nil {
		return nil, err
	}

	return rsvp, nil
}

func (s *rsvpService) MarkAsOpened(projectID uuid.UUID, name string) error {
	return s.repo.MarkAsOpened(projectID, name)
}

func (s *rsvpService) GetStatsByProjectID(projectID uuid.UUID) (map[string]interface{}, error) {
	attending, notAttending, pending, totalPax, err := s.repo.GetStatsByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"attending":     attending,
		"not_attending": notAttending,
		"pending":       pending,
		"total_pax":     totalPax,
	}, nil
}

func (s *rsvpService) Delete(projectID uuid.UUID, id uint) error {
	return s.repo.Delete(projectID, id)
}
