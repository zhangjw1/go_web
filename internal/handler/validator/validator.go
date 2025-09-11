package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// New creates a new custom validator
func New() *CustomValidator {
	v := validator.New()
	
	// Register custom tag name function to use json tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	return &CustomValidator{
		validator: v,
	}
}

// ValidateStruct validates a struct and returns formatted errors
func (cv *CustomValidator) ValidateStruct(obj interface{}) error {
	if err := cv.validator.Struct(obj); err != nil {
		var validationErrors ValidationErrors
		
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: getErrorMessage(err),
			})
		}
		
		return validationErrors
	}
	return nil
}

// Engine returns the underlying validator engine
func (cv *CustomValidator) Engine() interface{} {
	return cv.validator
}

// getErrorMessage returns a user-friendly error message for validation errors
func getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	value := fmt.Sprintf("%v", fe.Value())

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters long", field, param)
		}
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters long", field, param)
		}
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "unique":
		return fmt.Sprintf("%s must be unique", field)
	case "password":
		return fmt.Sprintf("%s must meet password requirements", field)
	case "username":
		return fmt.Sprintf("%s must be a valid username", field)
	default:
		return fmt.Sprintf("%s is invalid (value: %s, rule: %s)", field, value, tag)
	}
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Password validation
	v.RegisterValidation("password", validatePassword)
	
	// Username validation
	v.RegisterValidation("username", validateUsername)
	
	// Phone number validation
	v.RegisterValidation("phone", validatePhone)
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// Minimum length check
	if len(password) < 8 {
		return false
	}
	
	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false
	
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}
	
	// Require at least 3 out of 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}
	
	return count >= 3
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	
	// Length check
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	
	// Must start with a letter
	if username[0] < 'a' || (username[0] > 'z' && username[0] < 'A') || username[0] > 'Z' {
		return false
	}
	
	// Can only contain letters, numbers, underscores, and hyphens
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_' || char == '-') {
			return false
		}
	}
	
	return true
}

// validatePhone validates phone number format
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	
	// Remove common separators
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	
	// Check if it starts with + (international format)
	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}
	
	// Must be all digits and reasonable length
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}
	
	for _, char := range cleaned {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

// InitGinValidator initializes Gin's validator with our custom validator
func InitGinValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom tag name function
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		
		// Register custom validators
		registerCustomValidators(v)
	}
}