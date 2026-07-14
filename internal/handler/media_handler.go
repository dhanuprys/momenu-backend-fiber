package handler

import (
	"strconv"
	"strings"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/storage"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type MediaHandler struct {
	mediaService service.MediaService
	fileRepo     repository.FileRecordRepository
	quotaSvc     service.DiskQuotaService
}

func NewMediaHandler(
	mediaService service.MediaService,
	fileRepo repository.FileRecordRepository,
	quotaSvc service.DiskQuotaService,
) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
		fileRepo:     fileRepo,
		quotaSvc:     quotaSvc,
	}
}

type MediaCreateRequest struct {
	Bucket string `json:"bucket" validate:"required"`
	URL    string `json:"url" validate:"required"`
	Order  int    `json:"order"`
}

type MediaUpdateRequest struct {
	URL   string `json:"url" validate:"required"`
	Order int    `json:"order"`
}

func (h *MediaHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	medias, err := h.mediaService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve media mappings", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Media mappings retrieved successfully", medias, nil)
}

func (h *MediaHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var bucket string
	var url string
	var order int

	if c.Is("json") || strings.Contains(string(c.Request().Header.ContentType()), "application/json") {
		var req MediaCreateRequest
		if err := c.Bind().Body(&req); err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		}
		if errors := utils.ValidateStruct(req); errors != nil {
			return response.JSONValidationError(c, errors)
		}
		bucket = req.Bucket
		url = req.URL
		order = req.Order
	} else {
		bucket = c.FormValue("bucket")
		orderStr := c.FormValue("order")
		order, _ = strconv.Atoi(orderStr)

		if bucket == "" {
			return response.JSONError(c, fiber.StatusBadRequest, "bucket is required", "INVALID_PAYLOAD")
		}

		file, err := c.FormFile("image")
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "image file is required", "INVALID_PAYLOAD")
		}

		// Check Quota
		if err := h.quotaSvc.CheckQuota(project.ID, file.Size); err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "QUOTA_EXCEEDED")
		}

		// Save file locally using storage
		fileInfo, err := storage.SaveFile(file, "media", "image")
		if err != nil {
			if err == storage.ErrFileTooLarge || err == storage.ErrInvalidFileType {
				return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
			}
			return response.JSONError(c, fiber.StatusInternalServerError, "failed to save image", "INTERNAL_SERVER_ERROR")
		}
		url = fileInfo.URL

		// Create FileRecord
		userID := project.UserID
		record := &models.FileRecord{
			URL:          fileInfo.URL,
			FilePath:     fileInfo.FilePath,
			OriginalName: fileInfo.OriginalName,
			ContentType:  fileInfo.ContentType,
			Size:         fileInfo.Size,
			MediaType:    fileInfo.MediaType,
			ProjectID:    &project.ID,
			UploadedByID: &userID,
		}

		if err := h.fileRepo.Create(record); err != nil {
			_ = storage.DeleteFile(fileInfo.URL)
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to save file metadata", "INTERNAL_SERVER_ERROR")
		}
	}

	media, err := h.mediaService.Create(project.ID, project.ThemeID, bucket, url, order)
	if err != nil {
		if err.Error() == "invalid bucket for this theme" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create media mapping", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Media mapping created successfully", media, nil)
}

func (h *MediaHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	mediaIdParam := c.Params("mediaId")
	mediaID, err := strconv.ParseUint(mediaIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid media ID", "INVALID_ID")
	}

	var req MediaUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	media, err := h.mediaService.Update(uint(mediaID), project.ID, req.URL, req.Order)
	if err != nil {
		if err.Error() == "media mapping not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update media mapping", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Media mapping updated successfully", media, nil)
}

func (h *MediaHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	mediaIdParam := c.Params("mediaId")
	mediaID, err := strconv.ParseUint(mediaIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid media ID", "INVALID_ID")
	}

	// Need to get the media URL before deleting to clean up the file
	medias, _ := h.mediaService.GetByProjectID(project.ID)
	var urlToDelete string
	for _, m := range medias {
		if m.ID == uint(mediaID) {
			urlToDelete = m.URL
			break
		}
	}

	if err := h.mediaService.Delete(uint(mediaID), project.ID); err != nil {
		if err.Error() == "media mapping not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete media mapping", "INTERNAL_SERVER_ERROR")
	}

	// Clean up file if it's a local upload
	if urlToDelete != "" {
		if fileRecord, err := h.fileRepo.GetByURL(urlToDelete); err == nil && fileRecord != nil {
			_ = h.fileRepo.Delete(fileRecord.ID) // Soft delete from DB
		}
		_ = storage.DeleteFile(urlToDelete)
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Media mapping deleted successfully", nil, nil)
}
