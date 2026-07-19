package handler

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm/clause"
)

type TextOverrideHandler struct{}

func NewTextOverrideHandler() *TextOverrideHandler {
	return &TextOverrideHandler{}
}

// Request structure for bulk upsert
type UpsertTextOverrideRequest struct {
	Overrides []TextOverrideItem `json:"overrides"`
}

type TextOverrideItem struct {
	SlotKey    string `json:"slot_key" validate:"required,max=100"`
	Value      string `json:"value" validate:"max=5000"`
	Bold       bool   `json:"bold"`
	Italic     bool   `json:"italic"`
	Underline  bool   `json:"underline"`
	TextAlign  string `json:"text_align" validate:"max=20"`
	FontFamily string `json:"font_family" validate:"max=100"`
}

func (h *TextOverrideHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var overrides []models.TextOverride
	if err := database.DB.Where("project_id = ?", project.ID).Find(&overrides).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve text overrides", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Text overrides retrieved successfully", overrides, nil)
}

func (h *TextOverrideHandler) GetBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")

	var project models.Project
	if err := database.DB.Where("slug = ?", slug).First(&project).Error; err != nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
	}

	var overrides []models.TextOverride
	if err := database.DB.Where("project_id = ?", project.ID).Find(&overrides).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve text overrides", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Text overrides retrieved successfully", overrides, nil)
}

func (h *TextOverrideHandler) Upsert(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req UpsertTextOverrideRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to start transaction", "INTERNAL_ERROR")
	}
	defer tx.Rollback()

	if len(req.Overrides) == 0 {
		if err := tx.Where("project_id = ?", project.ID).Delete(&models.TextOverride{}).Error; err != nil {
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to reset overrides", "INTERNAL_ERROR")
		}
		tx.Commit()
		return response.JSONSuccess[any](c, fiber.StatusOK, "All overrides reset to default", nil, nil)
	}

	// Prepare data for upsert
	var dbOverrides []models.TextOverride
	var incomingKeys []string
	for _, item := range req.Overrides {
		incomingKeys = append(incomingKeys, item.SlotKey)
		dbOverrides = append(dbOverrides, models.TextOverride{
			ProjectID:  project.ID,
			SlotKey:    item.SlotKey,
			Value:      item.Value,
			Bold:       item.Bold,
			Italic:     item.Italic,
			Underline:  item.Underline,
			TextAlign:  item.TextAlign,
			FontFamily: item.FontFamily,
		})
	}

	// Delete existing overrides not in the incoming payload (meaning they were reset to default)
	if err := tx.Where("project_id = ? AND slot_key NOT IN ?", project.ID, incomingKeys).Delete(&models.TextOverride{}).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to process reset overrides", "INTERNAL_ERROR")
	}

	// Use GORM Clauses for Upsert (ON CONFLICT)
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "slot_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "bold", "italic", "underline", "text_align", "font_family", "updated_at"}),
	}).Create(&dbOverrides).Error

	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to save text overrides", "INTERNAL_ERROR")
	}

	if err := tx.Commit().Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to commit changes", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Text overrides saved successfully", dbOverrides, nil)
}

func (h *TextOverrideHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	slotKey := c.Params("slotKey")

	if slotKey == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Slot key is required", "BAD_REQUEST")
	}

	if err := database.DB.Where("project_id = ? AND slot_key = ?", project.ID, slotKey).Delete(&models.TextOverride{}).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete text override", "INTERNAL_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Text override deleted successfully", nil, nil)
}
