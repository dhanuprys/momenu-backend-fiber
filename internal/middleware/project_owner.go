package middleware

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ProjectOwner verifies that the authenticated user owns the requested project.
// Must be used AFTER AuthRequired middleware.
func ProjectOwner(projectRepo repository.ProjectRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return response.JSONError(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		}

		projectIdParam := c.Params("projectId")
		projectID, err := uuid.Parse(projectIdParam)
		if err != nil {
			return response.JSONError(c, fiber.StatusBadRequest, "Invalid project ID format", "INVALID_PROJECT_ID")
		}

		project, err := projectRepo.GetProjectByID(projectID)
		if err != nil {
			return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve project", "INTERNAL_SERVER_ERROR")
		}
		if project == nil {
			return response.JSONError(c, fiber.StatusNotFound, "Project not found", "PROJECT_NOT_FOUND")
		}

		if project.UserID != userID {
			return response.JSONError(c, fiber.StatusForbidden, "You do not have permission to access this project", "FORBIDDEN")
		}

		// Inject the project into locals to avoid re-querying in the handler
		c.Locals("project", project)

		return c.Next()
	}
}
