package handler

import (
	"fmt"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type DiskHandler struct {
	fileRepo    repository.FileRecordRepository
	projectRepo repository.ProjectRepository
	quotaSvc    service.DiskQuotaService
	db          *gorm.DB
}

func NewDiskHandler(
	fileRepo repository.FileRecordRepository,
	projectRepo repository.ProjectRepository,
	quotaSvc service.DiskQuotaService,
	db *gorm.DB,
) *DiskHandler {
	return &DiskHandler{
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
		quotaSvc:    quotaSvc,
		db:          db,
	}
}

// GetGlobalStats returns overall disk usage statistics (Admin only)
func (h *DiskHandler) GetGlobalStats(c fiber.Ctx) error {
	totalSize, totalFiles, optimizedCount, spaceSaved, err := h.fileRepo.GetDiskUsageStats()
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to get disk stats", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Disk stats retrieved successfully", fiber.Map{
		"total_size_bytes":  totalSize,
		"total_size_human":  formatBytes(totalSize),
		"total_files":       totalFiles,
		"optimized_files":   optimizedCount,
		"unoptimized_files": totalFiles - optimizedCount,
		"space_saved_bytes": spaceSaved,
		"space_saved_human": formatBytes(spaceSaved),
	}, nil)
}

// GetProjectDiskUsage returns disk usage for a specific project (Project Owner)
func (h *DiskHandler) GetProjectDiskUsage(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	quotaInfo, err := h.quotaSvc.GetQuotaInfo(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to get project quota info", "INTERNAL_SERVER_ERROR")
	}

	files, err := h.fileRepo.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to get project files", "INTERNAL_SERVER_ERROR")
	}

	// Prepare simplified file response for frontend
	type FileResponse struct {
		ID           uint   `json:"id"`
		URL          string `json:"url"`
		OriginalName string `json:"original_name"`
		ContentType  string `json:"content_type"`
		Size         int64  `json:"size"`
		IsOptimized  bool   `json:"is_optimized"`
		MediaType    string `json:"media_type"`
		CreatedAt    string `json:"created_at"`
	}
	
	fileResponses := make([]FileResponse, 0, len(files))
	for _, f := range files {
		fileResponses = append(fileResponses, FileResponse{
			ID:           f.ID,
			URL:          f.URL,
			OriginalName: f.OriginalName,
			ContentType:  f.ContentType,
			Size:         f.Size,
			IsOptimized:  f.IsOptimized,
			MediaType:    f.MediaType,
			CreatedAt:    f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Project disk usage retrieved successfully", fiber.Map{
		"quota": quotaInfo,
		"files": fileResponses,
	}, nil)
}

// UpdateProjectDiskQuota updates the disk quota limit for a project (Admin only)
func (h *DiskHandler) UpdateProjectDiskQuota(c fiber.Ctx) error {
	projectIDStr := c.Params("projectId")
	
	var req struct {
		DiskQuotaMB int64 `json:"disk_quota_mb" validate:"required,min=1"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
	}

	project, err := h.projectRepo.GetProjectBySlug(projectIDStr) // or GetProjectByID based on param
	if err != nil || project == nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "NOT_FOUND")
	}

	newQuotaBytes := req.DiskQuotaMB * 1024 * 1024
	
	if err := h.db.Model(&models.Project{}).Where("id = ?", project.ID).Update("disk_quota_bytes", newQuotaBytes).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update disk quota", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Project disk quota updated successfully", fiber.Map{
		"new_disk_quota_mb": req.DiskQuotaMB,
		"new_disk_quota_bytes": newQuotaBytes,
	}, nil)
}

// formatBytes formats bytes into human-readable strings
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
