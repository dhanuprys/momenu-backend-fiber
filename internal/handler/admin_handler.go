package handler

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) ListUsers(c fiber.Ctx) error {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve users", "INTERNAL_ERROR")
	}
	return response.JSONSuccess(c, fiber.StatusOK, "Users retrieved successfully", users, nil)
}

func (h *AdminHandler) LoginAsUser(c fiber.Ctx) error {
	idParam := c.Params("id")
	if idParam == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "User ID is required", "INVALID_REQUEST")
	}

	var targetUser models.User
	if err := database.DB.First(&targetUser, "id = ?", idParam).Error; err != nil {
		return response.JSONError(c, fiber.StatusNotFound, "User not found", "NOT_FOUND")
	}

	token, err := utils.GenerateToken(targetUser.ID, targetUser.Email, targetUser.IsAdmin, targetUser.Verified)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
	}

	refreshToken, err := utils.GenerateRefreshToken(targetUser.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to generate refresh token", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Login successful", fiber.Map{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          targetUser,
	}, nil)
}

func (h *AdminHandler) CreateUser(c fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to hash password", "INTERNAL_ERROR")
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  req.IsAdmin,
		Verified: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create user", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "User created successfully", user, nil)
}

func (h *AdminHandler) UpdateUserStatus(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Verified bool `json:"verified"`
		IsAdmin  bool `json:"is_admin"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(models.User{Verified: req.Verified, IsAdmin: req.IsAdmin}).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update user status", "INTERNAL_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "User status updated", nil, nil)
}

func (h *AdminHandler) DeleteUser(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "User ID is required", "INVALID_REQUEST")
	}

	if err := database.DB.Unscoped().Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete user", "INTERNAL_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "User deleted successfully", nil, nil)
}

func (h *AdminHandler) ListProjects(c fiber.Ctx) error {
	projects := make([]models.Project, 0)
	if err := database.DB.Preload("Theme").Find(&projects).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve projects", "INTERNAL_ERROR")
	}
	return response.JSONSuccess(c, fiber.StatusOK, "Projects retrieved successfully", projects, nil)
}

func (h *AdminHandler) DeleteProject(c fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.Project{}, "id = ?", id).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete project", "INTERNAL_ERROR")
	}
	return response.JSONSuccess[any](c, fiber.StatusOK, "Project deleted successfully", nil, nil)
}

func (h *AdminHandler) CreateMusicCategory(c fiber.Ctx) error {
	var req models.MusicCategory
	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	if err := database.DB.Create(&req).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create music category", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Music category created successfully", req, nil)
}

func (h *AdminHandler) UpdateMusicCategory(c fiber.Ctx) error {
	id := c.Params("id")
	var req models.MusicCategory
	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	var category models.MusicCategory
	if err := database.DB.First(&category, id).Error; err != nil {
		return response.JSONError(c, fiber.StatusNotFound, "Music category not found", "NOT_FOUND")
	}

	category.Name = req.Name
	category.Slug = req.Slug
	category.Description = req.Description
	category.Order = req.Order

	if err := database.DB.Save(&category).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update music category", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Music category updated successfully", category, nil)
}

func (h *AdminHandler) DeleteMusicCategory(c fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.MusicCategory{}, "id = ?", id).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete music category", "INTERNAL_ERROR")
	}
	return response.JSONSuccess[any](c, fiber.StatusOK, "Music category deleted successfully", nil, nil)
}

func (h *AdminHandler) CreateMusic(c fiber.Ctx) error {
	var req models.Music
	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	if err := database.DB.Create(&req).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create music", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Music created successfully", req, nil)
}

func (h *AdminHandler) DeleteMusic(c fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.Music{}, "id = ?", id).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete music", "INTERNAL_ERROR")
	}
	return response.JSONSuccess[any](c, fiber.StatusOK, "Music deleted successfully", nil, nil)
}

func (h *AdminHandler) ListThemes(c fiber.Ctx) error {
	var themes []models.Theme
	if err := database.DB.Find(&themes).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve themes", "INTERNAL_ERROR")
	}

	for i := range themes {
		var count int64
		database.DB.Model(&models.Project{}).Where("theme_id = ?", themes[i].ID).Count(&count)
		themes[i].UsageCount = count
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Themes retrieved successfully", themes, nil)
}

func (h *AdminHandler) UpdateTheme(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Price       *float64 `json:"price"`
		Thumbnail   string   `json:"thumbnail"`
		Description string   `json:"description"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
	}

	var theme models.Theme
	if err := database.DB.First(&theme, "id = ?", id).Error; err != nil {
		return response.JSONError(c, fiber.StatusNotFound, "Theme not found", "NOT_FOUND")
	}

	updates := map[string]interface{}{
		"price":       req.Price,
		"thumbnail":   req.Thumbnail,
		"description": req.Description,
	}

	if err := database.DB.Model(&theme).Updates(updates).Error; err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update theme", "INTERNAL_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Theme updated successfully", theme, nil)
}
