package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type LiveStreamHandler struct {
	liveStreamService service.LiveStreamService
}

func NewLiveStreamHandler(liveStreamService service.LiveStreamService) *LiveStreamHandler {
	return &LiveStreamHandler{liveStreamService: liveStreamService}
}

type LiveStreamRequest struct {
	Platform string `json:"platform" validate:"required"`
	URL      string `json:"url" validate:"required,url"`
}

func (h *LiveStreamHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	streams, err := h.liveStreamService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve live streams", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Live streams retrieved successfully", streams, nil)
}

func (h *LiveStreamHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req LiveStreamRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	stream, err := h.liveStreamService.Create(project.ID, req.Platform, req.URL)
	if err != nil {
		if err.Error() == "invalid live stream platform" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create live stream", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Live stream created successfully", stream, nil)
}

func (h *LiveStreamHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	streamIdParam := c.Params("streamId")
	streamID, err := strconv.ParseUint(streamIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid live stream ID", "INVALID_ID")
	}

	var req LiveStreamRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	stream, err := h.liveStreamService.Update(uint(streamID), project.ID, req.Platform, req.URL)
	if err != nil {
		if err.Error() == "live stream not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		if err.Error() == "invalid live stream platform" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update live stream", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Live stream updated successfully", stream, nil)
}

func (h *LiveStreamHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	streamIdParam := c.Params("streamId")
	streamID, err := strconv.ParseUint(streamIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid live stream ID", "INVALID_ID")
	}

	if err := h.liveStreamService.Delete(uint(streamID), project.ID); err != nil {
		if err.Error() == "live stream not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete live stream", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Live stream deleted successfully", nil, nil)
}
