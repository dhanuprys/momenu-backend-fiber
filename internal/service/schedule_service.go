package service

import (
	"errors"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type ScheduleService interface {
	GetByProjectID(projectID uuid.UUID) ([]models.Schedule, error)
	Create(projectID uuid.UUID, title string, startTime time.Time, endTime time.Time, timezone string, location string, mapURL string) (*models.Schedule, error)
	Update(id uint, projectID uuid.UUID, title string, startTime time.Time, endTime time.Time, timezone string, location string, mapURL string) (*models.Schedule, error)
	Delete(id uint, projectID uuid.UUID) error
}

type scheduleService struct {
	repo repository.ScheduleRepository
}

func NewScheduleService(repo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{repo: repo}
}

func (s *scheduleService) GetByProjectID(projectID uuid.UUID) ([]models.Schedule, error) {
	return s.repo.GetByProjectID(projectID)
}

func (s *scheduleService) Create(projectID uuid.UUID, title string, startTime time.Time, endTime time.Time, timezone string, location string, mapURL string) (*models.Schedule, error) {
	schedule := &models.Schedule{
		ProjectID: projectID,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		Timezone:  timezone,
		Location:  location,
		MapURL:    mapURL,
	}

	if err := s.repo.Create(schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleService) Update(id uint, projectID uuid.UUID, title string, startTime time.Time, endTime time.Time, timezone string, location string, mapURL string) (*models.Schedule, error) {
	schedule, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if schedule == nil || schedule.ProjectID != projectID {
		return nil, errors.New("schedule not found or does not belong to project")
	}

	schedule.Title = title
	schedule.StartTime = startTime
	schedule.EndTime = endTime
	schedule.Timezone = timezone
	schedule.Location = location
	schedule.MapURL = mapURL

	if err := s.repo.Update(schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleService) Delete(id uint, projectID uuid.UUID) error {
	schedule, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if schedule == nil || schedule.ProjectID != projectID {
		return errors.New("schedule not found or does not belong to project")
	}
	return s.repo.Delete(id)
}
