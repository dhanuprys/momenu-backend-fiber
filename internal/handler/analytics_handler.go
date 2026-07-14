package handler

import (
	"strings"

	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

type RecordVisitRequest struct {
	ProjectID string `json:"project_id" validate:"required"`
	GuestName string `json:"guest_name"`
	Source    string `json:"source"`
}

func parseDeviceType(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mobi") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
			return "Tablet"
		}
		return "Mobile"
	}
	return "Desktop"
}

func (h *AnalyticsHandler) RecordVisit(c fiber.Ctx) error {
	var req RecordVisitRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	userAgent := c.Get("User-Agent")
	deviceType := parseDeviceType(userAgent)
	ipAddress := c.IP()

	if err := h.analyticsService.RecordVisit(req.ProjectID, req.GuestName, req.Source, userAgent, deviceType, ipAddress); err != nil {
		// We can fail silently or return error, but it's better to return success for tracking pixels
		// so we don't break the client, but for API let's return error
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to record visit", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Visit recorded successfully", nil, nil)
}

func (h *AnalyticsHandler) GetProjectAnalytics(c fiber.Ctx) error {
	projectIDStr := c.Params("projectId")
	userID := c.Locals("user_id").(uint)

	analyticsData, err := h.analyticsService.GetProjectAnalytics(projectIDStr, userID)
	if err != nil {
		if err.Error() == "unauthorized access to project analytics" {
			return response.JSONError(c, fiber.StatusForbidden, err.Error(), "FORBIDDEN")
		}
		if err.Error() == "project not found" || err.Error() == "invalid project id" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve analytics", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Analytics retrieved successfully", analyticsData, nil)
}
