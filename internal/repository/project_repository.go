package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	CreateProject(project *models.Project) error
	GetProjectsByUserID(userID uint, page, limit int) ([]models.Project, int64, error)
	GetProjectByID(id uuid.UUID) (*models.Project, error)
	GetProjectBySlug(slug string) (*models.Project, error)
	GetProjectOGMetadata(slug string) (*models.Project, error)
	UpdateProject(project *models.Project) error
	DeleteProject(id uuid.UUID) error
	UpdateFeatureToggle(toggle *models.FeatureToggle) error
	GetFeatureToggleByProjectID(projectID uuid.UUID) (*models.FeatureToggle, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) CreateProject(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) GetProjectsByUserID(userID uint, page, limit int) ([]models.Project, int64, error) {
	projects := make([]models.Project, 0)
	var total int64
	
	query := r.db.Model(&models.Project{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("ProjectVisits").Offset(offset).Limit(limit).Find(&projects).Error
	return projects, total, err
}

func (r *projectRepository) GetProjectByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	// Preload necessary relations
	err := r.db.
		Preload("Theme").
		Preload("FeatureToggle").
		Preload("Schedules").
		Preload("GiftRegistries").
		Preload("MediaMappings").
		Preload("DressCodes").
		Preload("LiveStreams").
		Preload("Music").
		Preload("ProjectVisits").
		Preload("TextOverrides").
		Preload("StyleOverrides").
		First(&project, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetProjectBySlug(slug string) (*models.Project, error) {
	var project models.Project
	err := r.db.Where("slug = ?", slug).
		Preload("Theme").
		Preload("FeatureToggle").
		Preload("Schedules").
		Preload("GiftRegistries").
		Preload("MediaMappings").
		Preload("DressCodes").
		Preload("LiveStreams").
		Preload("Music").
		Preload("ProjectVisits").
		Preload("TextOverrides").
		Preload("StyleOverrides").
		First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetProjectOGMetadata(slug string) (*models.Project, error) {
	var project models.Project
	err := r.db.Select("id", "title", "sharing_thumbnail", "slug", "status").
		Where("slug = ?", slug).
		First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) UpdateProject(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) DeleteProject(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Project{}).Error
}

func (r *projectRepository) UpdateFeatureToggle(toggle *models.FeatureToggle) error {
	return r.db.Save(toggle).Error
}

func (r *projectRepository) GetFeatureToggleByProjectID(projectID uuid.UUID) (*models.FeatureToggle, error) {
	var toggle models.FeatureToggle
	err := r.db.Where("project_id = ?", projectID).First(&toggle).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &toggle, nil
}
