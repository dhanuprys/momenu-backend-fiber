package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type ShareService interface {
	CreateSession(projectID uuid.UUID, expiresAt *time.Time) (*models.ProjectShareSession, error)
	ListSessions(projectID uuid.UUID) ([]models.ProjectShareSession, error)
	RevokeSession(sessionID string) error
	GetSharedData(sessionID string) (*models.Project, map[string]interface{}, error)
}

type shareService struct {
	shareRepo     repository.ShareRepository
	projectRepo   repository.ProjectRepository
	rsvpRepo      repository.RSVPRepository
	analyticsRepo repository.AnalyticsRepository
}

func NewShareService(shareRepo repository.ShareRepository, projectRepo repository.ProjectRepository, rsvpRepo repository.RSVPRepository, analyticsRepo repository.AnalyticsRepository) ShareService {
	return &shareService{
		shareRepo:     shareRepo,
		projectRepo:   projectRepo,
		rsvpRepo:      rsvpRepo,
		analyticsRepo: analyticsRepo,
	}
}

func generateShortID(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *shareService) CreateSession(projectID uuid.UUID, expiresAt *time.Time) (*models.ProjectShareSession, error) {
	// Check active session limit
	sessions, err := s.shareRepo.GetSessionsByProjectID(projectID)
	if err == nil {
		activeCount := 0
		for _, sess := range sessions {
			if !sess.IsRevoked {
				activeCount++
			}
		}
		if activeCount >= 5 {
			return nil, errors.New("maximum limit of 5 active share links reached")
		}
	}

	session := &models.ProjectShareSession{
		ProjectID: projectID,
		SessionID: "s_" + generateShortID(10), // e.g. s_1a2b3c4d5e
		ExpiresAt: expiresAt,
	}

	if err := s.shareRepo.CreateSession(session); err != nil {
		return nil, errors.New("failed to create share session")
	}

	return session, nil
}

func (s *shareService) ListSessions(projectID uuid.UUID) ([]models.ProjectShareSession, error) {
	return s.shareRepo.GetSessionsByProjectID(projectID)
}

func (s *shareService) RevokeSession(sessionID string) error {
	return s.shareRepo.RevokeSession(sessionID)
}

func (s *shareService) GetSharedData(sessionID string) (*models.Project, map[string]interface{}, error) {
	session, err := s.shareRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, nil, errors.New("session not found")
	}

	if session.IsRevoked {
		return nil, nil, errors.New("session is revoked")
	}

	if session.ExpiresAt != nil && time.Now().After(*session.ExpiresAt) {
		return nil, nil, errors.New("session is expired")
	}

	// Update last accessed in the background
	go func() {
		_ = s.shareRepo.UpdateLastAccessed(sessionID)
	}()

	project, err := s.projectRepo.GetProjectByID(session.ProjectID)
	if err != nil {
		return nil, nil, errors.New("project not found")
	}

	if project.Status == models.ProjectStatusDraft {
		return nil, nil, errors.New("PROJECT_IS_DRAFT")
	}

	// Fetch RSVP data and stats
	rsvps, _, _ := s.rsvpRepo.GetByProjectID(session.ProjectID, 1, 1000)
	attending, notAttending, pending, totalPax, _ := s.rsvpRepo.GetStatsByProjectID(session.ProjectID)

	// Fetch analytics data
	totalVisits, _ := s.analyticsRepo.GetTotalVisits(session.ProjectID)
	uniqueGuests, _ := s.analyticsRepo.GetUniqueGuests(session.ProjectID)
	dailyVisits, _ := s.analyticsRepo.GetVisitsOverTime(session.ProjectID, 7)
	sourceStats, _ := s.analyticsRepo.GetVisitsBySource(session.ProjectID)
	deviceStats, _ := s.analyticsRepo.GetVisitsByDevice(session.ProjectID)
	recentVisits, _ := s.analyticsRepo.GetRecentVisits(session.ProjectID, 10)

	analyticsSummary := map[string]interface{}{
		"total_rsvps":   len(rsvps),
		"rsvps":         rsvps,
		"session":       session,
		"rsvp_stats": map[string]interface{}{
			"attending":     attending,
			"not_attending": notAttending,
			"pending":       pending,
			"total_pax":     totalPax,
		},
		"total_visits":  totalVisits,
		"unique_guests": uniqueGuests,
		"daily_visits":  dailyVisits,
		"source_stats":  sourceStats,
		"device_stats":  deviceStats,
		"recent_visits": recentVisits,
	}

	return project, analyticsSummary, nil
}
