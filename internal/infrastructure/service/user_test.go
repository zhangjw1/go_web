package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/domain/service"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*model.User, error) {
	args := m.Called(ctx, id, updates)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params repository.UserQueryParams) ([]*model.User, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Search(ctx context.Context, query string, limit int) ([]*model.User, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) GetStats(ctx context.Context) (*repository.UserStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*repository.UserStats), args.Error(1)
}

// MockCacheManager is a mock implementation of cache.Manager for testing
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) SetUser(ctx context.Context, userID interface{}, user map[string]interface{}) error {
	args := m.Called(ctx, userID, user)
	return args.Error(0)
}

func (m *MockCacheManager) GetUser(ctx context.Context, userID interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheManager) DeleteUser(ctx context.Context, userID interface{}) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCacheManager) SetUserByUsername(ctx context.Context, username string, user map[string]interface{}) error {
	args := m.Called(ctx, username, user)
	return args.Error(0)
}

func (m *MockCacheManager) GetUserByUsername(ctx context.Context, username string) (map[string]interface{}, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockMessagingManager is a mock implementation of messaging.Manager for testing
type MockMessagingManager struct {
	mock.Mock
}

func (m *MockMessagingManager) PublishUserCreated(ctx context.Context, userID int64, userData map[string]interface{}) error {
	args := m.Called(ctx, userID, userData)
	return args.Error(0)
}

func (m *MockMessagingManager) PublishUserUpdated(ctx context.Context, userID int64, changes map[string]interface{}) error {
	args := m.Called(ctx, userID, changes)
	return args.Error(0)
}

func (m *MockMessagingManager) PublishUserDeleted(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockLogger is a mock implementation of logger.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) WithError(err error) *MockLogger {
	return m
}

func (m *MockLogger) WithField(key string, value interface{}) *MockLogger {
	return m
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func TestUserServiceImpl_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		request *service.CreateUserRequest
		setup   func(*MockUserRepository, *MockCacheManager, *MockMessagingManager)
		wantErr bool
		errType error
	}{
		{
			name: "successful user creation",
			request: &service.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			setup: func(repo *MockUserRepository, cache *MockCacheManager, messaging *MockMessagingManager) {
				// Mock username check - not found
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, repository.ErrUserNotFound)
				// Mock email check - not found
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, repository.ErrUserNotFound)
				// Mock user creation
				createdUser := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Test User",
					IsActive: true,
				}
				repo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(createdUser, nil)
				// Mock cache operations
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
				// Mock messaging
				messaging.On("PublishUserCreated", mock.Anything, int64(1), mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user already exists by username",
			request: &service.CreateUserRequest{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(repo *MockUserRepository, cache *MockCacheManager, messaging *MockMessagingManager) {
				existingUser := &model.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				repo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)
			},
			wantErr: true,
		},
		{
			name: "user already exists by email",
			request: &service.CreateUserRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func(repo *MockUserRepository, cache *MockCacheManager, messaging *MockMessagingManager) {
				// Mock username check - not found
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, repository.ErrUserNotFound)
				// Mock email check - found existing user
				existingUser := &model.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid request - empty username",
			request: &service.CreateUserRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			setup:   func(*MockUserRepository, *MockCacheManager, *MockMessagingManager) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheManager{}
			mockMessaging := &MockMessagingManager{}
			mockLogger := &MockLogger{}

			// Setup mocks
			tt.setup(mockRepo, mockCache, mockMessaging)

			// Create service
			userService := &UserServiceImpl{
				userRepo:  mockRepo,
				cache:     mockCache,
				messaging: mockMessaging,
				logger:    mockLogger,
			}

			// Execute test
			result, err := userService.CreateUser(context.Background(), tt.request)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Username, result.Username)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.FullName, result.FullName)
				assert.True(t, result.IsActive)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockMessaging.AssertExpectations(t)
		})
	}
}

func TestUserServiceImpl_GetUserByID(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		setup   func(*MockUserRepository, *MockCacheManager)
		wantErr bool
		errType error
	}{
		{
			name:   "successful get user by ID from database",
			userID: 1,
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Mock cache miss
				cache.On("GetUser", mock.Anything, int64(1)).Return(map[string]interface{}{}, assert.AnError)
				// Mock database success
				user := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					IsActive: true,
				}
				repo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)
				// Mock cache set
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Mock cache miss
				cache.On("GetUser", mock.Anything, int64(999)).Return(map[string]interface{}{}, assert.AnError)
				// Mock database not found
				repo.On("GetByID", mock.Anything, int64(999)).Return(nil, repository.ErrUserNotFound)
			},
			wantErr: true,
			errType: service.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheManager{}
			mockMessaging := &MockMessagingManager{}
			mockLogger := &MockLogger{}

			// Setup mocks
			tt.setup(mockRepo, mockCache)

			// Create service
			userService := &UserServiceImpl{
				userRepo:  mockRepo,
				cache:     mockCache,
				messaging: mockMessaging,
				logger:    mockLogger,
			}

			// Execute test
			result, err := userService.GetUserByID(context.Background(), tt.userID)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserServiceImpl_UpdateUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		request *service.UpdateUserRequest
		setup   func(*MockUserRepository, *MockCacheManager, *MockMessagingManager)
		wantErr bool
	}{
		{
			name:   "successful user update",
			userID: 1,
			request: &service.UpdateUserRequest{
				FullName: stringPtr("Updated Name"),
				Bio:      stringPtr("Updated bio"),
			},
			setup: func(repo *MockUserRepository, cache *MockCacheManager, messaging *MockMessagingManager) {
				existingUser := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Original Name",
					Bio:      "Original bio",
					IsActive: true,
				}
				
				// Mock get existing user (cache miss, then database)
				cache.On("GetUser", mock.Anything, int64(1)).Return(map[string]interface{}{}, assert.AnError)
				repo.On("GetByID", mock.Anything, int64(1)).Return(existingUser, nil)
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
				
				// Mock update
				updatedUser := &model.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Updated Name",
					Bio:      "Updated bio",
					IsActive: true,
				}
				repo.On("Update", mock.Anything, int64(1), mock.Anything).Return(updatedUser, nil)
				
				// Mock cache operations
				cache.On("DeleteUser", mock.Anything, int64(1)).Return(nil)
				
				// Mock messaging
				messaging.On("PublishUserUpdated", mock.Anything, int64(1), mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "invalid request - no fields to update",
			userID: 1,
			request: &service.UpdateUserRequest{},
			setup:   func(*MockUserRepository, *MockCacheManager, *MockMessagingManager) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheManager{}
			mockMessaging := &MockMessagingManager{}
			mockLogger := &MockLogger{}

			// Setup mocks
			tt.setup(mockRepo, mockCache, mockMessaging)

			// Create service
			userService := &UserServiceImpl{
				userRepo:  mockRepo,
				cache:     mockCache,
				messaging: mockMessaging,
				logger:    mockLogger,
			}

			// Execute test
			result, err := userService.UpdateUser(context.Background(), tt.userID, tt.request)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockMessaging.AssertExpectations(t)
		})
	}
}

func TestUserServiceImpl_DeleteUser(t *testing.T) {
	userID := int64(1)
	
	// Create mocks
	mockRepo := &MockUserRepository{}
	mockCache := &MockCacheManager{}
	mockMessaging := &MockMessagingManager{}
	mockLogger := &MockLogger{}

	existingUser := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	// Setup mocks
	// Mock get existing user (cache miss, then database)
	mockCache.On("GetUser", mock.Anything, userID).Return(map[string]interface{}{}, assert.AnError)
	mockRepo.On("GetByID", mock.Anything, userID).Return(existingUser, nil)
	mockCache.On("SetUser", mock.Anything, userID, mock.Anything).Return(nil)
	mockCache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
	
	// Mock delete
	mockRepo.On("Delete", mock.Anything, userID).Return(nil)
	
	// Mock cache clear
	mockCache.On("DeleteUser", mock.Anything, userID).Return(nil)
	
	// Mock messaging
	mockMessaging.On("PublishUserDeleted", mock.Anything, userID).Return(nil)

	// Create service
	userService := &UserServiceImpl{
		userRepo:  mockRepo,
		cache:     mockCache,
		messaging: mockMessaging,
		logger:    mockLogger,
	}

	// Execute test
	err := userService.DeleteUser(context.Background(), userID)

	// Verify results
	assert.NoError(t, err)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockMessaging.AssertExpectations(t)
}

func TestUserServiceImpl_AuthenticateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		setup    func(*MockUserRepository, *MockCacheManager)
		wantErr  bool
		errType  error
	}{
		{
			name:     "successful authentication",
			username: "testuser",
			password: "password123",
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Create user with hashed password
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					IsActive:     true,
				}
				
				// Mock get user by username (cache miss, then database)
				cache.On("GetUserByUsername", mock.Anything, "testuser").Return(map[string]interface{}{}, assert.AnError)
				repo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password123",
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Mock get user by username (cache miss, then database not found)
				cache.On("GetUserByUsername", mock.Anything, "nonexistent").Return(map[string]interface{}{}, assert.AnError)
				repo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, repository.ErrUserNotFound)
			},
			wantErr: true,
			errType: service.ErrInvalidCredentials,
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpassword",
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Create user with different hashed password
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					IsActive:     true,
				}
				
				// Mock get user by username (cache miss, then database)
				cache.On("GetUserByUsername", mock.Anything, "testuser").Return(map[string]interface{}{}, assert.AnError)
				repo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
			},
			wantErr: true,
			errType: service.ErrInvalidCredentials,
		},
		{
			name:     "inactive user",
			username: "testuser",
			password: "password123",
			setup: func(repo *MockUserRepository, cache *MockCacheManager) {
				// Create inactive user with correct password
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &model.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					IsActive:     false, // Inactive user
				}
				
				// Mock get user by username (cache miss, then database)
				cache.On("GetUserByUsername", mock.Anything, "testuser").Return(map[string]interface{}{}, assert.AnError)
				repo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
				cache.On("SetUser", mock.Anything, int64(1), mock.Anything).Return(nil)
				cache.On("SetUserByUsername", mock.Anything, "testuser", mock.Anything).Return(nil)
			},
			wantErr: true,
			errType: service.ErrUserInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheManager{}
			mockMessaging := &MockMessagingManager{}
			mockLogger := &MockLogger{}

			// Setup mocks
			tt.setup(mockRepo, mockCache)

			// Create service
			userService := &UserServiceImpl{
				userRepo:  mockRepo,
				cache:     mockCache,
				messaging: mockMessaging,
				logger:    mockLogger,
			}

			// Execute test
			result, err := userService.AuthenticateUser(context.Background(), tt.username, tt.password)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.username, result.Username)
				assert.True(t, result.IsActive)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}