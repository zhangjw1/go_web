package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/infrastructure/logger"
)

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	*BaseRepository[*model.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, log *logger.Logger) repository.UserRepository {
	baseRepo := NewBaseRepository(db, log, &model.User{})
	return &UserRepositoryImpl{
		BaseRepository: baseRepo,
	}
}

// GetByUsername retrieves a user by username
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if username == "" {
		return nil, repository.NewRepositoryError("get_by_username", "User", 0, "username is required", nil)
	}

	var user model.User
	result := r.db.WithContext(ctx).Where("username = ?", strings.ToLower(username)).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithField("username", username).Debug("User not found by username")
			return nil, repository.NewRepositoryError("get_by_username", "User", 0, "entity not found", result.Error)
		}

		r.logger.WithError(result.Error).WithField("username", username).Error("Failed to get user by username")
		return nil, repository.NewRepositoryError("get_by_username", "User", 0, "database error", result.Error)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, repository.NewRepositoryError("get_by_email", "User", 0, "email is required", nil)
	}

	var user model.User
	result := r.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithField("email", email).Debug("User not found by email")
			return nil, repository.NewRepositoryError("get_by_email", "User", 0, "entity not found", result.Error)
		}

		r.logger.WithError(result.Error).WithField("email", email).Error("Failed to get user by email")
		return nil, repository.NewRepositoryError("get_by_email", "User", 0, "database error", result.Error)
	}

	return &user, nil
}

// GetWithProfile retrieves a user with their profile
func (r *UserRepositoryImpl) GetWithProfile(ctx context.Context, id uint) (*model.User, error) {
	if id == 0 {
		return nil, repository.NewRepositoryError("get_with_profile", "User", id, "invalid ID", nil)
	}

	var user model.User
	result := r.db.WithContext(ctx).Preload("Profile").First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithField("id", id).Debug("User not found")
			return nil, repository.NewRepositoryError("get_with_profile", "User", id, "entity not found", result.Error)
		}

		r.logger.WithError(result.Error).WithField("id", id).Error("Failed to get user with profile")
		return nil, repository.NewRepositoryError("get_with_profile", "User", id, "database error", result.Error)
	}

	return &user, nil
}

// ListWithProfiles retrieves users with their profiles
func (r *UserRepositoryImpl) ListWithProfiles(ctx context.Context, params *model.UserQueryParams) (*model.PaginationResult, error) {
	if params == nil {
		params = &model.UserQueryParams{}
	}
	params.SetDefaults()

	// Build query
	query := r.db.WithContext(ctx).Model(&model.User{}).Preload("Profile")

	// Apply filters
	query = r.applyUserFilters(query, &params.UserFilter)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count users with profiles")
		return nil, repository.NewRepositoryError("list_with_profiles", "User", 0, "database error", err)
	}

	// Apply pagination and sorting
	var users []model.User
	query = query.Order(params.UserSort.GetOrderBy()).
		Offset(params.GetOffset()).
		Limit(params.GetLimit())

	if err := query.Find(&users).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list users with profiles")
		return nil, repository.NewRepositoryError("list_with_profiles", "User", 0, "database error", err)
	}

	// Convert to pointers for consistency
	userPtrs := make([]*model.User, len(users))
	for i := range users {
		userPtrs[i] = &users[i]
	}

	result := model.NewPaginationResult(&params.PaginationParams, total, userPtrs)
	r.logger.WithField("total", total).WithField("page", params.Page).Debug("Users with profiles listed successfully")

	return result, nil
}

// UpdateLastLogin updates the user's last login time and count
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, id uint) error {
	if id == 0 {
		return repository.NewRepositoryError("update_last_login", "User", id, "invalid ID", nil)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": now,
		"login_count":   gorm.Expr("login_count + 1"),
		"updated_at":    now,
	})

	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("id", id).Error("Failed to update last login")
		return repository.NewRepositoryError("update_last_login", "User", id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("id", id).Debug("User not found for last login update")
		return repository.NewRepositoryError("update_last_login", "User", id, "entity not found", nil)
	}

	r.logger.WithField("id", id).Info("User last login updated successfully")
	return nil
}

// UpdateEmailVerification updates the user's email verification status
func (r *UserRepositoryImpl) UpdateEmailVerification(ctx context.Context, id uint, verified bool) error {
	if id == 0 {
		return repository.NewRepositoryError("update_email_verification", "User", id, "invalid ID", nil)
	}

	updates := map[string]interface{}{
		"email_verified": verified,
		"updated_at":     time.Now(),
	}

	if verified {
		updates["email_verified_at"] = time.Now()
	} else {
		updates["email_verified_at"] = nil
	}

	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("id", id).WithField("verified", verified).Error("Failed to update email verification")
		return repository.NewRepositoryError("update_email_verification", "User", id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("id", id).Debug("User not found for email verification update")
		return repository.NewRepositoryError("update_email_verification", "User", id, "entity not found", nil)
	}

	r.logger.WithField("id", id).WithField("verified", verified).Info("User email verification updated successfully")
	return nil
}

// UpdateStatus updates the user's status
func (r *UserRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status model.UserStatus) error {
	if id == 0 {
		return repository.NewRepositoryError("update_status", "User", id, "invalid ID", nil)
	}

	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("id", id).WithField("status", status).Error("Failed to update user status")
		return repository.NewRepositoryError("update_status", "User", id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("id", id).Debug("User not found for status update")
		return repository.NewRepositoryError("update_status", "User", id, "entity not found", nil)
	}

	r.logger.WithField("id", id).WithField("status", status).Info("User status updated successfully")
	return nil
}

// GetActiveUsers retrieves all active users
func (r *UserRepositoryImpl) GetActiveUsers(ctx context.Context, params *model.PaginationParams) (*model.PaginationResult, error) {
	if params == nil {
		params = &model.PaginationParams{}
	}
	params.SetDefaults()

	// Count total active users
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{}).Where("status = ?", model.UserStatusActive)
	if err := query.Count(&total).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count active users")
		return nil, repository.NewRepositoryError("get_active_users", "User", 0, "database error", err)
	}

	// Get paginated active users
	var users []model.User
	if err := query.Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get active users")
		return nil, repository.NewRepositoryError("get_active_users", "User", 0, "database error", err)
	}

	// Convert to pointers
	userPtrs := make([]*model.User, len(users))
	for i := range users {
		userPtrs[i] = &users[i]
	}

	result := model.NewPaginationResult(params, total, userPtrs)
	r.logger.WithField("total", total).WithField("page", params.Page).Debug("Active users retrieved successfully")

	return result, nil
}

// SearchUsers searches users by username, email, or name
func (r *UserRepositoryImpl) SearchUsers(ctx context.Context, query string, params *model.PaginationParams) (*model.PaginationResult, error) {
	if params == nil {
		params = &model.PaginationParams{}
	}
	params.SetDefaults()

	if query == "" {
		return r.GetActiveUsers(ctx, params)
	}

	searchPattern := "%" + strings.ToLower(query) + "%"

	// Build search query
	dbQuery := r.db.WithContext(ctx).Model(&model.User{}).Where(
		"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// Count total matching users
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		r.logger.WithError(err).WithField("query", query).Error("Failed to count search results")
		return nil, repository.NewRepositoryError("search_users", "User", 0, "database error", err)
	}

	// Get paginated search results
	var users []model.User
	if err := dbQuery.Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users).Error; err != nil {
		r.logger.WithError(err).WithField("query", query).Error("Failed to search users")
		return nil, repository.NewRepositoryError("search_users", "User", 0, "database error", err)
	}

	// Convert to pointers
	userPtrs := make([]*model.User, len(users))
	for i := range users {
		userPtrs[i] = &users[i]
	}

	result := model.NewPaginationResult(params, total, userPtrs)
	r.logger.WithField("query", query).WithField("total", total).WithField("page", params.Page).Debug("User search completed successfully")

	return result, nil
}

// List overrides the base List method to apply user-specific filters
func (r *UserRepositoryImpl) List(ctx context.Context, params *model.QueryParams) (*model.PaginationResult, error) {
	if params == nil {
		params = &model.QueryParams{}
	}
	params.SetDefaults()

	// Build query
	query := r.db.WithContext(ctx).Model(&model.User{})

	// Apply search filter
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count users")
		return nil, repository.NewRepositoryError("list", "User", 0, "database error", err)
	}

	// Apply pagination and sorting
	var users []model.User
	query = query.Order(params.SortParams.GetOrderBy()).
		Offset(params.GetOffset()).
		Limit(params.GetLimit())

	if err := query.Find(&users).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list users")
		return nil, repository.NewRepositoryError("list", "User", 0, "database error", err)
	}

	// Convert to pointers
	userPtrs := make([]*model.User, len(users))
	for i := range users {
		userPtrs[i] = &users[i]
	}

	result := model.NewPaginationResult(&params.PaginationParams, total, userPtrs)
	r.logger.WithField("total", total).WithField("page", params.Page).Debug("Users listed successfully")

	return result, nil
}

// applyUserFilters applies user-specific filters to the query
func (r *UserRepositoryImpl) applyUserFilters(query *gorm.DB, filter *model.UserFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Search filter
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Username filter
	if filter.Username != "" {
		query = query.Where("username = ?", strings.ToLower(filter.Username))
	}

	// Email filter
	if filter.Email != "" {
		query = query.Where("email = ?", strings.ToLower(filter.Email))
	}

	// Status filter
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Email verified filter
	if filter.EmailVerified != nil {
		query = query.Where("email_verified = ?", *filter.EmailVerified)
	}

	// Created at filter
	if filter.CreatedAt != nil {
		if filter.CreatedAt.From != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAt.From)
		}
		if filter.CreatedAt.To != nil {
			query = query.Where("created_at <= ?", *filter.CreatedAt.To)
		}
	}

	// Updated at filter
	if filter.UpdatedAt != nil {
		if filter.UpdatedAt.From != nil {
			query = query.Where("updated_at >= ?", *filter.UpdatedAt.From)
		}
		if filter.UpdatedAt.To != nil {
			query = query.Where("updated_at <= ?", *filter.UpdatedAt.To)
		}
	}

	return query
}

// WithTransaction returns a new user repository instance using the provided transaction
func (r *UserRepositoryImpl) WithTransaction(tx *gorm.DB) repository.UserRepository {
	baseRepo := r.BaseRepository.WithTransaction(tx)
	return &UserRepositoryImpl{
		BaseRepository: baseRepo,
	}
}
