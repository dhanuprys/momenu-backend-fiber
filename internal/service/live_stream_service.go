package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type LiveStreamService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.LiveStream, error)
	Create(projectID uuid.UUID, platform string, url string) (*models.LiveStream, error)
	Update(id uint, projectID uuid.UUID, platform string, url string) (*models.LiveStream, error)
	Delete(id uint, projectID uuid.UUID) error
}

type liveStreamService struct {
	repo repository.LiveStreamRepository
}

func NewLiveStreamService(repo repository.LiveStreamRepository) LiveStreamService {
	return &liveStreamService{repo: repo}
}

func (s *liveStreamService) GetByProjectID(projectID uuid.UUID) ([]models.LiveStream, error) {
	return s.repo.GetByProjectID(projectID)
}

func (s *liveStreamService) Create(projectID uuid.UUID, platform string, url string) (*models.LiveStream, error) {
	p := models.LiveStreamPlatform(platform)
	if p != models.PlatformYouTube && p != models.PlatformInstagram && p != models.PlatformTiktok && p != models.PlatformZoom && p != models.PlatformGmeet && p != models.PlatformOther {
		return nil, errors.New("invalid live stream platform")
	}

	stream := &models.LiveStream{
		ProjectID: projectID,
		Platform:  p,
		URL:       url,
	}

	if err := s.repo.Create(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *liveStreamService) Update(id uint, projectID uuid.UUID, platform string, url string) (*models.LiveStream, error) {
	stream, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if stream == nil || stream.ProjectID != projectID {
		return nil, errors.New("live stream not found or does not belong to project")
	}

	p := models.LiveStreamPlatform(platform)
	if p != models.PlatformYouTube && p != models.PlatformInstagram && p != models.PlatformTiktok && p != models.PlatformZoom && p != models.PlatformGmeet && p != models.PlatformOther {
		return nil, errors.New("invalid live stream platform")
	}

	stream.Platform = p
	stream.URL = url

	if err := s.repo.Update(stream); err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *liveStreamService) Delete(id uint, projectID uuid.UUID) error {
	stream, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if stream == nil || stream.ProjectID != projectID {
		return errors.New("live stream not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
