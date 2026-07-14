package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
)

type InvitationService interface {
	GetInvitationBySlug(slug string) (*models.Project, error)
	GetOGMetadata(slug string) (*models.Project, error)
}

type invitationService struct {
	projectRepo repository.ProjectRepository
}

func NewInvitationService(projectRepo repository.ProjectRepository) InvitationService {
	return &invitationService{
		projectRepo: projectRepo,
	}
}

func (s *invitationService) GetInvitationBySlug(slug string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectBySlug(slug)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("invitation not found")
	}

	// We don't restrict by status here because the owner might want to preview it via the slug.
	// We handle status-based visibility (e.g., returning 403 if Draft/Archived) in the handler or middleware based on the user's auth status.
	// But since this is a public endpoint, if it's not published, we should return an error.
	// Wait, to allow preview, the handler will check if the user is the owner (by passing a flag or checking token).
	// Let's just return the project here.

	return project, nil
}

func (s *invitationService) GetOGMetadata(slug string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectOGMetadata(slug)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("invitation not found")
	}

	return project, nil
}
