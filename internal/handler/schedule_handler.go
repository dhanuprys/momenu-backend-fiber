package handler

import (
	"strconv"
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type ScheduleHandler struct {
	scheduleService service.ScheduleService
}

func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

type ScheduleRequest struct {
	Title     string    `json:"title" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   *time.Time `json:"end_time"`
	Timezone  string    `json:"timezone" validate:"required"`
	Location  string    `json:"location"`
	MapURL    string    `json:"map_url"`
}

func (h *ScheduleHandler) List(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	schedules, err := h.scheduleService.GetByProjectID(project.ID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve schedules", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Schedules retrieved successfully", schedules, nil)
}

func (h *ScheduleHandler) Create(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	var req ScheduleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	schedule, err := h.scheduleService.Create(project.ID, req.Title, req.StartTime, req.EndTime, req.Timezone, req.Location, req.MapURL)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to create schedule", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "Schedule created successfully", schedule, nil)
}

func (h *ScheduleHandler) Update(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	scheduleIdParam := c.Params("scheduleId")
	scheduleID, err := strconv.ParseUint(scheduleIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid schedule ID", "INVALID_ID")
	}

	var req ScheduleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	schedule, err := h.scheduleService.Update(uint(scheduleID), project.ID, req.Title, req.StartTime, req.EndTime, req.Timezone, req.Location, req.MapURL)
	if err != nil {
		if err.Error() == "schedule not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to update schedule", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Schedule updated successfully", schedule, nil)
}

func (h *ScheduleHandler) Delete(c fiber.Ctx) error {
	project := c.Locals("project").(*models.Project)

	scheduleIdParam := c.Params("scheduleId")
	scheduleID, err := strconv.ParseUint(scheduleIdParam, 10, 32)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid schedule ID", "INVALID_ID")
	}

	if err := h.scheduleService.Delete(uint(scheduleID), project.ID); err != nil {
		if err.Error() == "schedule not found or does not belong to project" {
			return response.JSONError(c, fiber.StatusNotFound, err.Error(), "NOT_FOUND")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to delete schedule", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess[any](c, fiber.StatusOK, "Schedule deleted successfully", nil, nil)
}
