package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestBaseModel(t *testing.T) {
	model := &BaseModel{}

	// Test initial state
	assert.Equal(t, uint(0), model.GetID())
	assert.True(t, model.GetCreatedAt().IsZero())
	assert.True(t, model.GetUpdatedAt().IsZero())
	assert.Nil(t, model.GetDeletedAt())
	assert.False(t, model.IsDeleted())
}

func TestBaseModelBeforeCreate(t *testing.T) {
	model := &BaseModel{}
	
	// Mock GORM DB (we don't actually need it for this test)
	var db *gorm.DB
	
	err := model.BeforeCreate(db)
	assert.NoError(t, err)
	
	// Check that timestamps were set
	assert.False(t, model.CreatedAt.IsZero())
	assert.False(t, model.UpdatedAt.IsZero())
	assert.Equal(t, model.CreatedAt, model.UpdatedAt)
}

func TestBaseModelBeforeUpdate(t *testing.T) {
	model := &BaseModel{}
	
	// Set initial timestamps
	initialTime := time.Now().Add(-1 * time.Hour)
	model.CreatedAt = initialTime
	model.UpdatedAt = initialTime
	
	// Mock GORM DB
	var db *gorm.DB
	
	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Millisecond)
	
	err := model.BeforeUpdate(db)
	assert.NoError(t, err)
	
	// Check that only UpdatedAt was changed
	assert.Equal(t, initialTime, model.CreatedAt)
	assert.True(t, model.UpdatedAt.After(initialTime))
}

func TestBaseModelSoftDelete(t *testing.T) {
	model := &BaseModel{}
	
	// Initially not deleted
	assert.False(t, model.IsDeleted())
	assert.Nil(t, model.GetDeletedAt())
	
	// Simulate soft delete
	now := time.Now()
	model.DeletedAt = gorm.DeletedAt{
		Time:  now,
		Valid: true,
	}
	
	// Check soft delete state
	assert.True(t, model.IsDeleted())
	assert.NotNil(t, model.GetDeletedAt())
	assert.Equal(t, now, *model.GetDeletedAt())
}

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		params   PaginationParams
		expected PaginationParams
		offset   int
		limit    int
	}{
		{
			name:   "valid params",
			params: PaginationParams{Page: 2, PageSize: 10},
			expected: PaginationParams{Page: 2, PageSize: 10},
			offset: 10,
			limit:  10,
		},
		{
			name:   "zero values",
			params: PaginationParams{Page: 0, PageSize: 0},
			expected: PaginationParams{Page: 1, PageSize: 10},
			offset: 0,
			limit:  10,
		},
		{
			name:   "negative values",
			params: PaginationParams{Page: -1, PageSize: -5},
			expected: PaginationParams{Page: 1, PageSize: 10},
			offset: 0,
			limit:  10,
		},
		{
			name:   "page size too large",
			params: PaginationParams{Page: 1, PageSize: 200},
			expected: PaginationParams{Page: 1, PageSize: 100},
			offset: 0,
			limit:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.SetDefaults()
			
			assert.Equal(t, tt.expected.Page, params.Page)
			assert.Equal(t, tt.expected.PageSize, params.PageSize)
			assert.Equal(t, tt.offset, params.GetOffset())
			assert.Equal(t, tt.limit, params.GetLimit())
		})
	}
}

func TestPaginationResult(t *testing.T) {
	params := &PaginationParams{Page: 2, PageSize: 10}
	total := int64(45)
	data := []string{"item1", "item2", "item3"}

	result := NewPaginationResult(params, total, data)

	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 10, result.PageSize)
	assert.Equal(t, int64(45), result.Total)
	assert.Equal(t, 5, result.TotalPages) // ceil(45/10) = 5
	assert.True(t, result.HasNext)        // page 2 of 5
	assert.True(t, result.HasPrev)        // page 2 > 1
	assert.Equal(t, data, result.Data)
}

func TestPaginationResultEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		pageSize  int
		total     int64
		hasNext   bool
		hasPrev   bool
		totalPages int
	}{
		{
			name:      "first page",
			page:      1,
			pageSize:  10,
			total:     25,
			hasNext:   true,
			hasPrev:   false,
			totalPages: 3,
		},
		{
			name:      "last page",
			page:      3,
			pageSize:  10,
			total:     25,
			hasNext:   false,
			hasPrev:   true,
			totalPages: 3,
		},
		{
			name:      "single page",
			page:      1,
			pageSize:  10,
			total:     5,
			hasNext:   false,
			hasPrev:   false,
			totalPages: 1,
		},
		{
			name:      "empty result",
			page:      1,
			pageSize:  10,
			total:     0,
			hasNext:   false,
			hasPrev:   false,
			totalPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &PaginationParams{Page: tt.page, PageSize: tt.pageSize}
			result := NewPaginationResult(params, tt.total, nil)

			assert.Equal(t, tt.hasNext, result.HasNext)
			assert.Equal(t, tt.hasPrev, result.HasPrev)
			assert.Equal(t, tt.totalPages, result.TotalPages)
		})
	}
}

func TestSortParams(t *testing.T) {
	tests := []struct {
		name     string
		params   SortParams
		expected string
	}{
		{
			name:     "default sort",
			params:   SortParams{},
			expected: "id desc",
		},
		{
			name:     "custom sort asc",
			params:   SortParams{SortBy: "name", SortOrder: "asc"},
			expected: "name asc",
		},
		{
			name:     "custom sort desc",
			params:   SortParams{SortBy: "created_at", SortOrder: "desc"},
			expected: "created_at desc",
		},
		{
			name:     "sort by only",
			params:   SortParams{SortBy: "email"},
			expected: "email asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.SetDefaults()
			orderBy := params.GetOrderBy()
			assert.Equal(t, tt.expected, orderBy)
		})
	}
}

func TestTimeRange(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name  string
		tr    TimeRange
		valid bool
	}{
		{
			name:  "valid range",
			tr:    TimeRange{From: &past, To: &future},
			valid: true,
		},
		{
			name:  "same time",
			tr:    TimeRange{From: &now, To: &now},
			valid: true,
		},
		{
			name:  "invalid range",
			tr:    TimeRange{From: &future, To: &past},
			valid: false,
		},
		{
			name:  "from only",
			tr:    TimeRange{From: &past, To: nil},
			valid: true,
		},
		{
			name:  "to only",
			tr:    TimeRange{From: nil, To: &future},
			valid: true,
		},
		{
			name:  "empty range",
			tr:    TimeRange{From: nil, To: nil},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.tr.IsValid())
		})
	}
}

func TestQueryParams(t *testing.T) {
	params := QueryParams{
		PaginationParams: PaginationParams{Page: 0, PageSize: 0},
		SortParams:       SortParams{SortBy: "", SortOrder: ""},
		FilterParams:     FilterParams{Search: "test"},
	}

	params.SetDefaults()

	// Check that defaults were set
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 10, params.PageSize)
	assert.Equal(t, "id", params.SortBy)
	assert.Equal(t, "desc", params.SortOrder)
	assert.Equal(t, "test", params.Search) // Should remain unchanged
}