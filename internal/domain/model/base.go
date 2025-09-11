package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate is called before creating a record
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// BeforeUpdate is called before updating a record
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// IsDeleted checks if the record is soft deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt.Valid
}

// GetID returns the ID of the model
func (b *BaseModel) GetID() uint {
	return b.ID
}

// GetCreatedAt returns the creation time
func (b *BaseModel) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns the last update time
func (b *BaseModel) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// GetDeletedAt returns the deletion time if soft deleted
func (b *BaseModel) GetDeletedAt() *time.Time {
	if b.DeletedAt.Valid {
		return &b.DeletedAt.Time
	}
	return nil
}

// TableName interface for models that want to specify custom table names
type TableNamer interface {
	TableName() string
}

// Validator interface for models that want to implement custom validation
type Validator interface {
	Validate() error
}

// BeforeCreateHook interface for models that want to implement before create hooks
type BeforeCreateHook interface {
	BeforeCreate(tx *gorm.DB) error
}

// BeforeUpdateHook interface for models that want to implement before update hooks
type BeforeUpdateHook interface {
	BeforeUpdate(tx *gorm.DB) error
}

// BeforeDeleteHook interface for models that want to implement before delete hooks
type BeforeDeleteHook interface {
	BeforeDelete(tx *gorm.DB) error
}

// AfterCreateHook interface for models that want to implement after create hooks
type AfterCreateHook interface {
	AfterCreate(tx *gorm.DB) error
}

// AfterUpdateHook interface for models that want to implement after update hooks
type AfterUpdateHook interface {
	AfterUpdate(tx *gorm.DB) error
}

// AfterDeleteHook interface for models that want to implement after delete hooks
type AfterDeleteHook interface {
	AfterDelete(tx *gorm.DB) error
}

// Model interface that all domain models should implement
type Model interface {
	GetID() uint
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	IsDeleted() bool
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// GetOffset calculates the offset for database queries
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for database queries
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// SetDefaults sets default values for pagination parameters
func (p *PaginationParams) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// PaginationResult represents paginated query results
type PaginationResult struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
	Data       interface{} `json:"data"`
}

// NewPaginationResult creates a new pagination result
func NewPaginationResult(params *PaginationParams, total int64, data interface{}) *PaginationResult {
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
		Data:       data,
	}
}

// SortParams represents sorting parameters
type SortParams struct {
	SortBy    string `json:"sort_by" form:"sort_by"`
	SortOrder string `json:"sort_order" form:"sort_order" binding:"oneof=asc desc"`
}

// GetOrderBy returns the ORDER BY clause for database queries
func (s *SortParams) GetOrderBy() string {
	if s.SortBy == "" {
		return "id desc" // Default sorting
	}

	order := "asc"
	if s.SortOrder == "desc" {
		order = "desc"
	}

	return s.SortBy + " " + order
}

// SetDefaults sets default values for sort parameters
func (s *SortParams) SetDefaults() {
	if s.SortBy == "" {
		s.SortBy = "id"
		// When using default sort by id, use desc order (newest first)
		if s.SortOrder == "" {
			s.SortOrder = "desc"
		}
	} else {
		// When user specifies sort by, default to asc order
		if s.SortOrder == "" {
			s.SortOrder = "asc"
		}
	}
}

// FilterParams represents common filter parameters
type FilterParams struct {
	Search    string     `json:"search" form:"search"`
	Status    string     `json:"status" form:"status"`
	CreatedAt *TimeRange `json:"created_at" form:"created_at"`
	UpdatedAt *TimeRange `json:"updated_at" form:"updated_at"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	From *time.Time `json:"from" form:"from"`
	To   *time.Time `json:"to" form:"to"`
}

// IsValid checks if the time range is valid
func (tr *TimeRange) IsValid() bool {
	if tr.From != nil && tr.To != nil {
		return tr.From.Before(*tr.To) || tr.From.Equal(*tr.To)
	}
	return true
}

// QueryParams combines pagination, sorting, and filtering parameters
type QueryParams struct {
	PaginationParams
	SortParams
	FilterParams
}

// SetDefaults sets default values for all query parameters
func (q *QueryParams) SetDefaults() {
	q.PaginationParams.SetDefaults()
	q.SortParams.SetDefaults()
}
