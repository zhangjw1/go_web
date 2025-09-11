package repository

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"

	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/domain/repository"
	"go-web-starter/internal/infrastructure/logger"
)

// BaseRepository provides common repository functionality
type BaseRepository[T model.Model] struct {
	db     *gorm.DB
	logger *logger.Logger
	entity T
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T model.Model](db *gorm.DB, log *logger.Logger, entity T) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:     db,
		logger: log,
		entity: entity,
	}
}

// getEntityName returns the entity name for logging
func (r *BaseRepository[T]) getEntityName() string {
	return reflect.TypeOf(r.entity).Elem().Name()
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity T) error {
	entityName := r.getEntityName()
	
	// Validate entity if it implements Validator interface
	if validator, ok := any(entity).(model.Validator); ok {
		if err := validator.Validate(); err != nil {
			r.logger.WithError(err).WithField("entity", entityName).Error("Entity validation failed")
			return repository.NewRepositoryError("create", entityName, 0, "validation failed", err)
		}
	}

	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("entity", entityName).Error("Failed to create entity")
		
		// Check for duplicate key error
		if isDuplicateKeyError(result.Error) {
			return repository.NewRepositoryError("create", entityName, 0, "entity already exists", result.Error)
		}
		
		return repository.NewRepositoryError("create", entityName, 0, "database error", result.Error)
	}

	r.logger.WithField("entity", entityName).WithField("id", entity.GetID()).Info("Entity created successfully")
	return nil
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (T, error) {
	var zero T
	entityName := r.getEntityName()

	if id == 0 {
		return zero, repository.NewRepositoryError("get", entityName, id, "invalid ID", nil)
	}

	entity := r.createNewEntity()
	result := r.db.WithContext(ctx).First(entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithField("entity", entityName).WithField("id", id).Debug("Entity not found")
			return zero, repository.NewRepositoryError("get", entityName, id, "entity not found", result.Error)
		}
		
		r.logger.WithError(result.Error).WithField("entity", entityName).WithField("id", id).Error("Failed to get entity")
		return zero, repository.NewRepositoryError("get", entityName, id, "database error", result.Error)
	}

	return entity, nil
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity T) error {
	entityName := r.getEntityName()
	id := entity.GetID()

	if id == 0 {
		return repository.NewRepositoryError("update", entityName, id, "invalid ID", nil)
	}

	// Validate entity if it implements Validator interface
	if validator, ok := any(entity).(model.Validator); ok {
		if err := validator.Validate(); err != nil {
			r.logger.WithError(err).WithField("entity", entityName).WithField("id", id).Error("Entity validation failed")
			return repository.NewRepositoryError("update", entityName, id, "validation failed", err)
		}
	}

	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("entity", entityName).WithField("id", id).Error("Failed to update entity")
		return repository.NewRepositoryError("update", entityName, id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("entity", entityName).WithField("id", id).Debug("Entity not found for update")
		return repository.NewRepositoryError("update", entityName, id, "entity not found", nil)
	}

	r.logger.WithField("entity", entityName).WithField("id", id).Info("Entity updated successfully")
	return nil
}

// Delete soft deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	entityName := r.getEntityName()

	if id == 0 {
		return repository.NewRepositoryError("delete", entityName, id, "invalid ID", nil)
	}

	entity := r.createNewEntity()
	result := r.db.WithContext(ctx).Delete(entity, id)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("entity", entityName).WithField("id", id).Error("Failed to delete entity")
		return repository.NewRepositoryError("delete", entityName, id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("entity", entityName).WithField("id", id).Debug("Entity not found for deletion")
		return repository.NewRepositoryError("delete", entityName, id, "entity not found", nil)
	}

	r.logger.WithField("entity", entityName).WithField("id", id).Info("Entity deleted successfully")
	return nil
}

// HardDelete permanently deletes an entity by ID
func (r *BaseRepository[T]) HardDelete(ctx context.Context, id uint) error {
	entityName := r.getEntityName()

	if id == 0 {
		return repository.NewRepositoryError("hard_delete", entityName, id, "invalid ID", nil)
	}

	entity := r.createNewEntity()
	result := r.db.WithContext(ctx).Unscoped().Delete(entity, id)
	if result.Error != nil {
		r.logger.WithError(result.Error).WithField("entity", entityName).WithField("id", id).Error("Failed to hard delete entity")
		return repository.NewRepositoryError("hard_delete", entityName, id, "database error", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.WithField("entity", entityName).WithField("id", id).Debug("Entity not found for hard deletion")
		return repository.NewRepositoryError("hard_delete", entityName, id, "entity not found", nil)
	}

	r.logger.WithField("entity", entityName).WithField("id", id).Info("Entity hard deleted successfully")
	return nil
}

// List retrieves entities with pagination and filtering
func (r *BaseRepository[T]) List(ctx context.Context, params *model.QueryParams) (*model.PaginationResult, error) {
	entityName := r.getEntityName()

	if params == nil {
		params = &model.QueryParams{}
	}
	params.SetDefaults()

	// Create a slice to hold the results
	entities := r.createEntitySlice()
	
	// Build query
	query := r.db.WithContext(ctx).Model(r.entity)
	
	// Apply search filter if provided
	if params.Search != "" {
		// This is a basic implementation - specific repositories should override this
		query = query.Where("id LIKE ?", "%"+params.Search+"%")
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.WithError(err).WithField("entity", entityName).Error("Failed to count entities")
		return nil, repository.NewRepositoryError("list", entityName, 0, "database error", err)
	}

	// Apply pagination and sorting
	query = query.Order(params.SortParams.GetOrderBy()).
		Offset(params.GetOffset()).
		Limit(params.GetLimit())

	// Execute query
	if err := query.Find(entities).Error; err != nil {
		r.logger.WithError(err).WithField("entity", entityName).Error("Failed to list entities")
		return nil, repository.NewRepositoryError("list", entityName, 0, "database error", err)
	}

	result := model.NewPaginationResult(&params.PaginationParams, total, entities)
	r.logger.WithField("entity", entityName).WithField("total", total).WithField("page", params.Page).Debug("Entities listed successfully")
	
	return result, nil
}

// Count returns the total count of entities matching the filter
func (r *BaseRepository[T]) Count(ctx context.Context, filter interface{}) (int64, error) {
	entityName := r.getEntityName()

	query := r.db.WithContext(ctx).Model(r.entity)
	
	// Apply filter if provided
	if filter != nil {
		// This is a basic implementation - specific repositories should override this
		query = query.Where(filter)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.WithError(err).WithField("entity", entityName).Error("Failed to count entities")
		return 0, repository.NewRepositoryError("count", entityName, 0, "database error", err)
	}

	return count, nil
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id uint) (bool, error) {
	entityName := r.getEntityName()

	if id == 0 {
		return false, repository.NewRepositoryError("exists", entityName, id, "invalid ID", nil)
	}

	var count int64
	err := r.db.WithContext(ctx).Model(r.entity).Where("id = ?", id).Count(&count).Error
	if err != nil {
		r.logger.WithError(err).WithField("entity", entityName).WithField("id", id).Error("Failed to check entity existence")
		return false, repository.NewRepositoryError("exists", entityName, id, "database error", err)
	}

	return count > 0, nil
}

// createNewEntity creates a new instance of the entity type
func (r *BaseRepository[T]) createNewEntity() T {
	entityType := reflect.TypeOf(r.entity).Elem()
	newEntity := reflect.New(entityType).Interface()
	return newEntity.(T)
}

// createEntitySlice creates a slice of the entity type
func (r *BaseRepository[T]) createEntitySlice() interface{} {
	entityType := reflect.TypeOf(r.entity).Elem()
	sliceType := reflect.SliceOf(reflect.PtrTo(entityType))
	return reflect.New(sliceType).Interface()
}

// isDuplicateKeyError checks if the error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// MySQL duplicate entry error
	return contains(errStr, "Duplicate entry") || 
		   contains(errStr, "duplicate key") ||
		   contains(errStr, "UNIQUE constraint failed")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

// containsSubstring checks if string contains substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WithTransaction returns a new repository instance using the provided transaction
func (r *BaseRepository[T]) WithTransaction(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:     tx,
		logger: r.logger,
		entity: r.entity,
	}
}