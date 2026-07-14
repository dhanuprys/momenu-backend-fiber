package service

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProjectService interface {
	CreateProject(userID uint, title string, themeID string, musicID *uint, payload json.RawMessage) (*models.Project, []response.ValidationError, error)
	GetProjectsByUserID(userID uint, page, limit int) ([]models.Project, int64, error)
	GetProjectByID(id uuid.UUID) (*models.Project, error)
	GetProjectBySlug(slug string) (*models.Project, error)
	UpdateProject(projectID uuid.UUID, title string, slug string, payload json.RawMessage, sharingThumbnail string, musicID *uint) (*models.Project, []response.ValidationError, error)
	UpdateStatus(projectID uuid.UUID, status models.ProjectStatus, userID uint) (*models.Project, error)
	DeleteProject(id uuid.UUID) error
	GetFeatureToggle(projectID uuid.UUID) (*models.FeatureToggle, error)
	UpdateFeatureToggle(toggle *models.FeatureToggle) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
	themeRepo   repository.ThemeRepository
	userRepo    repository.UserRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, themeRepo repository.ThemeRepository, userRepo repository.UserRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		themeRepo:   themeRepo,
		userRepo:    userRepo,
	}
}

func (s *projectService) CreateProject(userID uint, title string, themeID string, musicID *uint, payload json.RawMessage) (*models.Project, []response.ValidationError, error) {
	// 1. Validate Theme exists
	theme, err := s.themeRepo.GetThemeByID(themeID)
	if err != nil {
		return nil, nil, err
	}
	if theme == nil {
		return nil, nil, errors.New("theme not found")
	}

	// 2. Validate Payload against Schema
	if models.GetFieldSchema(theme.EventType) == nil {
		return nil, nil, errors.New("unsupported event type")
	}

	// Create a map representation of the payload to pass to the validator
	var payloadMap map[string]interface{}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &payloadMap); err != nil {
			return nil, nil, errors.New("invalid json payload format")
		}
	} else {
		payloadMap = make(map[string]interface{})
	}

	validationErrors := utils.ValidatePayload(theme.EventType, payloadMap)
	if len(validationErrors) > 0 {
		return nil, validationErrors, errors.New("validation failed")
	}

	// 3. Create Project
	project := &models.Project{
		UserID:    userID,
		Title:     title,
		ThemeID:   themeID,
		EventType: theme.EventType, // Inherit from Theme
		Status:         models.ProjectStatusDraft,
		Slug:           utils.GenerateSlug(title),
		MusicID:        musicID,
		Payload:        datatypes.JSON(payload),
		DiskQuotaBytes: config.AppConfig.DefaultProjectDiskQuotaMB * 1024 * 1024,
	}

	// The repository saves it. GORM will autogenerate the ID.
	if err := s.projectRepo.CreateProject(project); err != nil {
		return nil, nil, err
	}

	// 4. Auto-create default Feature Toggle
	toggle := &models.FeatureToggle{
		ProjectID:      project.ID,
		ShowRSVP:       true,
		ShowWishes:     true,
		ShowGallery:    true,
		ShowGifts:      true,
		ShowLiveStream: false,
	}
	if err := s.projectRepo.UpdateFeatureToggle(toggle); err != nil {
		return nil, nil, err
	}

	// Reload project with relations
	p, err := s.projectRepo.GetProjectByID(project.ID)
	return p, nil, err
}

func (s *projectService) GetProjectsByUserID(userID uint, page, limit int) ([]models.Project, int64, error) {
	return s.projectRepo.GetProjectsByUserID(userID, page, limit)
}

func (s *projectService) GetProjectByID(id uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	return project, nil
}

func (s *projectService) GetProjectBySlug(slug string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectBySlug(slug)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	return project, nil
}

func (s *projectService) UpdateProject(projectID uuid.UUID, title string, slug string, payload json.RawMessage, sharingThumbnail string, musicID *uint) (*models.Project, []response.ValidationError, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, nil, err
	}
	if project == nil {
		return nil, nil, errors.New("project not found")
	}

	// Validate slug format
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(slug) {
		return nil, []response.ValidationError{{
			Field:   "slug",
			Message: "Slug can only contain lowercase letters, numbers, and hyphens",
		}}, errors.New("validation failed")
	}

	// Validate reserved slugs
	reservedSlugs := map[string]bool{
		"app": true, "auth": true, "buat-akun": true, "masuk": true,
		"preview": true, "tema": true, "api": true, "admin": true,
		"login": true, "register": true,
	}
	if reservedSlugs[slug] {
		return nil, []response.ValidationError{{
			Field:   "slug",
			Message: "This slug is reserved and cannot be used",
		}}, errors.New("validation failed")
	}

	// Check if slug is unique
	if slug != project.Slug {
		existingProject, err := s.projectRepo.GetProjectBySlug(slug)
		if err != nil {
			return nil, nil, err
		}
		if existingProject != nil {
			return nil, []response.ValidationError{{
				Field:   "slug",
				Message: "Slug is already taken",
			}}, errors.New("validation failed")
		}
		project.Slug = slug
	}

	if payload != nil {
		// Validate Payload against Schema
		if models.GetFieldSchema(project.EventType) == nil {
			return nil, nil, errors.New("unsupported event type")
		}
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(payload, &payloadMap); err != nil {
			return nil, nil, errors.New("invalid json payload format")
		}

		validationErrors := utils.ValidatePayload(project.EventType, payloadMap)
		if len(validationErrors) > 0 {
			return nil, validationErrors, errors.New("validation failed")
		}
		project.Payload = datatypes.JSON(payload)
	}

	if title != "" {
		project.Title = title
	}

	if sharingThumbnail != "" {
		project.SharingThumbnail = sharingThumbnail
	}

	project.MusicID = musicID

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, nil, err
	}

	return project, nil, nil
}

func (s *projectService) UpdateStatus(projectID uuid.UUID, status models.ProjectStatus, userID uint) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Business Rule: Only verified users can publish
	if status == models.ProjectStatusPublished {
		_, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			return nil, err
		}
		// if user == nil || !user.Verified {
		// 	return nil, errors.New("only verified users can publish projects")
		// }
	}

	project.Status = status
	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) DeleteProject(id uuid.UUID) error {
	return s.projectRepo.DeleteProject(id)
}

func (s *projectService) GetFeatureToggle(projectID uuid.UUID) (*models.FeatureToggle, error) {
	toggle, err := s.projectRepo.GetFeatureToggleByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if toggle == nil {
		return nil, errors.New("feature toggle not found")
	}
	return toggle, nil
}

func (s *projectService) UpdateFeatureToggle(toggle *models.FeatureToggle) error {
	return s.projectRepo.UpdateFeatureToggle(toggle)
}
