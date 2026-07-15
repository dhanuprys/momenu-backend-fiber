package handler

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/storage"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UploadHandler struct {
	projectRepo repository.ProjectRepository
	fileRepo    repository.FileRecordRepository
	quotaSvc    service.DiskQuotaService
}

func NewUploadHandler(
	projectRepo repository.ProjectRepository,
	fileRepo repository.FileRecordRepository,
	quotaSvc service.DiskQuotaService,
) *UploadHandler {
	return &UploadHandler{
		projectRepo: projectRepo,
		fileRepo:    fileRepo,
		quotaSvc:    quotaSvc,
	}
}

func (h *UploadHandler) Upload(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return response.JSONError(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	}

	contentType := string(c.Request().Header.ContentType())
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil || params["boundary"] == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid content type", "INVALID_PAYLOAD")
	}

	var bodyReader io.Reader
	if c.Request().BodyStream() != nil {
		bodyReader = c.Request().BodyStream()
	} else {
		bodyReader = bytes.NewReader(c.Request().Body())
	}
	
	reader := multipart.NewReader(bodyReader, params["boundary"])

	var projectIDStr, mediaType string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Failed to parse form", "INVALID_PAYLOAD")
		}

		if part.FileName() == "" {
			// Form field
			b, _ := io.ReadAll(part)
			val := string(b)
			if part.FormName() == "project_id" {
				projectIDStr = val
			} else if part.FormName() == "type" {
				mediaType = val
			}
			continue
		}

		// File part reached
		if projectIDStr == "" {
			return response.JSONError(c, fiber.StatusBadRequest, "project_id must precede file part", "INVALID_PAYLOAD")
		}
		if mediaType == "" {
			mediaType = "image"
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Invalid project_id format", "INVALID_PAYLOAD")
		}

		project, err := h.projectRepo.GetProjectByID(projectID)
		if err != nil || project == nil || project.UserID != userID {
			return response.JSONError(c, fiber.StatusForbidden, "Access denied. Project not found or you don't have permission.", "FORBIDDEN")
		}

		quotaInfo, err := h.quotaSvc.GetQuotaInfo(projectID)
		if err != nil {
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to get quota", "INTERNAL_SERVER_ERROR")
		}

		var maxAllowedSize int64 = storage.MaxImageSize
		if mediaType == "video" {
			maxAllowedSize = storage.MaxVideoSize
		}
		if quotaInfo.RemainingBytes < maxAllowedSize {
			maxAllowedSize = quotaInfo.RemainingBytes
		}
		if maxAllowedSize <= 0 {
			return response.JSONError(c, fiber.StatusBadRequest, "Quota exceeded", "QUOTA_EXCEEDED")
		}

		var fileInfo *storage.FileRecordInfo
		if mediaType == "thumbnail" {
			fileInfo, err = storage.ProcessThumbnail(part, part.FileName(), "media", maxAllowedSize)
		} else {
			fileInfo, err = storage.StreamFile(part, part.FileName(), "media", mediaType, maxAllowedSize)
		}

		if err != nil {
			if err == storage.ErrFileTooLarge {
				return response.JSONError(c, fiber.StatusBadRequest, "File too large or quota exceeded", "QUOTA_EXCEEDED")
			}
			if err == storage.ErrNotLandscape {
				return response.JSONError(c, fiber.StatusBadRequest, "Gambar thumbnail harus berformat landscape (lebar > tinggi)", "INVALID_DIMENSIONS")
			}
			if err == storage.ErrImageTooLarge {
				return response.JSONError(c, fiber.StatusBadRequest, "Dimensi gambar terlalu besar (Maks 8192x8192)", "INVALID_DIMENSIONS")
			}
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to upload file", "INTERNAL_SERVER_ERROR")
		}

		var isOptimized bool
		var optimizedSize *int64
		if mediaType == "thumbnail" {
			isOptimized = true
			optimizedSize = fileInfo.OptimizedSize
		}

		record := &models.FileRecord{
			URL:           fileInfo.URL,
			FilePath:      fileInfo.FilePath,
			OriginalName:  fileInfo.OriginalName,
			ContentType:   fileInfo.ContentType,
			Size:          fileInfo.Size,
			IsOptimized:   isOptimized,
			OptimizedSize: optimizedSize,
			MediaType:     fileInfo.MediaType,
			ProjectID:     &projectID,
			UploadedByID:  &userID,
		}

		if err := h.fileRepo.Create(record); err != nil {
			_ = storage.DeleteFile(fileInfo.URL)
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to save file metadata", "INTERNAL_SERVER_ERROR")
		}

		return response.JSONSuccess(c, fiber.StatusCreated, "File uploaded successfully", fiber.Map{
			"url":          fileInfo.URL,
			"filename":     fileInfo.OriginalName,
			"size":         fileInfo.Size,
			"content_type": fileInfo.ContentType,
		}, nil)
	}

	return response.JSONError(c, fiber.StatusBadRequest, "No file found", "INVALID_PAYLOAD")
}

func (h *UploadHandler) AdminUpload(c fiber.Ctx) error {
	contentType := string(c.Request().Header.ContentType())
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil || params["boundary"] == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid content type", "INVALID_PAYLOAD")
	}

	var bodyReader io.Reader
	if c.Request().BodyStream() != nil {
		bodyReader = c.Request().BodyStream()
	} else {
		bodyReader = bytes.NewReader(c.Request().Body())
	}
	
	reader := multipart.NewReader(bodyReader, params["boundary"])
	var mediaType string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Failed to parse form", "INVALID_PAYLOAD")
		}

		if part.FileName() == "" {
			b, _ := io.ReadAll(part)
			if part.FormName() == "type" {
				mediaType = string(b)
			}
			continue
		}

		if mediaType == "" {
			mediaType = "audio" // Default for admin
		}

		var maxAllowedSize int64 = storage.MaxImageSize
		if mediaType == "video" {
			maxAllowedSize = storage.MaxVideoSize
		}

		fileInfo, err := storage.StreamFile(part, part.FileName(), "media", mediaType, maxAllowedSize)
		if err != nil {
			if err == storage.ErrFileTooLarge {
				return response.JSONError(c, fiber.StatusBadRequest, "File too large", "BAD_REQUEST")
			}
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to upload file", "INTERNAL_SERVER_ERROR")
		}

		record := &models.FileRecord{
			URL:          fileInfo.URL,
			FilePath:     fileInfo.FilePath,
			OriginalName: fileInfo.OriginalName,
			ContentType:  fileInfo.ContentType,
			Size:         fileInfo.Size,
			MediaType:    fileInfo.MediaType,
		}

		if err := h.fileRepo.Create(record); err != nil {
			_ = storage.DeleteFile(fileInfo.URL)
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to save file metadata", "INTERNAL_SERVER_ERROR")
		}

		return response.JSONSuccess(c, fiber.StatusCreated, "File uploaded successfully", fiber.Map{
			"url":          fileInfo.URL,
			"filename":     fileInfo.OriginalName,
			"size":         fileInfo.Size,
			"content_type": fileInfo.ContentType,
		}, nil)
	}

	return response.JSONError(c, fiber.StatusBadRequest, "No file found", "INVALID_PAYLOAD")
}
