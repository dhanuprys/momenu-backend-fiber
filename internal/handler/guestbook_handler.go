package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type GuestbookHandler struct {
	guestbookService service.GuestbookService
	projectService   service.ProjectService
	rsvpService      service.RSVPService
}

func NewGuestbookHandler(guestbookService service.GuestbookService, projectService service.ProjectService, rsvpService service.RSVPService) *GuestbookHandler {
	return &GuestbookHandler{
		guestbookService: guestbookService,
		projectService:   projectService,
		rsvpService:      rsvpService,
	}
}

type GuestbookRequest struct {
	Name    string `json:"name" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// PublicList handles fetching guestbook entries for a public project
func (h *GuestbookHandler) PublicList(c fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.projectService.GetProjectBySlug(slug)
	if err != nil || project == nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
	}

	// Only allow if published
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Project is not active", "FORBIDDEN")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	entries, total, err := h.guestbookService.GetByProjectID(project.ID, page, limit)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve guestbook entries", "INTERNAL_SERVER_ERROR")
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	pagination := &response.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Guestbook entries retrieved successfully", entries, pagination)
}

// PublicCreate handles guests submitting a wish
func (h *GuestbookHandler) PublicCreate(c fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.projectService.GetProjectBySlug(slug)
	if err != nil || project == nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
	}

	// Only allow if published
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Project is not active", "FORBIDDEN")
	}

	var req GuestbookRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	// Check if registered guest is required
	if project.FeatureToggle.RequireRegisteredGuest {
		_, err := h.rsvpService.GetGuestByName(project.ID, req.Name)
		if err != nil {
			return response.JSONError(c, fiber.StatusForbidden, "Hanya tamu terdaftar yang dapat memberikan ucapan", "FORBIDDEN")
		}
	}

	entry, err := h.guestbookService.Create(project.ID, req.Name, req.Message)
	if err != nil {
		if err.Error() == "guestbook feature is disabled for this project" {
			return response.JSONError(c, fiber.StatusForbidden, err.Error(), "FEATURE_DISABLED")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to submit guestbook entry", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Guestbook entry submitted successfully", entry, nil)
}

// List handles owner fetching guestbook entries
func (h *GuestbookHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	entries, total, err := h.guestbookService.GetByProjectID(project.ID, page, limit)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve guestbook entries", "INTERNAL_SERVER_ERROR")
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	pagination := &response.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Guestbook entries retrieved successfully", entries, pagination)
}

// Delete handles owner deleting a specific entry
func (h *GuestbookHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	entryIdParam := c.Params("entryId")
	entryID, err := strconv.ParseUint(entryIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid entry ID", "INVALID_ID")
	}

	if err := h.guestbookService.Delete(uint(entryID), project.ID); err != nil {
		if err.Error() == "guestbook entry not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete guestbook entry", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Guestbook entry deleted successfully", nil, nil)
}
