package service

import (
	"errors"
	"fmt"

	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/google/uuid"
)

type DiskQuotaService interface {
	CheckQuota(projectID uuid.UUID, incomingFileSize int64) error
	GetQuotaInfo(projectID uuid.UUID) (*QuotaInfo, error)
}

type QuotaInfo struct {
	UsedBytes      int64   `json:"used_bytes"`
	UsedHuman      string  `json:"used_human"`
	LimitBytes     int64   `json:"limit_bytes"`
	LimitHuman     string  `json:"limit_human"`
	RemainingBytes int64   `json:"remaining_bytes"`
	RemainingHuman string  `json:"remaining_human"`
	UsagePercent   float64 `json:"usage_percent"`
}

type diskQuotaService struct {
	fileRepo    repository.FileRecordRepository
	projectRepo repository.ProjectRepository
}

func NewDiskQuotaService(fileRepo repository.FileRecordRepository, projectRepo repository.ProjectRepository) DiskQuotaService {
	return &diskQuotaService{
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
	}
}

func (s *diskQuotaService) CheckQuota(projectID uuid.UUID, incomingFileSize int64) error {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New("project not found")
	}

	usedBytes, err := s.fileRepo.GetProjectDiskUsage(projectID)
	if err != nil {
		return err
	}

	if usedBytes+incomingFileSize > project.DiskQuotaBytes {
		return fmt.Errorf("kuota penyimpanan proyek terlampaui (digunakan: %s / kuota: %s)",
			formatBytes(usedBytes), formatBytes(project.DiskQuotaBytes))
	}

	return nil
}

func (s *diskQuotaService) GetQuotaInfo(projectID uuid.UUID) (*QuotaInfo, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	usedBytes, err := s.fileRepo.GetProjectDiskUsage(projectID)
	if err != nil {
		return nil, err
	}

	remainingBytes := project.DiskQuotaBytes - usedBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}

	usagePercent := float64(0)
	if project.DiskQuotaBytes > 0 {
		usagePercent = (float64(usedBytes) / float64(project.DiskQuotaBytes)) * 100
	}

	return &QuotaInfo{
		UsedBytes:      usedBytes,
		UsedHuman:      formatBytes(usedBytes),
		LimitBytes:     project.DiskQuotaBytes,
		LimitHuman:     formatBytes(project.DiskQuotaBytes),
		RemainingBytes: remainingBytes,
		RemainingHuman: formatBytes(remainingBytes),
		UsagePercent:   usagePercent,
	}, nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
