package handler

import (
	"fmt"
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"
)

type RSVPHandler struct {
	rsvpService    service.RSVPService
	projectService service.ProjectService
}

func NewRSVPHandler(rsvpService service.RSVPService, projectService service.ProjectService) *RSVPHandler {
	return &RSVPHandler{
		rsvpService:    rsvpService,
		projectService: projectService,
	}
}

type RSVPRequest struct {
	Name       string `json:"name" validate:"required"`
	Attending  bool   `json:"attending"`
	GuestCount int    `json:"guest_count"`
}

// List handles owner fetching RSVPs
func (h *RSVPHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

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

	rsvps, total, err := h.rsvpService.GetByProjectID(project.ID, page, limit)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve RSVPs", "INTERNAL_SERVER_ERROR")
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

	return response.JSONSuccess(c, fiber.StatusOK, "RSVPs retrieved successfully", rsvps, pagination)
}

// Stats handles owner fetching RSVP stats
func (h *RSVPHandler) Stats(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	stats, err := h.rsvpService.GetStatsByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve RSVP stats", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "RSVP stats retrieved successfully", stats, nil)
}

// PublicUpsert handles guests submitting RSVPs
func (h *RSVPHandler) PublicUpsert(c fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.projectService.GetProjectBySlug(slug)
	if err != nil || project == nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
	}

	// Only allow RSVP if published
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Project is not active", "FORBIDDEN")
	}

	var req RSVPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	rsvp, err := h.rsvpService.Upsert(project.ID, req.Name, req.Attending, req.GuestCount)
	if err != nil {
		if err.Error() == "RSVP feature is disabled for this project" {
			return response.JSONError(c, fiber.StatusForbidden, err.Error(), "FEATURE_DISABLED")
		}
		if err.Error() == "guest count must be greater than 0 if attending" {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "BAD_REQUEST")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to submit RSVP", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "RSVP submitted successfully", rsvp, nil)
}

type OwnerRSVPRequest struct {
	Name           string  `json:"name" validate:"required"`
	SpecialMessage string  `json:"special_message"`
	Whatsapp       *string `json:"whatsapp"`
}

// OwnerUpsert handles the project owner manually adding/updating a guest
func (h *RSVPHandler) OwnerUpsert(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req OwnerRSVPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	rsvp, err := h.rsvpService.OwnerUpsert(project.ID, req.Name, req.SpecialMessage, req.Whatsapp)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to upsert guest", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Guest added/updated successfully", rsvp, nil)
}

// PublicGetGuest handles retrieving a specific guest's RSVP record (for silent sync)
func (h *RSVPHandler) PublicGetGuest(c fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.projectService.GetProjectBySlug(slug)
	if err != nil || project == nil {
		return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
	}

	// Only allow if published
	if project.Status != models.ProjectStatusPublished {
		return response.JSONError(c, fiber.StatusForbidden, "Project is not active", "FORBIDDEN")
	}

	name := c.Query("name")
	if name == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Guest name is required", "BAD_REQUEST")
	}

	rsvp, err := h.rsvpService.GetGuestByName(project.ID, name)
	if err != nil {
		return response.JSONError(c, fiber.StatusNotFound, "Guest not found", "NOT_FOUND")
	}

	// Mark as opened
	h.rsvpService.MarkAsOpened(project.ID, name)

	return response.JSONSuccess(c, fiber.StatusOK, "Guest retrieved successfully", rsvp, nil)
}

// Delete handles the project owner removing a guest
func (h *RSVPHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)
	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid guest ID", "BAD_REQUEST")
	}

	if err := h.rsvpService.Delete(project.ID, uint(id)); err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete guest", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Guest deleted successfully", nil, nil)
}

func (h *RSVPHandler) ExportXLSX(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	rsvps, err := h.rsvpService.GetAllByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve RSVPs", "INTERNAL_SERVER_ERROR")
	}

	f := excelize.NewFile()
	sheetName := "Guest List"
	f.SetSheetName("Sheet1", sheetName)

	// Headers
	f.SetCellValue(sheetName, "A1", "Nama Tamu")
	f.SetCellValue(sheetName, "B1", "WhatsApp")
	f.SetCellValue(sheetName, "C1", "Hadir")
	f.SetCellValue(sheetName, "D1", "Jumlah")
	f.SetCellValue(sheetName, "E1", "Pesan")
	f.SetCellValue(sheetName, "F1", "Sudah Merespon")
	f.SetCellValue(sheetName, "G1", "Sudah Dibuka")

	for i, rsvp := range rsvps {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rsvp.Name)
		if rsvp.Whatsapp != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), *rsvp.Whatsapp)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), rsvp.Attending)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rsvp.GuestCount)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), rsvp.SpecialMessage)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), rsvp.IsResponded)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), rsvp.HasOpened)
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=guests_%s.xlsx", project.ID.String()))

	if err := f.Write(c.Response().BodyWriter()); err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to generate Excel", "INTERNAL_SERVER_ERROR")
	}
	return nil
}

func (h *RSVPHandler) ImportXLSX(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	file, err := c.FormFile("file")
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "File is required", "BAD_REQUEST")
	}

	f, err := file.Open()
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to open file", "INTERNAL_SERVER_ERROR")
	}
	defer f.Close()

	xlsx, err := excelize.OpenReader(f)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid Excel file", "BAD_REQUEST")
	}

	sheetName := xlsx.GetSheetName(0) // Get first sheet
	rows, err := xlsx.GetRows(sheetName)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to read rows", "INTERNAL_SERVER_ERROR")
	}

	successCount := 0
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 1 || row[0] == "" {
			continue // Skip empty names
		}

		name := row[0]
		var whatsapp *string
		if len(row) >= 2 && row[1] != "" {
			wa := row[1]
			whatsapp = &wa
		}

		_, err := h.rsvpService.OwnerUpsert(project.ID, name, "", whatsapp)
		if err == nil {
			successCount++
		}
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Guests imported successfully", map[string]int{"imported": successCount}, nil)
}
