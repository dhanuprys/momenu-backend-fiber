package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
)

// ValidatePayload dynamically validates a payload map against the field schema
// for a given EventType. Returns a list of validation errors consumable by the frontend.
func ValidatePayload(eventType models.EventType, payload map[string]interface{}) []response.ValidationError {
	schema := models.GetFieldSchema(eventType)
	if schema == nil {
		return []response.ValidationError{{Field: "event_type", Message: "Unknown event type"}}
	}

	var errors []response.ValidationError

	for _, group := range schema {
		for _, field := range group.Fields {
			fieldErrors := validateFieldValue(field, payload[field.Key], field.Key)
			errors = append(errors, fieldErrors...)
		}
	}

	return errors
}

// validateFieldValue validates a single field value against its definition.
// The prefix is used to construct nested field paths (e.g., "peserta[0].nama").
func validateFieldValue(field models.FieldDefinition, val interface{}, fieldPath string) []response.ValidationError {
	var errors []response.ValidationError

	exists := val != nil

	// Handle group type specially
	if field.Type == models.FieldTypeGroup {
		return validateGroupField(field, val, fieldPath)
	}

	// Check required
	if field.Required && (!exists || isEmptyValue(val)) {
		return append(errors, response.ValidationError{
			Field:   fieldPath,
			Message: fmt.Sprintf("%s is required", field.Label),
		})
	}

	// Skip validation if the field is not present and not required
	if !exists || isEmptyValue(val) {
		return errors
	}

	// Type check
	if err := checkFieldType(field, val); err != "" {
		return append(errors, response.ValidationError{
			Field:   fieldPath,
			Message: err,
		})
	}

	// Validation rules
	for _, rule := range field.Validations {
		if err := applyValidationRule(field, val, rule); err != "" {
			errors = append(errors, response.ValidationError{
				Field:   fieldPath,
				Message: err,
			})
		}
	}

	return errors
}

// validateGroupField handles validation for group (repeatable) field types.
func validateGroupField(field models.FieldDefinition, val interface{}, fieldPath string) []response.ValidationError {
	var errors []response.ValidationError

	if val == nil {
		if field.Required {
			errors = append(errors, response.ValidationError{
				Field:   fieldPath,
				Message: fmt.Sprintf("%s is required", field.Label),
			})
		}
		return errors
	}

	// Group values must be an array
	items, ok := val.([]interface{})
	if !ok {
		return append(errors, response.ValidationError{
			Field:   fieldPath,
			Message: fmt.Sprintf("%s must be an array of items", field.Label),
		})
	}

	// Validate min/max items
	if field.MinItems > 0 && len(items) < field.MinItems {
		errors = append(errors, response.ValidationError{
			Field:   fieldPath,
			Message: fmt.Sprintf("%s harus memiliki minimal %d entri", field.Label, field.MinItems),
		})
	}
	if field.MaxItems > 0 && len(items) > field.MaxItems {
		errors = append(errors, response.ValidationError{
			Field:   fieldPath,
			Message: fmt.Sprintf("%s maksimal %d entri", field.Label, field.MaxItems),
		})
	}

	// Validate each item's sub-fields
	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			errors = append(errors, response.ValidationError{
				Field:   fmt.Sprintf("%s[%d]", fieldPath, i),
				Message: fmt.Sprintf("Item %d in %s must be an object", i+1, field.Label),
			})
			continue
		}

		for _, subField := range field.Fields {
			subPath := fmt.Sprintf("%s[%d].%s", fieldPath, i, subField.Key)
			subErrors := validateFieldValue(subField, itemMap[subField.Key], subPath)
			errors = append(errors, subErrors...)
		}
	}

	return errors
}

func isEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}
	if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
		return true
	}
	return false
}

func checkFieldType(field models.FieldDefinition, val interface{}) string {
	switch field.Type {
	case models.FieldTypeString, models.FieldTypeText:
		if _, ok := val.(string); !ok {
			return fmt.Sprintf("%s must be a string", field.Label)
		}
	case models.FieldTypeNumber:
		// JSON numbers can arrive as float64
		if _, ok := val.(float64); !ok {
			return fmt.Sprintf("%s must be a number", field.Label)
		}
	case models.FieldTypeURL:
		s, ok := val.(string)
		if !ok {
			return fmt.Sprintf("%s must be a string", field.Label)
		}
		if _, err := url.ParseRequestURI(s); err != nil {
			return fmt.Sprintf("%s must be a valid URL", field.Label)
		}
	case models.FieldTypeImage:
		s, ok := val.(string)
		if !ok {
			return fmt.Sprintf("%s must be a string (file path)", field.Label)
		}
		// Security: only allow paths from our upload directory
		if !strings.HasPrefix(s, "/uploads/") {
			return fmt.Sprintf("%s must be a valid uploaded file path", field.Label)
		}
		// Security: prevent directory traversal
		if strings.Contains(s, "..") {
			return fmt.Sprintf("%s contains an invalid file path", field.Label)
		}
	}
	return ""
}

func applyValidationRule(field models.FieldDefinition, val interface{}, rule string) string {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	ruleName := parts[0]
	ruleParam := parts[1]
	paramNum, _ := strconv.Atoi(ruleParam)

	switch field.Type {
	case models.FieldTypeString, models.FieldTypeText, models.FieldTypeURL:
		s, _ := val.(string)
		switch ruleName {
		case "min":
			if len(s) < paramNum {
				return fmt.Sprintf("%s must be at least %s characters", field.Label, ruleParam)
			}
		case "max":
			if len(s) > paramNum {
				return fmt.Sprintf("%s must be at most %s characters", field.Label, ruleParam)
			}
		}
	case models.FieldTypeNumber:
		n, _ := val.(float64)
		switch ruleName {
		case "min":
			if n < float64(paramNum) {
				return fmt.Sprintf("%s must be at least %s", field.Label, ruleParam)
			}
		case "max":
			if n > float64(paramNum) {
				return fmt.Sprintf("%s must be at most %s", field.Label, ruleParam)
			}
		}
	}
	return ""
}
