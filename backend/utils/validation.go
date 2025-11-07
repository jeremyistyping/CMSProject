package utils

import (
	"github.com/go-playground/validator/v10"
)

// FormatValidationError converts validator errors into a simple field -> message map
func FormatValidationError(err error) map[string]string {
	if err == nil {
		return nil
	}

	// If it's a validator.ValidationErrors, format per-field messages
	if ve, ok := err.(validator.ValidationErrors); ok {
		out := make(map[string]string)
		for _, fe := range ve {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				out[field] = field + " is required"
			case "email":
				out[field] = field + " must be a valid email address"
			case "min":
				out[field] = field + " must be at least " + fe.Param() + " characters long"
			case "max":
				out[field] = field + " must be at most " + fe.Param() + " characters long"
			case "oneof":
				out[field] = field + " must be one of: " + fe.Param()
			case "alphanum":
				out[field] = field + " must contain only alphanumeric characters"
			default:
				out[field] = field + " is invalid"
			}
		}
		return out
	}

	// Fallback for non-validation errors
	return map[string]string{
		"_error": err.Error(),
	}
}
