package repository

import (
	"context"

	"go-web-starter/internal/domain/model"
)

// Repository defines the common interface for all repositories
type Repository[T model.Model] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity T) error

	// GetByID retrieves an entity by its ID
	GetByID(ctx context.Context, id uint) (T, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity T) error

	// Delete soft deletes an entity by ID
	Delete(ctx context.Context, id uint) error

	// HardDelete permanently deletes an entity by ID
	HardDelete(ctx context.Context, id uint) error

	// List retrieves entities with pagination and filtering
	List(ctx context.Context, params *model.QueryParams) (*model.PaginationResult, error)

	// Count returns the total count of entities matching the filter
	Count(ctx context.Context, filter interface{}) (int64, error)

	// Exists checks if an entity exists by ID
	Exists(ctx context.Context, id uint) (bool, error)
}

// UserRepository defines the interface for user-specific operations
type UserRepository interface {
	Repository[*model.User]

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*model.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// GetWithProfile retrieves a user with their profile
	GetWithProfile(ctx context.Context, id uint) (*model.User, error)

	// ListWithProfiles retrieves users with their profiles
	ListWithProfiles(ctx context.Context, params *model.UserQueryParams) (*model.PaginationResult, error)

	// UpdateLastLogin updates the user's last login time and count
	UpdateLastLogin(ctx context.Context, id uint) error

	// UpdateEmailVerification updates the user's email verification status
	UpdateEmailVerification(ctx context.Context, id uint, verified bool) error

	// UpdateStatus updates the user's status
	UpdateStatus(ctx context.Context, id uint, status model.UserStatus) error

	// GetActiveUsers retrieves all active users
	GetActiveUsers(ctx context.Context, params *model.PaginationParams) (*model.PaginationResult, error)

	// SearchUsers searches users by username, email, or name
	SearchUsers(ctx context.Context, query string, params *model.PaginationParams) (*model.PaginationResult, error)
}

// ProfileRepository defines the interface for profile-specific operations
type ProfileRepository interface {
	Repository[*model.Profile]

	// GetByUserID retrieves a profile by user ID
	GetByUserID(ctx context.Context, userID uint) (*model.Profile, error)

	// GetWithUser retrieves a profile with the associated user
	GetWithUser(ctx context.Context, id uint) (*model.Profile, error)

	// UpdateAvatar updates the user's avatar
	UpdateAvatar(ctx context.Context, userID uint, avatar string) error

	// SearchByLocation searches profiles by country and/or city
	SearchByLocation(ctx context.Context, country, city string, params *model.PaginationParams) (*model.PaginationResult, error)

	// GetByCompany retrieves profiles by company name
	GetByCompany(ctx context.Context, company string, params *model.PaginationParams) (*model.PaginationResult, error)
}

// RepositoryManager manages all repositories
type RepositoryManager interface {
	// User returns the user repository
	User() UserRepository

	// Profile returns the profile repository
	Profile() ProfileRepository

	// Transaction executes a function within a database transaction
	Transaction(ctx context.Context, fn func(RepositoryManager) error) error

	// Close closes all repository connections
	Close() error
}

// TransactionManager defines the interface for transaction management
type TransactionManager interface {
	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// User returns the user repository within this transaction
	User() UserRepository

	// Profile returns the profile repository within this transaction
	Profile() ProfileRepository
}

// RepositoryError represents repository-specific errors
type RepositoryError struct {
	Op      string // Operation that failed
	Entity  string // Entity type
	ID      uint   // Entity ID (if applicable)
	Message string // Error message
	Err     error  // Underlying error
}

func (e *RepositoryError) Error() string {
	if e.ID > 0 {
		return e.Op + " " + e.Entity + " (ID: " + string(rune(e.ID)) + "): " + e.Message
	}
	return e.Op + " " + e.Entity + ": " + e.Message
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Common repository errors
var (
	ErrNotFound      = &RepositoryError{Message: "entity not found"}
	ErrAlreadyExists = &RepositoryError{Message: "entity already exists"}
	ErrInvalidInput  = &RepositoryError{Message: "invalid input"}
	ErrDatabase      = &RepositoryError{Message: "database error"}
	ErrUserNotFound  = &RepositoryError{Message: "user not found"}
)

// NewRepositoryError creates a new repository error
func NewRepositoryError(op, entity string, id uint, message string, err error) *RepositoryError {
	return &RepositoryError{
		Op:      op,
		Entity:  entity,
		ID:      id,
		Message: message,
		Err:     err,
	}
}

// IsNotFoundError checks if the error is a "not found" error
func IsNotFoundError(err error) bool {
	if repoErr, ok := err.(*RepositoryError); ok {
		return repoErr.Message == "entity not found"
	}
	return false
}

// IsAlreadyExistsError checks if the error is an "already exists" error
func IsAlreadyExistsError(err error) bool {
	if repoErr, ok := err.(*RepositoryError); ok {
		return repoErr.Message == "entity already exists"
	}
	return false
}

// UserQueryParams represents query parameters for user operations
type UserQueryParams struct {
	model.UserQueryParams
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	InactiveUsers    int64 `json:"inactive_users"`
	PendingUsers     int64 `json:"pending_users"`
	SuspendedUsers   int64 `json:"suspended_users"`
	NewUsersToday    int64 `json:"new_users_today"`
	NewUsersThisWeek int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`
}