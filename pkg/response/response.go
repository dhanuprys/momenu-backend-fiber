package response

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Pagination defines the structure for pagination metadata
type Pagination struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
}

// Meta defines standard metadata included in every response
type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  string      `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Success structure for standardizing API success responses
type Success[T any] struct {
	Success bool   `json:"success" default:"true"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    Meta   `json:"meta"`
}

// ValidationError defines the structure for validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error structure for standardizing API error responses
type Error struct {
	Success bool              `json:"success" default:"false"`
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Meta    Meta              `json:"meta"`
}

func buildMeta(c fiber.Ctx, pagination *Pagination) Meta {
	reqID := c.Get("X-Request-Id")
	if reqID == "" {
		if id, ok := c.Locals("requestid").(string); ok {
			reqID = id
		} else {
			reqID = uuid.NewString()
		}
	}
	return Meta{
		RequestID:  reqID,
		Timestamp:  time.Now().Format(time.RFC3339),
		Pagination: pagination,
	}
}

// JSONSuccess is a helper function to return a success response
func JSONSuccess[T any](c fiber.Ctx, status int, message string, data T, pagination *Pagination) error {
	return c.Status(status).JSON(Success[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    buildMeta(c, pagination),
	})
}

// JSONError is a helper function to return an error response
func JSONError(c fiber.Ctx, status int, message string, code string) error {
	return c.Status(status).JSON(Error{
		Success: false,
		Message: message,
		Code:    code,
		Meta:    buildMeta(c, nil),
	})
}

// JSONValidationError is a helper function to return validation errors
func JSONValidationError(c fiber.Ctx, errors []ValidationError) error {
	return c.Status(fiber.StatusBadRequest).JSON(Error{
		Success: false,
		Message: "Validation failed",
		Code:    "VALIDATION_ERROR",
		Errors:  errors,
		Meta:    buildMeta(c, nil),
	})
}
