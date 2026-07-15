package handler

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ShareHandler struct {
	shareService service.ShareService
}

func NewShareHandler(shareService service.ShareService) *ShareHandler {
	return &ShareHandler{shareService: shareService}
}

type CreateSessionRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
}

func (h *ShareHandler) CreateSession(c fiber.Ctx) error {
	projectIDStr := c.Params("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid project ID", "INVALID_ID")
	}

	var req CreateSessionRequest
	if err := c.Bind().Body(&req); err != nil {
		// Log the error but continue with null expiresAt if body is just not provided
	}

	session, err := h.shareService.CreateSession(projectID, req.ExpiresAt)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, err.Error(), "CREATE_FAILED")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Share session created", session, nil)
}

func (h *ShareHandler) ListSessions(c fiber.Ctx) error {
	projectIDStr := c.Params("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid project ID", "INVALID_ID")
	}

	sessions, err := h.shareService.ListSessions(projectID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to list sessions", "FETCH_FAILED")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Sessions retrieved", sessions, nil)
}

func (h *ShareHandler) RevokeSession(c fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	if err := h.shareService.RevokeSession(sessionID); err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to revoke session", "REVOKE_FAILED")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Session revoked successfully", nil, nil)
}

func (h *ShareHandler) GetSharedData(c fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	project, analytics, err := h.shareService.GetSharedData(sessionID)
	if err != nil {
		if err.Error() == "PROJECT_IS_DRAFT" {
			return response.JSONError(c, fiber.StatusForbidden, "Project is still in draft mode", "PROJECT_IS_DRAFT")
		}
		// Do not leak existence if expired or revoked
		return response.JSONError(c, fiber.StatusNotFound, "Share session not found or invalid", "INVALID_SESSION")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Shared data retrieved", fiber.Map{
		"project":   project,
		"analytics": analytics,
	}, nil)
}
