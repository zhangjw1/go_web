package dto

import (
	"strings"
	"time"

	"go-web-starter/internal/domain/model"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,username" validate:"required,username" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,password" validate:"required,password" example:"password123"`
	FullName string `json:"full_name" binding:"omitempty,max=100" validate:"omitempty,max=100" example:"John Doe"`
	Bio      string `json:"bio" binding:"omitempty,max=500" validate:"omitempty,max=500" example:"Software developer"`
	Avatar   string `json:"avatar" binding:"omitempty,url" validate:"omitempty,url" example:"https://example.com/avatar.jpg"`
	Phone    string `json:"phone" binding:"omitempty,phone" validate:"omitempty,phone" example:"+1234567890"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"omitempty,email" example:"john@example.com"`
	FullName string `json:"full_name" binding:"max=100" example:"John Doe"`
	Bio      string `json:"bio" binding:"max=500" example:"Software developer"`
	Avatar   string `json:"avatar" binding:"url" example:"https://example.com/avatar.jpg"`
}

// ChangePasswordRequest represents the request to change user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=100" example:"newpassword123"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Username  string    `json:"username" example:"john_doe"`
	Email     string    `json:"email" example:"john@example.com"`
	FullName  string    `json:"full_name" example:"John Doe"`
	Bio       string    `json:"bio" example:"Software developer"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// UserListResponse represents the user list response
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// UserQueryParams represents query parameters for user list
type UserQueryParams struct {
	Page     int    `form:"page,default=1" binding:"min=1" example:"1"`
	PerPage  int    `form:"per_page,default=10" binding:"min=1,max=100" example:"10"`
	Search   string `form:"search" example:"john"`
	SortBy   string `form:"sort_by,default=created_at" binding:"oneof=id username email created_at updated_at" example:"created_at"`
	SortDesc bool   `form:"sort_desc,default=false" example:"false"`
	IsActive *bool  `form:"is_active" example:"true"`
}

// ToUserResponse converts a User model to UserResponse DTO
func ToUserResponse(user *model.User) *UserResponse {
	response := &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		IsActive:  user.IsActive(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Include profile information if available
	if user.Profile != nil {
		response.FullName = user.GetFullName()
		response.Bio = user.Profile.Bio
		response.Avatar = user.Profile.Avatar
	} else {
		response.FullName = user.GetFullName()
	}

	return response
}

// ToUserListResponse converts a slice of User models to UserListResponse DTO
func ToUserListResponse(users []model.User, total int) *UserListResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *ToUserResponse(&user)
	}

	return &UserListResponse{
		Users: userResponses,
		Total: total,
	}
}

// ToUser converts CreateUserRequest to User model
func (req *CreateUserRequest) ToUser() *model.User {
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Note: This should be hashed before saving
		Status:   model.UserStatusActive,
	}

	// Parse full name into first and last name
	if req.FullName != "" {
		names := strings.Fields(req.FullName)
		if len(names) > 0 {
			user.FirstName = names[0]
		}
		if len(names) > 1 {
			user.LastName = strings.Join(names[1:], " ")
		}
	}

	// Create profile if additional information is provided
	if req.Bio != "" || req.Avatar != "" {
		user.Profile = &model.Profile{
			Bio:    req.Bio,
			Avatar: req.Avatar,
		}
	}

	return user
}

// ApplyToUser applies UpdateUserRequest to existing User model
func (req *UpdateUserRequest) ApplyToUser(user *model.User) {
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Parse full name into first and last name
	if req.FullName != "" {
		names := strings.Fields(req.FullName)
		if len(names) > 0 {
			user.FirstName = names[0]
		}
		if len(names) > 1 {
			user.LastName = strings.Join(names[1:], " ")
		}
	}

	// Update profile information
	if user.Profile == nil {
		user.Profile = &model.Profile{
			UserID: user.ID,
		}
	}

	if req.Bio != "" {
		user.Profile.Bio = req.Bio
	}
	if req.Avatar != "" {
		user.Profile.Avatar = req.Avatar
	}
}

// Validate performs additional validation on CreateUserRequest
func (req *CreateUserRequest) Validate() error {
	// Add custom validation logic here if needed
	// For example, check if username contains only allowed characters
	return nil
}

// Validate performs additional validation on UpdateUserRequest
func (req *UpdateUserRequest) Validate() error {
	// Add custom validation logic here if needed
	return nil
}

// Validate performs additional validation on ChangePasswordRequest
func (req *ChangePasswordRequest) Validate() error {
	// Add custom validation logic here if needed
	// For example, check password complexity
	return nil
}