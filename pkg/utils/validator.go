package utils

import (
	"fmt"

	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct validates a struct based on its "validate" tags and formats
// the output into a slice of ValidationError that the frontend can easily consume.
func ValidateStruct(data interface{}) []response.ValidationError {
	var errors []response.ValidationError
	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element response.ValidationError
			element.Field = err.Field()
			element.Message = msgForTag(err)
			errors = append(errors, element)
		}
	}
	return errors
}

// msgForTag converts common validation tags into user-friendly error messages
func msgForTag(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters long", err.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters long", err.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", err.Param())
	}
	return "Invalid value"
}
