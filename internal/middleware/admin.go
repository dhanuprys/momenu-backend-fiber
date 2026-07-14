package middleware

import (
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

// AdminRequired validates if the authenticated user has admin privileges.
// It must be used after AuthRequired middleware.
func AdminRequired(c fiber.Ctx) error {
	isAdmin, ok := c.Locals("is_admin").(bool)
	if !ok || !isAdmin {
		return response.JSONError(c, fiber.StatusForbidden, "Admin access required", "FORBIDDEN")
	}

	return c.Next()
}
