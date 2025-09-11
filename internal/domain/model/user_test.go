package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserValidation(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Status:   UserStatusActive,
			},
			wantErr: false,
		},
		{
			name: "empty username",
			user: User{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username is required",
		},
		{
			name: "short username",
			user: User{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username must be at least 3 characters long",
		},
		{
			name: "long username",
			user: User{
				Username: "this_is_a_very_long_username_that_exceeds_fifty_characters",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username must be at most 50 characters long",
		},
		{
			name: "invalid username characters",
			user: User{
				Username: "test-user!",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username can only contain letters, numbers, and underscores",
		},
		{
			name: "empty email",
			user: User{
				Username: "testuser",
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			user: User{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "empty password",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "short password",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
			},
			wantErr: true,
			errMsg:  "password must be at least 6 characters long",
		},
		{
			name: "invalid status",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Status:   "invalid_status",
			},
			wantErr: true,
			errMsg:  "invalid user status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserBeforeCreate(t *testing.T) {
	user := &User{
		Username: "TestUser",
		Email:    "TEST@EXAMPLE.COM",
		Password: "password123",
	}

	// Mock GORM DB
	var db *gorm.DB

	err := user.BeforeCreate(db)
	assert.NoError(t, err)

	// Check that email and username are normalized
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, UserStatusPending, user.Status)

	// Check timestamps
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestUserBeforeUpdate(t *testing.T) {
	user := &User{
		Username: "TestUser",
		Email:    "TEST@EXAMPLE.COM",
		Password: "password123",
	}

	// Set initial timestamps
	initialTime := time.Now().Add(-1 * time.Hour)
	user.CreatedAt = initialTime
	user.UpdatedAt = initialTime

	// Mock GORM DB
	var db *gorm.DB

	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Millisecond)

	err := user.BeforeUpdate(db)
	assert.NoError(t, err)

	// Check that email and username are normalized
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)

	// Check that only UpdatedAt was changed
	assert.Equal(t, initialTime, user.CreatedAt)
	assert.True(t, user.UpdatedAt.After(initialTime))
}

func TestUserStatusMethods(t *testing.T) {
	user := &User{}

	// Test IsActive
	user.Status = UserStatusActive
	assert.True(t, user.IsActive())
	assert.False(t, user.IsPending())
	assert.False(t, user.IsSuspended())

	// Test IsPending
	user.Status = UserStatusPending
	assert.False(t, user.IsActive())
	assert.True(t, user.IsPending())
	assert.False(t, user.IsSuspended())

	// Test IsSuspended
	user.Status = UserStatusSuspended
	assert.False(t, user.IsActive())
	assert.False(t, user.IsPending())
	assert.True(t, user.IsSuspended())
}

func TestUserGetFullName(t *testing.T) {
	tests := []struct {
		name      string
		user      User
		expected  string
	}{
		{
			name: "both first and last name",
			user: User{
				Username:  "testuser",
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "John Doe",
		},
		{
			name: "only first name",
			user: User{
				Username:  "testuser",
				FirstName: "John",
				LastName:  "",
			},
			expected: "John",
		},
		{
			name: "only last name",
			user: User{
				Username:  "testuser",
				FirstName: "",
				LastName:  "Doe",
			},
			expected: "Doe",
		},
		{
			name: "no names",
			user: User{
				Username:  "testuser",
				FirstName: "",
				LastName:  "",
			},
			expected: "testuser",
		},
		{
			name: "names with spaces",
			user: User{
				Username:  "testuser",
				FirstName: " John ",
				LastName:  " Doe ",
			},
			expected: "John   Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetFullName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserMarkEmailAsVerified(t *testing.T) {
	user := &User{
		EmailVerified: false,
	}

	assert.False(t, user.EmailVerified)
	assert.Nil(t, user.EmailVerifiedAt)

	user.MarkEmailAsVerified()

	assert.True(t, user.EmailVerified)
	assert.NotNil(t, user.EmailVerifiedAt)
	assert.True(t, time.Since(*user.EmailVerifiedAt) < time.Second)
}

func TestUserRecordLogin(t *testing.T) {
	user := &User{
		LoginCount: 0,
	}

	assert.Equal(t, 0, user.LoginCount)
	assert.Nil(t, user.LastLoginAt)

	user.RecordLogin()

	assert.Equal(t, 1, user.LoginCount)
	assert.NotNil(t, user.LastLoginAt)
	assert.True(t, time.Since(*user.LastLoginAt) < time.Second)

	// Record another login
	time.Sleep(1 * time.Millisecond)
	firstLoginTime := *user.LastLoginAt
	user.RecordLogin()

	assert.Equal(t, 2, user.LoginCount)
	assert.True(t, user.LastLoginAt.After(firstLoginTime))
}

func TestUserStatusChangeMethods(t *testing.T) {
	user := &User{Status: UserStatusPending}

	// Test Activate
	user.Activate()
	assert.Equal(t, UserStatusActive, user.Status)

	// Test Deactivate
	user.Deactivate()
	assert.Equal(t, UserStatusInactive, user.Status)

	// Test Suspend
	user.Suspend()
	assert.Equal(t, UserStatusSuspended, user.Status)
}

func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid profile",
			profile: Profile{
				UserID: 1,
				Phone:  "+1234567890",
				Website: "https://example.com",
				Gender: "male",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			profile: Profile{
				Phone: "+1234567890",
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "invalid phone format",
			profile: Profile{
				UserID: 1,
				Phone:  "invalid-phone",
			},
			wantErr: true,
			errMsg:  "invalid phone number format",
		},
		{
			name: "invalid website format",
			profile: Profile{
				UserID:  1,
				Website: "invalid-website",
			},
			wantErr: true,
			errMsg:  "invalid website URL format",
		},
		{
			name: "invalid gender",
			profile: Profile{
				UserID: 1,
				Gender: "invalid-gender",
			},
			wantErr: true,
			errMsg:  "invalid gender value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileBeforeCreate(t *testing.T) {
	profile := &Profile{
		UserID: 1,
	}

	// Mock GORM DB
	var db *gorm.DB

	err := profile.BeforeCreate(db)
	assert.NoError(t, err)

	// Check default values
	assert.Equal(t, "UTC", profile.Timezone)
	assert.Equal(t, "en", profile.Language)

	// Check timestamps
	assert.False(t, profile.CreatedAt.IsZero())
	assert.False(t, profile.UpdatedAt.IsZero())
}

func TestProfileGetAge(t *testing.T) {
	profile := &Profile{}

	// Test with no date of birth
	assert.Equal(t, 0, profile.GetAge())

	// Test with date of birth
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	profile.DateOfBirth = &birthDate

	age := profile.GetAge()
	expectedAge := time.Now().Year() - 1990

	// Age might be off by 1 depending on current date vs birthday
	assert.True(t, age == expectedAge || age == expectedAge-1)
}

func TestUserSort(t *testing.T) {
	sort := &UserSort{}
	sort.SetDefaults()

	assert.Equal(t, "created_at", sort.SortBy)
	assert.Equal(t, "desc", sort.SortOrder)
}

func TestUserQueryParams(t *testing.T) {
	params := &UserQueryParams{
		PaginationParams: PaginationParams{Page: 0, PageSize: 0},
		UserSort:         UserSort{SortParams: SortParams{SortBy: "", SortOrder: ""}},
		UserFilter:       UserFilter{Username: "test"},
	}

	params.SetDefaults()

	// Check pagination defaults
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 10, params.PageSize)

	// Check sort defaults
	assert.Equal(t, "created_at", params.SortBy)
	assert.Equal(t, "desc", params.SortOrder)

	// Check filter remains unchanged
	assert.Equal(t, "test", params.Username)
}