package model

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserStatus represents user status
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents a user in the system
type User struct {
	BaseModel
	Username    string     `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email       string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password    string     `json:"-" gorm:"not null;size:255"` // Hidden in JSON
	FirstName   string     `json:"first_name" gorm:"size:50"`
	LastName    string     `json:"last_name" gorm:"size:50"`
	Status      UserStatus `json:"status" gorm:"default:'pending';size:20"`
	EmailVerified bool     `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int        `json:"login_count" gorm:"default:0"`
	
	// Relationship
	Profile *Profile `json:"profile,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook is called before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Call parent BeforeCreate
	if err := u.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	// Set default status if not provided
	if u.Status == "" {
		u.Status = UserStatusPending
	}

	// Normalize email and username
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))

	return nil
}

// BeforeUpdate hook is called before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Call parent BeforeUpdate
	if err := u.BaseModel.BeforeUpdate(tx); err != nil {
		return err
	}

	// Normalize email and username if they are being updated
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))

	return nil
}

// Validate validates the user model
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is required")
	}

	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(u.Username) > 50 {
		return errors.New("username must be at most 50 characters long")
	}

	// Username should only contain alphanumeric characters and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(u.Username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	if u.Email == "" {
		return errors.New("email is required")
	}

	// Email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}

	if u.Password == "" {
		return errors.New("password is required")
	}

	if len(u.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	// Validate status
	validStatuses := []UserStatus{UserStatusActive, UserStatusInactive, UserStatusPending, UserStatusSuspended}
	isValidStatus := false
	for _, status := range validStatuses {
		if u.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return errors.New("invalid user status")
	}

	return nil
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsPending checks if the user is pending
func (u *User) IsPending() bool {
	return u.Status == UserStatusPending
}

// IsSuspended checks if the user is suspended
func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if fullName == "" {
		return u.Username
	}
	return fullName
}

// MarkEmailAsVerified marks the user's email as verified
func (u *User) MarkEmailAsVerified() {
	u.EmailVerified = true
	now := time.Now()
	u.EmailVerifiedAt = &now
}

// RecordLogin records a user login
func (u *User) RecordLogin() {
	u.LoginCount++
	now := time.Now()
	u.LastLoginAt = &now
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = UserStatusActive
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
}

// Profile represents a user profile with additional information
type Profile struct {
	BaseModel
	UserID      uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	Avatar      string     `json:"avatar" gorm:"size:500"`
	Bio         string     `json:"bio" gorm:"type:text"`
	Phone       string     `json:"phone" gorm:"size:20"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender" gorm:"size:10"`
	Country     string     `json:"country" gorm:"size:100"`
	City        string     `json:"city" gorm:"size:100"`
	Address     string     `json:"address" gorm:"type:text"`
	Website     string     `json:"website" gorm:"size:255"`
	Company     string     `json:"company" gorm:"size:100"`
	JobTitle    string     `json:"job_title" gorm:"size:100"`
	Timezone    string     `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language    string     `json:"language" gorm:"size:10;default:'en'"`
	
	// Relationship
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Profile model
func (Profile) TableName() string {
	return "profiles"
}

// BeforeCreate hook is called before creating a profile
func (p *Profile) BeforeCreate(tx *gorm.DB) error {
	// Call parent BeforeCreate
	if err := p.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	// Set default values
	if p.Timezone == "" {
		p.Timezone = "UTC"
	}
	if p.Language == "" {
		p.Language = "en"
	}

	return nil
}

// Validate validates the profile model
func (p *Profile) Validate() error {
	if p.UserID == 0 {
		return errors.New("user_id is required")
	}

	// Phone validation (if provided)
	if p.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(p.Phone) {
			return errors.New("invalid phone number format")
		}
	}

	// Website validation (if provided)
	if p.Website != "" {
		websiteRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !websiteRegex.MatchString(p.Website) {
			return errors.New("invalid website URL format")
		}
	}

	// Gender validation (if provided)
	if p.Gender != "" {
		validGenders := []string{"male", "female", "other", "prefer_not_to_say"}
		isValidGender := false
		for _, gender := range validGenders {
			if p.Gender == gender {
				isValidGender = true
				break
			}
		}
		if !isValidGender {
			return errors.New("invalid gender value")
		}
	}

	return nil
}

// GetAge calculates and returns the user's age
func (p *Profile) GetAge() int {
	if p.DateOfBirth == nil {
		return 0
	}

	now := time.Now()
	age := now.Year() - p.DateOfBirth.Year()

	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < p.DateOfBirth.YearDay() {
		age--
	}

	return age
}

// GetDisplayName returns a display name for the profile
func (p *Profile) GetDisplayName() string {
	if p.User != nil {
		return p.User.GetFullName()
	}
	return ""
}

// UserWithProfile represents a user with their profile
type UserWithProfile struct {
	User
	Profile *Profile `json:"profile"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	FilterParams
	Username      string     `json:"username" form:"username"`
	Email         string     `json:"email" form:"email"`
	Status        UserStatus `json:"status" form:"status"`
	EmailVerified *bool      `json:"email_verified" form:"email_verified"`
}

// UserSort represents sorting options for user queries
type UserSort struct {
	SortParams
}

// SetDefaults sets default values for user sorting
func (s *UserSort) SetDefaults() {
	if s.SortBy == "" {
		s.SortBy = "created_at"
	}
	if s.SortOrder == "" {
		s.SortOrder = "desc"
	}
}

// UserQueryParams combines pagination, sorting, and filtering for user queries
type UserQueryParams struct {
	PaginationParams
	UserSort
	UserFilter
}

// SetDefaults sets default values for user query parameters
func (q *UserQueryParams) SetDefaults() {
	q.PaginationParams.SetDefaults()
	q.UserSort.SetDefaults()
}