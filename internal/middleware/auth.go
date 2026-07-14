package middleware

import (
	"strings"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/database"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

// AuthRequired validates the JWT token and injects user claims into the context.
func AuthRequired(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return response.JSONError(c, fiber.StatusUnauthorized, "Missing authorization header", "UNAUTHORIZED")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return response.JSONError(c, fiber.StatusUnauthorized, "Invalid authorization format", "UNAUTHORIZED")
	}

	tokenString := parts[1]
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return response.JSONError(c, fiber.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
	}

	// Verify that the user still exists in the database
	var userCount int64
	if err := database.DB.Model(&models.User{}).Where("id = ?", claims.UserID).Count(&userCount).Error; err != nil || userCount == 0 {
		return response.JSONError(c, fiber.StatusUnauthorized, "User no longer exists, please log in again", "UNAUTHORIZED")
	}

	// Inject claims into context
	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)
	c.Locals("is_admin", claims.IsAdmin)
	c.Locals("verified", claims.Verified)

	return c.Next()
}
