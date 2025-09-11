package service

import (
	"context"
	"strings"

	"go-web-starter/internal/domain/model"
)

// UserService defines the interface for user business logic
type UserService interface {
	// User CRUD operations
	CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	
	// User list and search
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*model.User, error)
	
	// User authentication and authorization
	AuthenticateUser(ctx context.Context, username, password string) (*model.User, error)
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
	
	// User status management
	ActivateUser(ctx context.Context, userID int64) error
	DeactivateUser(ctx context.Context, userID int64) error
	
	// User statistics
	GetUserStats(ctx context.Context) (*UserStats, error)
}

// MessageService defines the interface for message handling business logic
type MessageService interface {
	// User event handling
	HandleUserCreated(ctx context.Context, userID int64, userData map[string]interface{}) error
	HandleUserUpdated(ctx context.Context, userID int64, changes map[string]interface{}) error
	HandleUserDeleted(ctx context.Context, userID int64) error
	
	// System event handling
	HandleSystemStartup(ctx context.Context, serviceData map[string]interface{}) error
	HandleSystemShutdown(ctx context.Context) error
	HandleHealthCheck(ctx context.Context, healthData map[string]interface{}) error
	
	// Notification handling
	HandleEmailNotification(ctx context.Context, recipient, subject, content string, data map[string]interface{}) error
	HandleSMSNotification(ctx context.Context, recipient, content string, data map[string]interface{}) error
	
	// Audit event handling
	HandleAuditEvent(ctx context.Context, userID int64, action, resource string, details map[string]interface{}, ipAddress, userAgent string) error
}

// Request and Response DTOs

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100,password"`
	FullName string `json:"full_name,omitempty" validate:"max=100"`
	Bio      string `json:"bio,omitempty" validate:"max=500"`
	Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,username"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	FullName *string `json:"full_name,omitempty" validate:"omitempty,max=100"`
	Bio      *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Avatar   *string `json:"avatar,omitempty" validate:"omitempty,url"`
}

// ListUsersRequest represents a request to list users with pagination and filtering
type ListUsersRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PerPage  int    `json:"per_page" validate:"min=1,max=100"`
	Search   string `json:"search,omitempty"`
	SortBy   string `json:"sort_by,omitempty" validate:"omitempty,oneof=id username email created_at updated_at"`
	SortDesc bool   `json:"sort_desc,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []*model.User `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers   int64 `json:"total_users"`
	ActiveUsers  int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	NewUsersToday int64 `json:"new_users_today"`
	NewUsersThisWeek int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`
}

// Service errors
type ServiceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *ServiceError) Error() string {
	return e.Message
}

// Common service errors
var (
	ErrUserNotFound      = &ServiceError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrUserAlreadyExists = &ServiceError{Code: "USER_ALREADY_EXISTS", Message: "User already exists"}
	ErrInvalidCredentials = &ServiceError{Code: "INVALID_CREDENTIALS", Message: "Invalid credentials"}
	ErrInvalidPassword   = &ServiceError{Code: "INVALID_PASSWORD", Message: "Invalid password"}
	ErrUserInactive      = &ServiceError{Code: "USER_INACTIVE", Message: "User is inactive"}
	ErrInvalidRequest    = &ServiceError{Code: "INVALID_REQUEST", Message: "Invalid request"}
	ErrInternalError     = &ServiceError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrUnauthorized      = &ServiceError{Code: "UNAUTHORIZED", Message: "Unauthorized"}
	ErrForbidden         = &ServiceError{Code: "FORBIDDEN", Message: "Forbidden"}
)

// NewServiceError creates a new service error with details
func NewServiceError(code, message string, details interface{}) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Validation helper functions

// ValidateCreateUserRequest validates a create user request
func ValidateCreateUserRequest(req *CreateUserRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}
	
	if req.Username == "" {
		return NewServiceError("VALIDATION_ERROR", "Username is required", nil)
	}
	
	if req.Email == "" {
		return NewServiceError("VALIDATION_ERROR", "Email is required", nil)
	}
	
	if req.Password == "" {
		return NewServiceError("VALIDATION_ERROR", "Password is required", nil)
	}
	
	return nil
}

// ValidateUpdateUserRequest validates an update user request
func ValidateUpdateUserRequest(req *UpdateUserRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}
	
	// At least one field should be provided for update
	if req.Username == nil && req.Email == nil && req.FullName == nil && req.Bio == nil && req.Avatar == nil {
		return NewServiceError("VALIDATION_ERROR", "At least one field must be provided for update", nil)
	}
	
	return nil
}

// ValidateListUsersRequest validates a list users request
func ValidateListUsersRequest(req *ListUsersRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}
	
	if req.Page < 1 {
		req.Page = 1
	}
	
	if req.PerPage < 1 {
		req.PerPage = 10
	}
	
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	
	return nil
}

// Business logic helper functions

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	
	return pages
}

// IsValidSortField checks if the sort field is valid
func IsValidSortField(field string) bool {
	validFields := []string{"id", "username", "email", "created_at", "updated_at"}
	for _, validField := range validFields {
		if field == validField {
			return true
		}
	}
	return false
}

// SanitizeSearchQuery sanitizes a search query
func SanitizeSearchQuery(query string) string {
	// Remove leading and trailing whitespace
	query = strings.TrimSpace(query)
	
	// Limit length
	if len(query) > 100 {
		query = query[:100]
	}
	
	return query
}

// Context keys for request metadata
type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyTraceID   contextKey = "trace_id"
	ContextKeyIPAddress contextKey = "ip_address"
	ContextKeyUserAgent contextKey = "user_agent"
)

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(int64)
	return userID, ok
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(ContextKeyRequestID).(string)
	return requestID, ok
}

// GetTraceIDFromContext extracts trace ID from context
func GetTraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(ContextKeyTraceID).(string)
	return traceID, ok
}

// GetIPAddressFromContext extracts IP address from context
func GetIPAddressFromContext(ctx context.Context) (string, bool) {
	ipAddress, ok := ctx.Value(ContextKeyIPAddress).(string)
	return ipAddress, ok
}

// GetUserAgentFromContext extracts user agent from context
func GetUserAgentFromContext(ctx context.Context) (string, bool) {
	userAgent, ok := ctx.Value(ContextKeyUserAgent).(string)
	return userAgent, ok
}