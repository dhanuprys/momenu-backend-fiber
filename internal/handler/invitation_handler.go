package handler

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type InvitationHandler struct {
	invitationService service.InvitationService
	rsvpService       service.RSVPService
}

func NewInvitationHandler(invitationService service.InvitationService, rsvpService service.RSVPService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		rsvpService:       rsvpService,
	}
}

func (h *InvitationHandler) GetInvitation(c fiber.Ctx) error {
	slug := c.Params("slug")

	project, err := h.invitationService.GetInvitationBySlug(slug)
	if err != nil {
		if err.Error() == "invitation not found" {
			return response.JSONError(c, fiber.StatusNotFound, "Invitation not found", "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve invitation", "INTERNAL_SERVER_ERROR")
	}

	// Strictly require published status for public access
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Undangan belum aktif atau sedang tidak tersedia", "FORBIDDEN")
	}

	// Check if registered guest is required
	if project.FeatureToggle.RequireRegisteredGuest {
		guestName := c.Query("name")
		if guestName == "" {
			return response.JSONError(c, fiber.StatusForbidden, "Nama tamu diperlukan untuk mengakses undangan ini", "FORBIDDEN")
		}
		_, err := h.rsvpService.GetGuestByName(project.ID, guestName)
		if err != nil {
			return response.JSONError(c, fiber.StatusForbidden, "Maaf, nama Anda tidak terdaftar dalam daftar tamu", "FORBIDDEN")
		}
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Invitation retrieved successfully", project, nil)
}

func (h *InvitationHandler) GetOGMetadata(c fiber.Ctx) error {
	slug := c.Params("slug")

	project, err := h.invitationService.GetOGMetadata(slug)
	if err != nil {
		if err.Error() == "invitation not found" {
			return response.JSONError(c, fiber.StatusNotFound, "Invitation not found", "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve OG metadata", "INTERNAL_SERVER_ERROR")
	}

	// Strictly require published status for public access
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Undangan belum aktif atau sedang tidak tersedia", "FORBIDDEN")
	}

	// For OG Metadata, we deliberately SKIP the RequireRegisteredGuest check 
	// so that social media crawlers can still see the title/thumbnail.

	return response.JSONSuccess(c, fiber.StatusOK, "OG Metadata retrieved successfully", project, nil)
}
