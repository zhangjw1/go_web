package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/domain/service"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
)

// userServiceSimple is a simplified implementation of UserService
type userServiceSimple struct {
	userRepo  repository.UserRepository
	cache     *cache.Manager
	messaging *messaging.Manager
	logger    *logger.Logger
}

// NewUserServiceSimple creates a new simplified user service
func NewUserServiceSimple(
	userRepo repository.UserRepository,
	cache *cache.Manager,
	messaging *messaging.Manager,
	logger *logger.Logger,
) service.UserService {
	return &userServiceSimple{
		userRepo:  userRepo,
		cache:     cache,
		messaging: messaging,
		logger:    logger,
	}
}

// CreateUser creates a new user
func (s *userServiceSimple) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*model.User, error) {
	// Check if username already exists
	if existingUser, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if existingUser, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Create new user
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // In real implementation, this should be hashed
		Status:   model.UserStatusPending,
	}

	// Validate user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to create user")
		return nil, errors.New("failed to create user")
	}

	s.logger.WithField("user_id", user.ID).Info("User created successfully")
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userServiceSimple) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	userID := uint(id)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", id).Error("Failed to get user by ID")
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *userServiceSimple) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to get user by username")
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userServiceSimple) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.WithError(err).WithField("email", email).Error("Failed to get user by email")
		return nil, errors.New("user not found")
	}

	return user, nil
}

// ListUsers retrieves users with pagination
func (s *userServiceSimple) ListUsers(ctx context.Context, req *service.ListUsersRequest) (*service.ListUsersResponse, error) {
	// Create basic query params
	params := &model.QueryParams{
		PaginationParams: model.PaginationParams{
			Page:     req.Page,
			PageSize: req.PerPage,
		},
	}
	params.SetDefaults()

	result, err := s.userRepo.List(ctx, params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list users")
		return nil, errors.New("failed to get users")
	}

	users, ok := result.Data.([]*model.User)
	if !ok {
		s.logger.Error("Failed to cast users data")
		return nil, errors.New("failed to process users data")
	}

	return &service.ListUsersResponse{
		Users: users,
		Total: result.Total,
		Page:  result.Page,
	}, nil
}

// SearchUsers searches for users
func (s *userServiceSimple) SearchUsers(ctx context.Context, query string, limit int) ([]*model.User, error) {
	params := &model.PaginationParams{
		Page:     1,
		PageSize: limit,
	}
	
	result, err := s.userRepo.SearchUsers(ctx, query, params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to search users")
		return nil, errors.New("failed to search users")
	}

	users, ok := result.Data.([]*model.User)
	if !ok {
		s.logger.Error("Failed to cast users data")
		return nil, errors.New("failed to process users data")
	}

	return users, nil
}

// AuthenticateUser authenticates a user
func (s *userServiceSimple) AuthenticateUser(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// In real implementation, verify hashed password
	if user.Password != password {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive() {
		return nil, errors.New("user is not active")
	}

	return user, nil
}

// GetAllUsers retrieves all users with pagination (deprecated, use ListUsers)
func (s *userServiceSimple) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	// Create basic query params
	params := &model.QueryParams{
		PaginationParams: model.PaginationParams{
			Page:     1,
			PageSize: 50,
		},
	}
	params.SetDefaults()

	result, err := s.userRepo.List(ctx, params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get all users")
		return nil, errors.New("failed to get users")
	}

	users, ok := result.Data.([]*model.User)
	if !ok {
		s.logger.Error("Failed to cast users data")
		return nil, errors.New("failed to process users data")
	}

	return users, nil
}

// UpdateUser updates a user
func (s *userServiceSimple) UpdateUser(ctx context.Context, id int64, req *service.UpdateUserRequest) (*model.User, error) {
	// Get existing user
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FullName != nil {
		// Split full name into first and last name (simple approach)
		parts := strings.Fields(*req.FullName)
		if len(parts) > 0 {
			user.FirstName = parts[0]
			if len(parts) > 1 {
				user.LastName = strings.Join(parts[1:], " ")
			}
		}
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithField("user_id", id).Error("Failed to update user")
		return nil, errors.New("failed to update user")
	}

	s.logger.WithField("user_id", id).Info("User updated successfully")
	return user, nil
}

// DeleteUser deletes a user
func (s *userServiceSimple) DeleteUser(ctx context.Context, id int64) error {
	userID := uint(id)

	// Check if user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return errors.New("user not found")
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.WithError(err).WithField("user_id", id).Error("Failed to delete user")
		return errors.New("failed to delete user")
	}

	s.logger.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// ChangePassword changes a user's password
func (s *userServiceSimple) ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {
	// Get user
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// In a real implementation, you would verify the current password
	// For now, just update the password
	user.Password = newPassword

	// Save to database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithField("user_id", id).Error("Failed to change password")
		return errors.New("failed to change password")
	}

	s.logger.WithField("user_id", id).Info("Password changed successfully")
	return nil
}

// ActivateUser activates a user
func (s *userServiceSimple) ActivateUser(ctx context.Context, userID int64) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Activate()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to activate user")
		return errors.New("failed to activate user")
	}

	s.logger.WithField("user_id", userID).Info("User activated successfully")
	return nil
}

// DeactivateUser deactivates a user
func (s *userServiceSimple) DeactivateUser(ctx context.Context, userID int64) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Deactivate()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to deactivate user")
		return errors.New("failed to deactivate user")
	}

	s.logger.WithField("user_id", userID).Info("User deactivated successfully")
	return nil
}

// GetUserStats returns user statistics
func (s *userServiceSimple) GetUserStats(ctx context.Context) (*service.UserStats, error) {
	// TODO: Implement proper statistics
	return &service.UserStats{
		TotalUsers:        0,
		ActiveUsers:       0,
		InactiveUsers:     0,
		NewUsersToday:     0,
		NewUsersThisWeek:  0,
		NewUsersThisMonth: 0,
	}, nil
}