package service

import (
	"encoding/json"
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type MediaService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.MediaMapping, error)
	Create(projectID uuid.UUID, themeID string, bucket string, url string, order int) (*models.MediaMapping, error)
	Update(id uint, projectID uuid.UUID, url string, order int) (*models.MediaMapping, error)
	Delete(id uint, projectID uuid.UUID) error
}

type mediaService struct {
	repo      repository.MediaRepository
	themeRepo repository.ThemeRepository
}

func NewMediaService(repo repository.MediaRepository, themeRepo repository.ThemeRepository) MediaService {
	return &mediaService{
		repo:      repo,
		themeRepo: themeRepo,
	}
}

func (s *mediaService) GetByProjectID(projectID uuid.UUID) ([]models.MediaMapping, error) {
	return s.repo.GetByProjectID(projectID)
}

func (s *mediaService) Create(projectID uuid.UUID, themeID string, bucket string, url string, order int) (*models.MediaMapping, error) {
	// Validate bucket against theme
	theme, err := s.themeRepo.GetThemeByID(themeID)
	if err != nil {
		return nil, err
	}
	if theme == nil {
		return nil, errors.New("theme not found")
	}

	var mediaBuckets []models.MediaBucket
	if len(theme.MediaBuckets) > 0 {
		if err := json.Unmarshal(theme.MediaBuckets, &mediaBuckets); err != nil {
			return nil, errors.New("failed to parse theme media buckets")
		}
	}

	var mediaType models.MediaType
	var maxFiles int
	var bucketFound bool
	for _, b := range mediaBuckets {
		if b.Key == bucket {
			bucketFound = true
			mediaType = b.MediaType
			maxFiles = b.MaxFiles
			break
		}
	}

	if !bucketFound {
		return nil, errors.New("invalid bucket for this theme")
	}

	existingCount, err := s.repo.CountByProjectAndBucket(projectID, bucket)
	if err != nil {
		return nil, err
	}
	if existingCount >= int64(maxFiles) {
		return nil, errors.New("maximum number of files reached for this bucket")
	}

	media := &models.MediaMapping{
		ProjectID: projectID,
		Bucket:    bucket,
		MediaType: mediaType,
		URL:       url,
		Order:     order,
	}

	if err := s.repo.Create(media); err != nil {
		return nil, err
	}
	return media, nil
}

func (s *mediaService) Update(id uint, projectID uuid.UUID, url string, order int) (*models.MediaMapping, error) {
	media, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if media == nil || media.ProjectID != projectID {
		return nil, errors.New("media mapping not found or does not belong to project")
	}

	media.URL = url
	media.Order = order

	if err := s.repo.Update(media); err != nil {
		return nil, err
	}
	return media, nil
}

func (s *mediaService) Delete(id uint, projectID uuid.UUID) error {
	media, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if media == nil || media.ProjectID != projectID {
		return errors.New("media mapping not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
