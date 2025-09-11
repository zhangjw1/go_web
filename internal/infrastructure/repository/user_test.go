package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-web-starter/internal/config"
	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/infrastructure/logger"
)

func TestUserRepository(t *testing.T) {
	// Skip if no test database is available
	if testing.Short() {
		t.Skip("Skipping user repository tests in short mode")
	}

	// Create logger for testing
	loggerConfig := &config.LoggerConfig{
		Level:  "error", // Reduce log noise in tests
		Format: "json",
		Output: "stdout",
	}
	log, err := logger.New(loggerConfig)
	require.NoError(t, err)

	// This test would require a real database connection
	// For now, we'll test the repository creation and interface compliance
	t.Run("NewUserRepository", func(t *testing.T) {
		userRepo := NewUserRepository(nil, log)
		assert.NotNil(t, userRepo)
		
		// Verify it implements the interface
		var _ repository.UserRepository = userRepo
	})
}

func TestUserRepositoryValidation(t *testing.T) {
	// Create logger for testing
	loggerConfig := &config.LoggerConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}
	log, err := logger.New(loggerConfig)
	require.NoError(t, err)

	userRepo := NewUserRepository(nil, log)
	ctx := context.Background()

	t.Run("GetByUsername with empty username", func(t *testing.T) {
		user, err := userRepo.GetByUsername(ctx, "")
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("GetByEmail with empty email", func(t *testing.T) {
		user, err := userRepo.GetByEmail(ctx, "")
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("GetWithProfile with invalid ID", func(t *testing.T) {
		user, err := userRepo.GetWithProfile(ctx, 0)
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})

	t.Run("UpdateLastLogin with invalid ID", func(t *testing.T) {
		err := userRepo.UpdateLastLogin(ctx, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})

	t.Run("UpdateEmailVerification with invalid ID", func(t *testing.T) {
		err := userRepo.UpdateEmailVerification(ctx, 0, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})

	t.Run("UpdateStatus with invalid ID", func(t *testing.T) {
		err := userRepo.UpdateStatus(ctx, 0, model.UserStatusActive)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})
}

func TestUserRepositoryWithDatabase(t *testing.T) {
	// Skip if no test database is available
	if testing.Short() {
		t.Skip("Skipping database user repository tests in short mode")
	}

	// This test would require a real database connection
	// In a real test environment, you would:
	// 1. Set up a test database
	// 2. Create a database connection
	// 3. Test the repository operations
	// 4. Clean up the test database

	t.Skip("Skipping database user repository tests - requires real database")

	// Example of what the test would look like:
	/*
	// Create test database connection
	dbConfig := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "test",
		Password:        "test",
		Database:        "test_user_repo_db",
		LogLevel:        "error",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30,
	}

	log, _ := logger.New(loggerConfig)
	db, err := database.New(dbConfig, log)
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	migrationManager := database.NewMigrationManager(db.DB, log)
	err = migrationManager.MigrateAll()
	require.NoError(t, err)

	userRepo := NewUserRepository(db.DB, log)
	ctx := context.Background()

	t.Run("Create and Get User", func(t *testing.T) {
		user := &model.User{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Status:    model.UserStatusActive,
		}

		// Create user
		err := userRepo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Get user by ID
		foundUser, err := userRepo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, foundUser.Username)
		assert.Equal(t, user.Email, foundUser.Email)

		// Get user by username
		foundUser, err = userRepo.GetByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)

		// Get user by email
		foundUser, err = userRepo.GetByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
	})

	t.Run("Update User", func(t *testing.T) {
		// ... test update operations
	})

	t.Run("Delete User", func(t *testing.T) {
		// ... test delete operations
	})

	t.Run("List Users", func(t *testing.T) {
		// ... test list operations
	})
	*/
}

func TestRepositoryError(t *testing.T) {
	t.Run("NewRepositoryError", func(t *testing.T) {
		err := repository.NewRepositoryError("create", "User", 123, "test error", nil)
		assert.NotNil(t, err)
		assert.Equal(t, "create", err.Op)
		assert.Equal(t, "User", err.Entity)
		assert.Equal(t, uint(123), err.ID)
		assert.Equal(t, "test error", err.Message)
	})

	t.Run("Error message with ID", func(t *testing.T) {
		err := repository.NewRepositoryError("get", "User", 123, "not found", nil)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "get")
		assert.Contains(t, errMsg, "User")
		assert.Contains(t, errMsg, "not found")
	})

	t.Run("Error message without ID", func(t *testing.T) {
		err := repository.NewRepositoryError("create", "User", 0, "validation failed", nil)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "create")
		assert.Contains(t, errMsg, "User")
		assert.Contains(t, errMsg, "validation failed")
	})

	t.Run("IsNotFoundError", func(t *testing.T) {
		notFoundErr := repository.NewRepositoryError("get", "User", 123, "entity not found", nil)
		otherErr := repository.NewRepositoryError("create", "User", 0, "validation failed", nil)

		assert.True(t, repository.IsNotFoundError(notFoundErr))
		assert.False(t, repository.IsNotFoundError(otherErr))
	})

	t.Run("IsAlreadyExistsError", func(t *testing.T) {
		existsErr := repository.NewRepositoryError("create", "User", 0, "entity already exists", nil)
		otherErr := repository.NewRepositoryError("get", "User", 123, "not found", nil)

		assert.True(t, repository.IsAlreadyExistsError(existsErr))
		assert.False(t, repository.IsAlreadyExistsError(otherErr))
	})
}

func BenchmarkUserRepositoryCreation(b *testing.B) {
	loggerConfig := &config.LoggerConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}
	log, _ := logger.New(loggerConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userRepo := NewUserRepository(nil, log)
		_ = userRepo
	}
}