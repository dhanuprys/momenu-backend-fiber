package service

import (
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type AnalyticsService interface {
	RecordVisit(projectIDStr string, guestName, source, userAgent, deviceType, ipAddress string) error
	GetProjectAnalytics(projectIDStr string, userID uint) (map[string]interface{}, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	projectRepo   repository.ProjectRepository
	ipChecker     IPCheckerService
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, projectRepo repository.ProjectRepository, ipChecker IPCheckerService) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		projectRepo:   projectRepo,
		ipChecker:     ipChecker,
	}
}

func (s *analyticsService) RecordVisit(projectIDStr string, guestName, source, userAgent, deviceType, ipAddress string) error {
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return errors.New("invalid project id")
	}

	visit := &models.ProjectVisit{
		ProjectID:  projectID,
		GuestName:  guestName,
		Source:     source,
		UserAgent:  userAgent,
		DeviceType: deviceType,
		IPAddress:  ipAddress,
	}

	err = s.analyticsRepo.CreateVisit(visit)
	if err != nil {
		return err
	}

	// Lazily resolve IP to Country in a goroutine
	if ipAddress != "" && s.ipChecker != nil {
		go func(v *models.ProjectVisit) {
			country, err := s.ipChecker.GetCountryFromIP(v.IPAddress)
			if err == nil && country != "" {
				v.Country = &country
				_ = s.analyticsRepo.UpdateVisit(v)
			}
		}(visit)
	}

	return nil
}

func (s *analyticsService) GetProjectAnalytics(projectIDStr string, userID uint) (map[string]interface{}, error) {
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	// Verify ownership
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.New("failed to retrieve project")
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	if project.UserID != userID {
		return nil, errors.New("unauthorized access to project analytics")
	}

	totalVisits, err := s.analyticsRepo.GetTotalVisits(projectID)
	if err != nil {
		return nil, err
	}

	uniqueGuests, err := s.analyticsRepo.GetUniqueGuests(projectID)
	if err != nil {
		return nil, err
	}

	recentVisits, err := s.analyticsRepo.GetRecentVisits(projectID, 10)
	if err != nil {
		return nil, err
	}

	dailyVisits, err := s.analyticsRepo.GetVisitsOverTime(projectID, 7)
	if err != nil {
		return nil, err
	}

	sourceStats, err := s.analyticsRepo.GetVisitsBySource(projectID)
	if err != nil {
		return nil, err
	}

	deviceStats, err := s.analyticsRepo.GetVisitsByDevice(projectID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_visits":  totalVisits,
		"unique_guests": uniqueGuests,
		"recent_visits": recentVisits,
		"daily_visits":  dailyVisits,
		"source_stats":  sourceStats,
		"device_stats":  deviceStats,
	}, nil
}
