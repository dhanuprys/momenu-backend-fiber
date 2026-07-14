package handler

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/storage"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UploadHandler struct {
	projectRepo repository.ProjectRepository
}

func NewUploadHandler(projectRepo repository.ProjectRepository) *UploadHandler {
	return &UploadHandler{projectRepo: projectRepo}
}

func (h *UploadHandler) Upload(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return response.JSONError(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	}

	projectIDStr := c.FormValue("project_id")
	
	if projectIDStr == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "project_id is required", "INVALID_PAYLOAD")
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid project_id format", "INVALID_PAYLOAD")
	}

	project, err := h.projectRepo.GetProjectByID(projectID)
	if err != nil || project == nil || project.UserID != userID {
		return response.JSONError(c, fiber.StatusForbidden, "Access denied. Project not found or you don't have permission.", "FORBIDDEN")
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "File is required", "INVALID_PAYLOAD")
	}

	mediaType := c.FormValue("type", "image")
	if mediaType != "image" && mediaType != "video" && mediaType != "audio" {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid type. Must be 'image', 'video' or 'audio'", "INVALID_PAYLOAD")
	}

	// Save file using storage service
	publicURL, err := storage.SaveFile(file, "media", mediaType)
	if err != nil {
		if err == storage.ErrFileTooLarge || err == storage.ErrInvalidFileType {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to upload file", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "File uploaded successfully", fiber.Map{
		"url":           publicURL,
		"filename":      file.Filename,
		"size":          file.Size,
		"content_type":  file.Header.Get("Content-Type"),
	}, nil)
}

func (h *UploadHandler) AdminUpload(c fiber.Ctx) error {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "File is required", "INVALID_PAYLOAD")
	}

	mediaType := c.FormValue("type", "audio")
	if mediaType != "image" && mediaType != "video" && mediaType != "audio" {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid type. Must be 'image', 'video' or 'audio'", "INVALID_PAYLOAD")
	}

	// Save file using storage service
	publicURL, err := storage.SaveFile(file, "media", mediaType)
	if err != nil {
		if err == storage.ErrFileTooLarge || err == storage.ErrInvalidFileType {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to upload file", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "File uploaded successfully", fiber.Map{
		"url":           publicURL,
		"filename":      file.Filename,
		"size":          file.Size,
		"content_type":  file.Header.Get("Content-Type"),
	}, nil)
}
