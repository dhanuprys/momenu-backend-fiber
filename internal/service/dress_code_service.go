package service

import (
	"encoding/json"
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DressCodeService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.DressCode, error)
	Create(projectID uuid.UUID, label string, colors []string) (*models.DressCode, error)
	Update(id uint, projectID uuid.UUID, label string, colors []string) (*models.DressCode, error)
	Delete(id uint, projectID uuid.UUID) error
}

type dressCodeService struct {
	repo repository.DressCodeRepository
}

func NewDressCodeService(repo repository.DressCodeRepository) DressCodeService {
	return &dressCodeService{repo: repo}
}

func (s *dressCodeService) GetByProjectID(projectID uuid.UUID) ([]models.DressCode, error) {
	return s.repo.GetByProjectID(projectID)
}

func (s *dressCodeService) Create(projectID uuid.UUID, label string, colors []string) (*models.DressCode, error) {
	if len(colors) == 0 {
		colors = []string{} // ensure it's not nil for json marshal
	}
	colorsBytes, err := json.Marshal(colors)
	if err != nil {
		return nil, errors.New("invalid colors format")
	}

	dressCode := &models.DressCode{
		ProjectID: projectID,
		Label:     label,
		Colors:    datatypes.JSON(colorsBytes),
	}

	if err := s.repo.Create(dressCode); err != nil {
		return nil, err
	}
	return dressCode, nil
}

func (s *dressCodeService) Update(id uint, projectID uuid.UUID, label string, colors []string) (*models.DressCode, error) {
	dressCode, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if dressCode == nil || dressCode.ProjectID != projectID {
		return nil, errors.New("dress code not found or does not belong to project")
	}

	if len(colors) == 0 {
		colors = []string{}
	}
	colorsBytes, err := json.Marshal(colors)
	if err != nil {
		return nil, errors.New("invalid colors format")
	}

	dressCode.Label = label
	dressCode.Colors = datatypes.JSON(colorsBytes)

	if err := s.repo.Update(dressCode); err != nil {
		return nil, err
	}
	return dressCode, nil
}

func (s *dressCodeService) Delete(id uint, projectID uuid.UUID) error {
	dressCode, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if dressCode == nil || dressCode.ProjectID != projectID {
		return errors.New("dress code not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
