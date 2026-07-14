package handler

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type ThemeHandler struct {
	themeService service.ThemeService
}

func NewThemeHandler(themeService service.ThemeService) *ThemeHandler {
	return &ThemeHandler{themeService: themeService}
}

func (h *ThemeHandler) List(c fiber.Ctx) error {
	eventType := c.Query("event_type")
	themes, err := h.themeService.GetAllThemes(eventType)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve themes", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Themes retrieved successfully", themes, nil)
}

func (h *ThemeHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	theme, err := h.themeService.GetThemeByID(id)
	if err != nil {
		if err.Error() == "theme not found" {
			return response.JSONError(c, fiber.StatusNotFound, "Theme not found", "THEME_NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve theme", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Theme retrieved successfully", theme, nil)
}
