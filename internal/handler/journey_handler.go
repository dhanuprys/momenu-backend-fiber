package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type JourneyHandler struct {
	journeyService service.JourneyService
}

func NewJourneyHandler(journeyService service.JourneyService) *JourneyHandler {
	return &JourneyHandler{journeyService: journeyService}
}

type JourneyRequest struct {
	Title   string `json:"title" validate:"required"`
	Date    string `json:"date" validate:"required"`
	Content string `json:"content" validate:"required"`
	Order   int    `json:"order"`
}

func (h *JourneyHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	journeys, err := h.journeyService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve journeys", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Journeys retrieved successfully", journeys, nil)
}

func (h *JourneyHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req JourneyRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	// Validate (assuming standard validate is setup or doing manual here)
	if req.Title == "" || req.Date == "" || req.Content == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Title, Date, and Content are required", "VALIDATION_FAILED")
	}

	journey, err := h.journeyService.Create(project.ID, req.Title, req.Date, req.Content, req.Order)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create journey", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Journey created successfully", journey, nil)
}

func (h *JourneyHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	idStr := c.Params("journeyId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid journey ID", "INVALID_ID")
	}

	var req JourneyRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if req.Title == "" || req.Date == "" || req.Content == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Title, Date, and Content are required", "VALIDATION_FAILED")
	}

	journey, err := h.journeyService.Update(uint(id), project.ID, req.Title, req.Date, req.Content, req.Order)
	if err != nil {
		if err.Error() == "journey not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update journey", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Journey updated successfully", journey, nil)
}

func (h *JourneyHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	idStr := c.Params("journeyId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid journey ID", "INVALID_ID")
	}

	if err := h.journeyService.Delete(uint(id), project.ID); err != nil {
		if err.Error() == "journey not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete journey", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Journey deleted successfully", nil, nil)
}
