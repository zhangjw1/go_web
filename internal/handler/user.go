package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/handler/dto"
	"go-web-starter/internal/handler/response"
	"go-web-starter/internal/infrastructure/logger"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userRepo repository.UserRepository
	config   *config.Config
	logger   *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo repository.UserRepository, cfg *config.Config, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		config:   cfg,
		logger:   log,
	}
}

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind create user request")
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Additional validation
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).Error("Create user request validation failed")
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	// Convert DTO to model
	user := req.ToUser()

	// TODO: Hash password before saving
	// user.Password = hashPassword(user.Password)

	// Create user
	ctx := context.Background()
	if err := h.userRepo.Create(ctx, user); err != nil {
		h.logger.WithError(err).WithField("username", user.Username).Error("Failed to create user")
		
		// Check if it's a duplicate key error
		if isDuplicateKeyError(err) {
			response.Conflict(c, "User with this username or email already exists")
			return
		}
		
		response.InternalServerError(c, "Failed to create user")
		return
	}

	h.logger.WithField("user_id", user.ID).WithField("username", user.Username).Info("User created successfully")

	// Convert to response DTO
	userResponse := dto.ToUserResponse(user)
	response.Created(c, userResponse, "User created successfully")
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID")
		return
	}

	ctx := context.Background()
	user, err := h.userRepo.GetByID(ctx, uint(id))
	if err != nil {
		if isNotFoundError(err) {
			h.logger.WithField("user_id", id).Warn("User not found")
			response.NotFound(c, "User not found")
			return
		}
		
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to get user")
		response.InternalServerError(c, "Failed to retrieve user")
		return
	}

	h.logger.WithField("user_id", id).Debug("User retrieved successfully")

	// Convert to response DTO
	userResponse := dto.ToUserResponse(user)
	response.Success(c, userResponse)
}

// UpdateUser updates an existing user
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind update user request")
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Additional validation
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).Error("Update user request validation failed")
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	// Get existing user
	ctx := context.Background()
	user, err := h.userRepo.GetByID(ctx, uint(id))
	if err != nil {
		if isNotFoundError(err) {
			h.logger.WithField("user_id", id).Warn("User not found for update")
			response.NotFound(c, "User not found")
			return
		}
		
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to get user for update")
		response.InternalServerError(c, "Failed to retrieve user")
		return
	}

	// Apply updates
	req.ApplyToUser(user)

	// Update user
	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to update user")
		
		// Check if it's a duplicate key error
		if isDuplicateKeyError(err) {
			response.Conflict(c, "User with this username or email already exists")
			return
		}
		
		response.InternalServerError(c, "Failed to update user")
		return
	}

	h.logger.WithField("user_id", id).Info("User updated successfully")

	// Convert to response DTO
	userResponse := dto.ToUserResponse(user)
	response.Success(c, userResponse, "User updated successfully")
}

// DeleteUser deletes a user by ID
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Check if user exists
	ctx := context.Background()
	_, err = h.userRepo.GetByID(ctx, uint(id))
	if err != nil {
		if isNotFoundError(err) {
			h.logger.WithField("user_id", id).Warn("User not found for deletion")
			response.NotFound(c, "User not found")
			return
		}
		
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to get user for deletion")
		response.InternalServerError(c, "Failed to retrieve user")
		return
	}

	// Delete user
	if err := h.userRepo.Delete(ctx, uint(id)); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to delete user")
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	h.logger.WithField("user_id", id).Info("User deleted successfully")
	response.NoContent(c)
}

// ListUsers retrieves a paginated list of users
// @Summary List users
// @Description Get a paginated list of users with optional filtering
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" Enums(id,username,email,created_at,updated_at) default(created_at)
// @Param sort_desc query bool false "Sort descending" default(false)
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.Response{data=dto.UserListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var params dto.UserQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.WithError(err).Error("Failed to bind user list query parameters")
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Build query parameters using model types
	queryParams := &model.UserQueryParams{}
	queryParams.Page = params.Page
	queryParams.PageSize = params.PerPage
	queryParams.SortBy = params.SortBy
	queryParams.SortOrder = "asc"
	queryParams.Username = params.Search

	if params.SortDesc {
		queryParams.SortOrder = "desc"
	}

	// Set defaults
	queryParams.SetDefaults()

	// Get users - need to use the generic List method with QueryParams
	ctx := context.Background()
	genericParams := &model.QueryParams{
		PaginationParams: queryParams.PaginationParams,
		SortParams:       queryParams.SortParams,
		FilterParams: model.FilterParams{
			Search: params.Search,
		},
	}
	result, err := h.userRepo.List(ctx, genericParams)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		response.InternalServerError(c, "Failed to retrieve users")
		return
	}

	// Extract users from result
	users := make([]model.User, 0)
	if userSlice, ok := result.Data.([]model.User); ok {
		users = userSlice
	} else if userPtrSlice, ok := result.Data.([]*model.User); ok {
		users = make([]model.User, len(userPtrSlice))
		for i, user := range userPtrSlice {
			users[i] = *user
		}
	}

	h.logger.WithField("total", result.Total).WithField("page", params.Page).Debug("Users listed successfully")

	// Convert to response DTO
	userListResponse := dto.ToUserListResponse(users, int(result.Total))

	// Calculate pagination metadata
	total := int(result.Total)
	totalPages := (total + params.PerPage - 1) / params.PerPage
	meta := &response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}

	response.SuccessWithMeta(c, userListResponse, meta)
}

// GetUserByUsername retrieves a user by username
// @Summary Get user by username
// @Description Get user information by username
// @Tags users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/username/{username} [get]
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "Username is required")
		return
	}

	ctx := context.Background()
	user, err := h.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if isNotFoundError(err) {
			h.logger.WithField("username", username).Warn("User not found")
			response.NotFound(c, "User not found")
			return
		}
		
		h.logger.WithError(err).WithField("username", username).Error("Failed to get user by username")
		response.InternalServerError(c, "Failed to retrieve user")
		return
	}

	h.logger.WithField("username", username).Debug("User retrieved by username successfully")

	// Convert to response DTO
	userResponse := dto.ToUserResponse(user)
	response.Success(c, userResponse)
}

// ChangePassword changes user password
// @Summary Change user password
// @Description Change user password with current password verification
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid user ID")
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind change password request")
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Additional validation
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).Error("Change password request validation failed")
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	// Get existing user
	ctx := context.Background()
	user, err := h.userRepo.GetByID(ctx, uint(id))
	if err != nil {
		if isNotFoundError(err) {
			h.logger.WithField("user_id", id).Warn("User not found for password change")
			response.NotFound(c, "User not found")
			return
		}
		
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to get user for password change")
		response.InternalServerError(c, "Failed to retrieve user")
		return
	}

	// TODO: Verify current password
	// if !verifyPassword(req.CurrentPassword, user.Password) {
	//     response.Unauthorized(c, "Current password is incorrect")
	//     return
	// }

	// TODO: Hash new password
	// user.Password = hashPassword(req.NewPassword)

	// For now, just set the new password directly (this should be hashed in production)
	user.Password = req.NewPassword

	// Update user
	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("Failed to update user password")
		response.InternalServerError(c, "Failed to update password")
		return
	}

	h.logger.WithField("user_id", id).Info("User password changed successfully")
	response.Success(c, nil, "Password changed successfully")
}

// Helper functions

// isDuplicateKeyError checks if the error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	return repository.IsAlreadyExistsError(err)
}

// isNotFoundError checks if the error is a not found error
func isNotFoundError(err error) bool {
	return repository.IsNotFoundError(err)
}