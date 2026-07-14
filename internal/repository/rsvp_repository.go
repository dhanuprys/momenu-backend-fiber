package repository

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RSVPRepository interface {
	GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.RSVP, int64, error)
	GetAllByProjectID(projectID uuid.UUID) ([]models.RSVP, error)
	GetByName(projectID uuid.UUID, name string) (*models.RSVP, error)
	Upsert(rsvp *models.RSVP) error
	OwnerUpsert(rsvp *models.RSVP) error
	MarkAsOpened(projectID uuid.UUID, name string) error
	Delete(projectID uuid.UUID, id uint) error
	GetStatsByProjectID(projectID uuid.UUID) (int64, int64, int64, int64, error)
}

type rsvpRepository struct {
	db *gorm.DB
}

func NewRSVPRepository(db *gorm.DB) RSVPRepository {
	return &rsvpRepository{db: db}
}

func (r *rsvpRepository) GetByProjectID(projectID uuid.UUID, page, limit int) ([]models.RSVP, int64, error) {
	var rsvps []models.RSVP
	var total int64

	query := r.db.Model(&models.RSVP{}).Where("project_id = ?", projectID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&rsvps).Error
	return rsvps, total, err
}

func (r *rsvpRepository) GetAllByProjectID(projectID uuid.UUID) ([]models.RSVP, error) {
	var rsvps []models.RSVP
	err := r.db.Model(&models.RSVP{}).Where("project_id = ?", projectID).Find(&rsvps).Error
	return rsvps, err
}

func (r *rsvpRepository) GetByName(projectID uuid.UUID, name string) (*models.RSVP, error) {
	var rsvp models.RSVP
	err := r.db.Where("project_id = ? AND LOWER(name) = LOWER(?)", projectID, name).First(&rsvp).Error
	if err != nil {
		return nil, err
	}
	return &rsvp, nil
}

func (r *rsvpRepository) Upsert(rsvp *models.RSVP) error {
	// Public upsert: updates attending, guest_count, and is_responded
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"attending", "guest_count", "is_responded"}),
	}).Create(rsvp).Error
}

func (r *rsvpRepository) OwnerUpsert(rsvp *models.RSVP) error {
	// Owner upsert: updates special_message and leaves attending/guest_count/is_responded alone if already exists
	// Wait, if it exists, the owner might just be updating the special message.
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"special_message", "whatsapp"}),
	}).Create(rsvp).Error
}

func (r *rsvpRepository) MarkAsOpened(projectID uuid.UUID, name string) error {
	return r.db.Model(&models.RSVP{}).
		Where("project_id = ? AND LOWER(name) = LOWER(?)", projectID, name).
		Update("has_opened", true).Error
}

func (r *rsvpRepository) GetStatsByProjectID(projectID uuid.UUID) (int64, int64, int64, int64, error) {
	var attending, notAttending, pending int64

	err := r.db.Model(&models.RSVP{}).Where("project_id = ? AND attending = ? AND is_responded = ?", projectID, true, true).Count(&attending).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}

	err = r.db.Model(&models.RSVP{}).Where("project_id = ? AND attending = ? AND is_responded = ?", projectID, false, true).Count(&notAttending).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}
	
	err = r.db.Model(&models.RSVP{}).Where("project_id = ? AND is_responded = ?", projectID, false).Count(&pending).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Calculate total pax using sum
	type Result struct {
		TotalPax int64
	}
	var res Result
	err = r.db.Model(&models.RSVP{}).Where("project_id = ? AND attending = ? AND is_responded = ?", projectID, true, true).Select("sum(guest_count) as total_pax").Scan(&res).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return attending, notAttending, pending, res.TotalPax, nil
}

func (r *rsvpRepository) Delete(projectID uuid.UUID, id uint) error {
	return r.db.Where("project_id = ? AND id = ?", projectID, id).Delete(&models.RSVP{}).Error
}
