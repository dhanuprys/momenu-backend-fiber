package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShareRepository interface {
	CreateSession(session *models.ProjectShareSession) error
	GetSessionByID(sessionID string) (*models.ProjectShareSession, error)
	GetSessionsByProjectID(projectID uuid.UUID) ([]models.ProjectShareSession, error)
	RevokeSession(sessionID string) error
	UpdateLastAccessed(sessionID string) error
}

type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) ShareRepository {
	return &shareRepository{db: db}
}

func (r *shareRepository) CreateSession(session *models.ProjectShareSession) error {
	return r.db.Create(session).Error
}

func (r *shareRepository) GetSessionByID(sessionID string) (*models.ProjectShareSession, error) {
	var session models.ProjectShareSession
	if err := r.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *shareRepository) GetSessionsByProjectID(projectID uuid.UUID) ([]models.ProjectShareSession, error) {
	var sessions []models.ProjectShareSession
	if err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *shareRepository) RevokeSession(sessionID string) error {
	return r.db.Model(&models.ProjectShareSession{}).Where("session_id = ?", sessionID).Update("is_revoked", true).Error
}

func (r *shareRepository) UpdateLastAccessed(sessionID string) error {
	return r.db.Model(&models.ProjectShareSession{}).Where("session_id = ?", sessionID).Update("last_accessed_at", gorm.Expr("NOW()")).Error
}
