package database

import (
	"fmt"

	"gorm.io/gorm"

	"go-web-starter/internal/domain/model"
	"go-web-starter/internal/infrastructure/logger"
)

// MigrationManager handles all database migrations
type MigrationManager struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, log *logger.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: log,
	}
}

// GetAllModels returns all models that need to be migrated
func (m *MigrationManager) GetAllModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.Profile{},
	}
}

// MigrateAll runs migration for all models
func (m *MigrationManager) MigrateAll() error {
	m.logger.Info("Starting database migration for all models")

	models := m.GetAllModels()

	// Run auto migration
	if err := m.db.AutoMigrate(models...); err != nil {
		m.logger.WithError(err).Error("Failed to run auto migration")
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	// Create custom indexes
	if err := m.CreateCustomIndexes(); err != nil {
		m.logger.WithError(err).Error("Failed to create custom indexes")
		return fmt.Errorf("failed to create custom indexes: %w", err)
	}

	m.logger.Info("Database migration completed successfully")
	return nil
}

// CreateCustomIndexes creates custom database indexes for better performance
func (m *MigrationManager) CreateCustomIndexes() error {
	m.logger.Info("Creating custom database indexes")

	indexes := []struct {
		table   string
		name    string
		columns []string
		unique  bool
	}{
		// User indexes
		{
			table:   "users",
			name:    "idx_users_username",
			columns: []string{"username"},
			unique:  true,
		},
		{
			table:   "users",
			name:    "idx_users_email",
			columns: []string{"email"},
			unique:  true,
		},
		{
			table:   "users",
			name:    "idx_users_status",
			columns: []string{"status"},
			unique:  false,
		},
		{
			table:   "users",
			name:    "idx_users_email_verified",
			columns: []string{"email_verified"},
			unique:  false,
		},
		{
			table:   "users",
			name:    "idx_users_status_created_at",
			columns: []string{"status", "created_at"},
			unique:  false,
		},
		{
			table:   "users",
			name:    "idx_users_last_login_at",
			columns: []string{"last_login_at"},
			unique:  false,
		},

		// Profile indexes
		{
			table:   "profiles",
			name:    "idx_profiles_user_id",
			columns: []string{"user_id"},
			unique:  true,
		},
		{
			table:   "profiles",
			name:    "idx_profiles_phone",
			columns: []string{"phone"},
			unique:  false,
		},
		{
			table:   "profiles",
			name:    "idx_profiles_country_city",
			columns: []string{"country", "city"},
			unique:  false,
		},
		{
			table:   "profiles",
			name:    "idx_profiles_company",
			columns: []string{"company"},
			unique:  false,
		},
	}

	for _, idx := range indexes {
		if err := m.createIndex(idx.table, idx.name, idx.columns, idx.unique); err != nil {
			m.logger.WithError(err).WithField("index", idx.name).Error("Failed to create index")
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
		m.logger.WithField("index", idx.name).Info("Index created successfully")
	}

	m.logger.Info("Custom database indexes created successfully")
	return nil
}

// createIndex creates a single index
func (m *MigrationManager) createIndex(table, name string, columns []string, unique bool) error {
	// Check if index already exists
	if m.db.Migrator().HasIndex(table, name) {
		m.logger.WithField("index", name).Debug("Index already exists, skipping")
		return nil
	}

	// Build index creation SQL
	indexType := "INDEX"
	if unique {
		indexType = "UNIQUE INDEX"
	}

	sql := fmt.Sprintf("CREATE %s %s ON %s (%s)",
		indexType,
		name,
		table,
		joinColumns(columns),
	)

	return m.db.Exec(sql).Error
}

// DropCustomIndexes drops all custom indexes
func (m *MigrationManager) DropCustomIndexes() error {
	m.logger.Info("Dropping custom database indexes")

	indexes := []struct {
		table string
		name  string
	}{
		// User indexes (in reverse order)
		{"users", "idx_users_last_login_at"},
		{"users", "idx_users_status_created_at"},
		{"users", "idx_users_email_verified"},
		{"users", "idx_users_status"},
		{"users", "idx_users_email"},
		{"users", "idx_users_username"},

		// Profile indexes (in reverse order)
		{"profiles", "idx_profiles_company"},
		{"profiles", "idx_profiles_country_city"},
		{"profiles", "idx_profiles_phone"},
		{"profiles", "idx_profiles_user_id"},
	}

	for _, idx := range indexes {
		if err := m.dropIndex(idx.table, idx.name); err != nil {
			m.logger.WithError(err).WithField("index", idx.name).Error("Failed to drop index")
			return fmt.Errorf("failed to drop index %s: %w", idx.name, err)
		}
		m.logger.WithField("index", idx.name).Info("Index dropped successfully")
	}

	m.logger.Info("Custom database indexes dropped successfully")
	return nil
}

// dropIndex drops a single index
func (m *MigrationManager) dropIndex(table, name string) error {
	// Check if index exists
	if !m.db.Migrator().HasIndex(table, name) {
		m.logger.WithField("index", name).Debug("Index does not exist, skipping")
		return nil
	}

	return m.db.Migrator().DropIndex(table, name)
}

// SeedData seeds the database with initial data
func (m *MigrationManager) SeedData() error {
	m.logger.Info("Starting database seeding")

	// Check if we already have users (avoid duplicate seeding)
	var userCount int64
	if err := m.db.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if userCount > 0 {
		m.logger.Info("Database already has users, skipping seeding")
		return nil
	}

	// Create admin user
	adminUser := &model.User{
		Username:      "admin",
		Email:         "admin@example.com",
		Password:      "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName:     "Admin",
		LastName:      "User",
		Status:        model.UserStatusActive,
		EmailVerified: true,
	}

	if err := adminUser.Validate(); err != nil {
		return fmt.Errorf("admin user validation failed: %w", err)
	}

	if err := m.db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create admin profile
	adminProfile := &model.Profile{
		UserID:   adminUser.ID,
		Bio:      "System Administrator",
		JobTitle: "Administrator",
		Company:  "Go Web Starter",
	}

	if err := adminProfile.Validate(); err != nil {
		return fmt.Errorf("admin profile validation failed: %w", err)
	}

	if err := m.db.Create(adminProfile).Error; err != nil {
		return fmt.Errorf("failed to create admin profile: %w", err)
	}

	// Create test user
	testUser := &model.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName: "Test",
		LastName:  "User",
		Status:    model.UserStatusActive,
	}

	if err := testUser.Validate(); err != nil {
		return fmt.Errorf("test user validation failed: %w", err)
	}

	if err := m.db.Create(testUser).Error; err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	// Create test profile
	testProfile := &model.Profile{
		UserID:   testUser.ID,
		Bio:      "Test user for development and testing",
		JobTitle: "Developer",
		Company:  "Test Company",
	}

	if err := testProfile.Validate(); err != nil {
		return fmt.Errorf("test profile validation failed: %w", err)
	}

	if err := m.db.Create(testProfile).Error; err != nil {
		return fmt.Errorf("failed to create test profile: %w", err)
	}

	m.logger.WithField("users_created", 2).Info("Database seeding completed successfully")
	return nil
}

// ResetDatabase drops all tables and recreates them (dangerous operation)
func (m *MigrationManager) ResetDatabase() error {
	m.logger.Warn("Resetting database - this will drop all tables and data")

	models := m.GetAllModels()

	// Drop all tables in reverse order (to handle foreign key constraints)
	for i := len(models) - 1; i >= 0; i-- {
		model := models[i]
		modelName := fmt.Sprintf("%T", model)
		
		if err := m.db.Migrator().DropTable(model); err != nil {
			m.logger.WithError(err).WithField("model", modelName).Error("Failed to drop table")
			return fmt.Errorf("failed to drop table for model %s: %w", modelName, err)
		}
		m.logger.WithField("model", modelName).Info("Table dropped")
	}

	// Recreate tables
	if err := m.MigrateAll(); err != nil {
		return fmt.Errorf("failed to recreate tables: %w", err)
	}

	// Seed initial data
	if err := m.SeedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	m.logger.Info("Database reset completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status for all models
func (m *MigrationManager) GetMigrationStatus() map[string]interface{} {
	status := make(map[string]interface{})
	models := m.GetAllModels()

	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		hasTable := m.db.Migrator().HasTable(model)
		
		tableInfo := map[string]interface{}{
			"has_table": hasTable,
		}

		if hasTable {
			// Get column information
			columns, err := m.db.Migrator().ColumnTypes(model)
			if err != nil {
				tableInfo["error"] = err.Error()
			} else {
				columnNames := make([]string, len(columns))
				for i, col := range columns {
					columnNames[i] = col.Name()
				}
				tableInfo["columns"] = columnNames
			}

			// Check indexes
			indexes := []string{
				"idx_users_username", "idx_users_email", "idx_users_status",
				"idx_profiles_user_id", "idx_profiles_phone",
			}
			
			existingIndexes := make([]string, 0)
			for _, indexName := range indexes {
				if m.db.Migrator().HasIndex(model, indexName) {
					existingIndexes = append(existingIndexes, indexName)
				}
			}
			tableInfo["indexes"] = existingIndexes
		}

		status[modelName] = tableInfo
	}

	// Add general database info
	var userCount, profileCount int64
	m.db.Model(&model.User{}).Count(&userCount)
	m.db.Model(&model.Profile{}).Count(&profileCount)

	status["statistics"] = map[string]interface{}{
		"user_count":    userCount,
		"profile_count": profileCount,
	}

	return status
}

// ValidateSchema validates that the database schema matches the models
func (m *MigrationManager) ValidateSchema() error {
	m.logger.Info("Validating database schema")

	models := m.GetAllModels()
	
	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		
		// Check if table exists
		if !m.db.Migrator().HasTable(model) {
			return fmt.Errorf("table for model %s does not exist", modelName)
		}
		
		// Check if all required columns exist
		columns, err := m.db.Migrator().ColumnTypes(model)
		if err != nil {
			return fmt.Errorf("failed to get column types for model %s: %w", modelName, err)
		}
		
		if len(columns) == 0 {
			return fmt.Errorf("no columns found for model %s", modelName)
		}
		
		m.logger.WithField("model", modelName).WithField("columns", len(columns)).Debug("Schema validation passed")
	}

	m.logger.Info("Database schema validation completed successfully")
	return nil
}