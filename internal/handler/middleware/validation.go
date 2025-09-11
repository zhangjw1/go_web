package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"go-web-starter/internal/handler/response"
	customValidator "go-web-starter/internal/handler/validator"
)

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// ValidateJSON validates JSON request body against a struct
func ValidateJSON[T any](c *gin.Context, obj *T) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		handleValidationError(c, err)
		return false
	}
	return true
}

// ValidateQuery validates query parameters against a struct
func ValidateQuery[T any](c *gin.Context, obj *T) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		handleValidationError(c, err)
		return false
	}
	return true
}

// ValidateURI validates URI parameters against a struct
func ValidateURI[T any](c *gin.Context, obj *T) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		handleValidationError(c, err)
		return false
	}
	return true
}

// ValidateForm validates form data against a struct
func ValidateForm[T any](c *gin.Context, obj *T) bool {
	if err := c.ShouldBind(obj); err != nil {
		handleValidationError(c, err)
		return false
	}
	return true
}

// ValidateStruct validates a struct using custom validator
func ValidateStruct(c *gin.Context, obj interface{}) bool {
	cv := customValidator.New()
	if err := cv.ValidateStruct(obj); err != nil {
		if validationErrors, ok := err.(customValidator.ValidationErrors); ok {
			response.UnprocessableEntity(c, "Validation failed", formatValidationErrors(validationErrors))
			return false
		}
		response.BadRequest(c, "Validation error", err.Error())
		return false
	}
	return true
}

// handleValidationError handles different types of validation errors
func handleValidationError(c *gin.Context, err error) {
	switch e := err.(type) {
	case validator.ValidationErrors:
		errors := make([]map[string]interface{}, 0)
		for _, fieldError := range e {
			errors = append(errors, map[string]interface{}{
				"field":   getJSONFieldName(fieldError),
				"tag":     fieldError.Tag(),
				"value":   fieldError.Value(),
				"message": getValidationErrorMessage(fieldError),
			})
		}
		response.UnprocessableEntity(c, "Validation failed", errors)
	case *validator.InvalidValidationError:
		response.BadRequest(c, "Invalid validation", e.Error())
	default:
		// Handle JSON syntax errors, type conversion errors, etc.
		if isJSONSyntaxError(err) {
			response.BadRequest(c, "Invalid JSON format", err.Error())
		} else {
			response.BadRequest(c, "Request binding failed", err.Error())
		}
	}
}

// formatValidationErrors formats custom validation errors
func formatValidationErrors(errors customValidator.ValidationErrors) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(errors))
	for i, err := range errors {
		formatted[i] = map[string]interface{}{
			"field":   err.Field,
			"tag":     err.Tag,
			"value":   err.Value,
			"message": err.Message,
		}
	}
	return formatted
}

// getJSONFieldName extracts the JSON field name from struct field
func getJSONFieldName(fe validator.FieldError) string {
	// This would ideally use reflection to get the JSON tag
	// For now, just return the field name
	return fe.Field()
}

// getValidationErrorMessage returns user-friendly validation error messages
func getValidationErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + param + " characters long"
	case "max":
		return field + " must be at most " + param + " characters long"
	case "len":
		return field + " must be exactly " + param + " characters long"
	case "oneof":
		return field + " must be one of: " + param
	case "url":
		return field + " must be a valid URL"
	case "alpha":
		return field + " must contain only alphabetic characters"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "numeric":
		return field + " must be numeric"
	case "uuid":
		return field + " must be a valid UUID"
	case "gte":
		return field + " must be greater than or equal to " + param
	case "lte":
		return field + " must be less than or equal to " + param
	case "gt":
		return field + " must be greater than " + param
	case "lt":
		return field + " must be less than " + param
	case "password":
		return field + " must meet password requirements (at least 8 characters with mix of letters, numbers, and symbols)"
	case "username":
		return field + " must be a valid username (3-30 characters, start with letter, contain only letters, numbers, underscore, hyphen)"
	case "phone":
		return field + " must be a valid phone number"
	default:
		return field + " is invalid"
	}
}

// isJSONSyntaxError checks if the error is a JSON syntax error
func isJSONSyntaxError(err error) bool {
	// Check for common JSON syntax error patterns
	errStr := err.Error()
	return contains(errStr, "invalid character") ||
		contains(errStr, "unexpected end of JSON input") ||
		contains(errStr, "cannot unmarshal")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of substring in string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// RequestValidationMiddleware creates a middleware that validates requests
func RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set custom error handler for validation
		c.Set("validation_handler", handleValidationError)
		c.Next()
	}
}

// WithValidation is a helper function to wrap handlers with validation
func WithValidation[T any](handler func(*gin.Context, *T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if !ValidateJSON(c, &req) {
			return
		}

		// Additional struct validation
		if !ValidateStruct(c, &req) {
			return
		}

		handler(c, &req)
	}
}

// WithQueryValidation is a helper function to wrap handlers with query validation
func WithQueryValidation[T any](handler func(*gin.Context, *T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if !ValidateQuery(c, &req) {
			return
		}

		// Additional struct validation
		if !ValidateStruct(c, &req) {
			return
		}

		handler(c, &req)
	}
}
