package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type DressCodeHandler struct {
	dressCodeService service.DressCodeService
}

func NewDressCodeHandler(dressCodeService service.DressCodeService) *DressCodeHandler {
	return &DressCodeHandler{dressCodeService: dressCodeService}
}

type DressCodeRequest struct {
	Label  string   `json:"label" validate:"required"`
	Colors []string `json:"colors" validate:"required"`
}

func (h *DressCodeHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	dressCodes, err := h.dressCodeService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve dress codes", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Dress codes retrieved successfully", dressCodes, nil)
}

func (h *DressCodeHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req DressCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	dressCode, err := h.dressCodeService.Create(project.ID, req.Label, req.Colors)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create dress code", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Dress code created successfully", dressCode, nil)
}

func (h *DressCodeHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	
	dressCodeIdParam := c.Params("dressCodeId")
	dressCodeID, err := strconv.ParseUint(dressCodeIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid dress code ID", "INVALID_ID")
	}

	var req DressCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	dressCode, err := h.dressCodeService.Update(uint(dressCodeID), project.ID, req.Label, req.Colors)
	if err != nil {
		if err.Error() == "dress code not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update dress code", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Dress code updated successfully", dressCode, nil)
}

func (h *DressCodeHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	
	dressCodeIdParam := c.Params("dressCodeId")
	dressCodeID, err := strconv.ParseUint(dressCodeIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid dress code ID", "INVALID_ID")
	}

	if err := h.dressCodeService.Delete(uint(dressCodeID), project.ID); err != nil {
		if err.Error() == "dress code not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete dress code", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Dress code deleted successfully", nil, nil)
}
