package handler

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"

	"github.com/gofiber/fiber/v3"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

type StyleOverrideHandler struct{}

func NewStyleOverrideHandler() *StyleOverrideHandler {
	return &StyleOverrideHandler{}
}

// Request structure for bulk upsert
type UpsertStyleOverrideRequest struct {
	Overrides []StyleOverrideItem `json:"overrides"`
}

type StyleOverrideItem struct {
	SlotKey    string         `json:"slot_key" validate:"required"`
	Properties datatypes.JSON `json:"properties"`
}

func (h *StyleOverrideHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var overrides []models.StyleOverride
	if err := database.DB.Where("project_id = ?", project.ID).Find(&overrides).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve style overrides", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Style overrides retrieved successfully", overrides, nil)
}

func (h *StyleOverrideHandler) Upsert(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req UpsertStyleOverrideRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if len(req.Overrides) == 0 {
		return response.JSONSuccess[any](c, fiber.StatusOK, "No style overrides to update", nil, nil)
	}

	// Prepare data for upsert
	var dbOverrides []models.StyleOverride
	for _, item := range req.Overrides {
		dbOverrides = append(dbOverrides, models.StyleOverride{
			ProjectID:  project.ID,
			SlotKey:    item.SlotKey,
			Properties: item.Properties,
		})
	}

	// Use GORM Clauses for Upsert (ON CONFLICT)
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "slot_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"properties", "updated_at"}),
	}).Create(&dbOverrides).Error

	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to save style overrides", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Style overrides saved successfully", dbOverrides, nil)
}

func (h *StyleOverrideHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	slotKey := c.Params("slotKey")

	if slotKey == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Slot key is required", "BAD_REQUEST")
	}

	if err := database.DB.Where("project_id = ? AND slot_key = ?", project.ID, slotKey).Delete(&models.StyleOverride{}).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete style override", "INTERNAL_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Style override deleted successfully", nil, nil)
}
