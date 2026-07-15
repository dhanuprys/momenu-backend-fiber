package handler

import (
	"encoding/json"
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

type CreateProjectRequest struct {
	Title   string          `json:"title" validate:"required,min=3"`
	ThemeID string          `json:"theme_id" validate:"required"`
	MusicID *uint           `json:"music_id"`
	Payload json.RawMessage `json:"payload"`
}

type UpdateProjectRequest struct {
	Title            string          `json:"title"`
	Slug             string          `json:"slug" validate:"required,min=10,max=50"`
	Payload          json.RawMessage `json:"payload"`
	SharingThumbnail string          `json:"sharing_thumbnail"`
	MusicID          *uint           `json:"music_id"`
}

type UpdateProjectStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

func (h *ProjectHandler) Create(c fiber.Ctx) error {
	var req CreateProjectRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	userID := c.Locals("user_id").(uint)

	project, validationErrs, err := h.projectService.CreateProject(userID, req.Title, req.ThemeID, req.MusicID, req.Payload)
	if err != nil {
		if err.Error() == "validation failed" {
			return response.JSONValidationError(c, validationErrs)
		}
		if err.Error() == "theme not found" || err.Error() == "unsupported event type" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create project", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Project created successfully", project, nil)
}

// GetVersion returns the current update count of the project
func (h *ProjectHandler) GetVersion(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	return response.JSONSuccess(c, fiber.StatusOK, "Project version retrieved", fiber.Map{
		"update_count": project.UpdateCount,
	}, nil)
}

func (h *ProjectHandler) List(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	pageStr := c.Query("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	status := c.Query("status", "")

	projects, total, err := h.projectService.GetProjectsByUserID(userID, page, limit, status)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve projects", "INTERNAL_SERVER_ERROR")
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

	return response.JSONSuccess(c, fiber.StatusOK, "Projects retrieved successfully", projects, pagination)
}

func (h *ProjectHandler) Get(c fiber.Ctx) error {
	// Project is injected by owner middleware
	project := c.Locals("project").(*models.Project)
	return response.JSONSuccess(c, fiber.StatusOK, "Project retrieved successfully", project, nil)
}

func (h *ProjectHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req UpdateProjectRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	updatedProject, validationErrs, err := h.projectService.UpdateProject(project.ID, req.Title, req.Slug, req.Payload, req.SharingThumbnail, req.MusicID)
	if err != nil {
		if err.Error() == "validation failed" {
			return response.JSONValidationError(c, validationErrs)
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update project", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Project updated successfully", updatedProject, nil)
}

func (h *ProjectHandler) UpdateStatus(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	userID := c.Locals("user_id").(uint)

	var req UpdateProjectStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	// Simple validation for enum
	status := models.ProjectStatus(req.Status)
	if status != models.ProjectStatusDraft && status != models.ProjectStatusPublished && status != models.ProjectStatusArchived {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid status", "INVALID_STATUS")
	}

	updatedProject, err := h.projectService.UpdateStatus(project.ID, status, userID)
	if err != nil {
		if err.Error() == "only verified users can publish projects" {
			return response.JSONError(c, fiber.StatusForbidden, err.Error(), "VERIFICATION_REQUIRED")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update project status", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Project status updated successfully", updatedProject, nil)
}

func (h *ProjectHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	if err := h.projectService.DeleteProject(project.ID); err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete project", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Project deleted successfully", nil, nil)
}

func (h *ProjectHandler) GetFeatureToggle(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	toggle, err := h.projectService.GetFeatureToggle(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve feature toggle", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Feature toggle retrieved successfully", toggle, nil)
}

type UpdateFeatureToggleRequest struct {
	ShowRSVP               bool `json:"show_rsvp"`
	ShowWishes             bool `json:"show_wishes"`
	ShowGallery            bool `json:"show_gallery"`
	ShowGifts              bool `json:"show_gifts"`
	ShowLiveStream         bool `json:"show_live_stream"`
	RequireRegisteredGuest bool `json:"require_registered_guest"`
}

func (h *ProjectHandler) UpdateFeatureToggle(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req UpdateFeatureToggleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	toggle, err := h.projectService.GetFeatureToggle(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve feature toggle", "INTERNAL_SERVER_ERROR")
	}

	toggle.ShowRSVP = req.ShowRSVP
	toggle.ShowWishes = req.ShowWishes
	toggle.ShowGallery = req.ShowGallery
	toggle.ShowGifts = req.ShowGifts
	toggle.ShowLiveStream = req.ShowLiveStream
	toggle.RequireRegisteredGuest = req.RequireRegisteredGuest

	if err := h.projectService.UpdateFeatureToggle(toggle); err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update feature toggle", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Feature toggle updated successfully", toggle, nil)
}
