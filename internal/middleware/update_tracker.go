package middleware

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/worker"
	"github.com/gofiber/fiber/v3"
)

// TrackProjectUpdate tracks successful modifying requests and increments the project update counter
func TrackProjectUpdate(c fiber.Ctx) error {
	// Execute the handler first
	err := c.Next()

	// Only process successful requests
	if err == nil {
		method := c.Method()
		statusCode := c.Response().StatusCode()

		// If it's a mutating request and it succeeded
		if (method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch || method == fiber.MethodDelete) && (statusCode >= 200 && statusCode < 400) {
			
			// Extract project from locals (injected by ProjectOwner middleware)
			if proj, ok := c.Locals("project").(*models.Project); ok {
				worker.IncrementUpdateCount(proj.ID)
			}
		}
	}

	return err
}
