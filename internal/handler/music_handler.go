package handler

import (
	"strconv"

	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type MusicHandler struct {
	musicService service.MusicService
}

func NewMusicHandler(musicService service.MusicService) *MusicHandler {
	return &MusicHandler{musicService: musicService}
}

func (h *MusicHandler) ListCategories(c fiber.Ctx) error {
	categories, err := h.musicService.ListCategories()
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve music categories", "INTERNAL_SERVER_ERROR")
	}
	return response.JSONSuccess(c, fiber.StatusOK, "Music categories retrieved successfully", categories, nil)
}

func (h *MusicHandler) ListMusics(c fiber.Ctx) error {
	categoryIDStr := c.Query("category_id")
	
	if categoryIDStr != "" {
		categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Invalid category_id", "INVALID_PARAMETER")
		}
		
		musics, err := h.musicService.ListMusicsByCategory(uint(categoryID))
		if err != nil {
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve musics", "INTERNAL_SERVER_ERROR")
		}
		return response.JSONSuccess(c, fiber.StatusOK, "Musics retrieved successfully", musics, nil)
	}

	musics, err := h.musicService.ListAllMusics()
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve musics", "INTERNAL_SERVER_ERROR")
	}
	return response.JSONSuccess(c, fiber.StatusOK, "Musics retrieved successfully", musics, nil)
}
