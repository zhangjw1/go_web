package database

import (
	"fmt"

	"gorm.io/gorm"

	"go-web-starter/internal/infrastructure/logger"
)

// Migrator handles database migrations
type Migrator struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB, log *logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: log,
	}
}

// AutoMigrate runs auto migration for the given models
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	m.logger.Info("Starting database auto migration")

	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		m.logger.WithField("model", modelName).Info("Migrating model")

		if err := m.db.AutoMigrate(model); err != nil {
			m.logger.WithError(err).WithField("model", modelName).Error("Failed to migrate model")
			return fmt.Errorf("failed to migrate model %s: %w", modelName, err)
		}

		m.logger.WithField("model", modelName).Info("Model migrated successfully")
	}

	m.logger.Info("Database auto migration completed successfully")
	return nil
}

// CreateIndexes creates custom indexes for better performance
func (m *Migrator) CreateIndexes() error {
	m.logger.Info("Creating custom database indexes")

	// Define custom indexes here
	indexes := []struct {
		table   string
		name    string
		columns []string
		unique  bool
	}{
		// Example indexes - will be updated when we add specific models
		// {
		//     table:   "users",
		//     name:    "idx_users_email",
		//     columns: []string{"email"},
		//     unique:  true,
		// },
		// {
		//     table:   "users",
		//     name:    "idx_users_status_created_at",
		//     columns: []string{"status", "created_at"},
		//     unique:  false,
		// },
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
func (m *Migrator) createIndex(table, name string, columns []string, unique bool) error {
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

// joinColumns joins column names with commas
func joinColumns(columns []string) string {
	if len(columns) == 0 {
		return ""
	}

	result := columns[0]
	for i := 1; i < len(columns); i++ {
		result += ", " + columns[i]
	}
	return result
}

// DropIndexes drops custom indexes (useful for rollbacks)
func (m *Migrator) DropIndexes() error {
	m.logger.Info("Dropping custom database indexes")

	// Define indexes to drop (same as CreateIndexes but in reverse order)
	indexes := []struct {
		table string
		name  string
	}{
		// Example indexes - will be updated when we add specific models
		// {
		//     table: "users",
		//     name:  "idx_users_status_created_at",
		// },
		// {
		//     table: "users",
		//     name:  "idx_users_email",
		// },
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
func (m *Migrator) dropIndex(table, name string) error {
	// Check if index exists
	if !m.db.Migrator().HasIndex(table, name) {
		m.logger.WithField("index", name).Debug("Index does not exist, skipping")
		return nil
	}

	return m.db.Migrator().DropIndex(table, name)
}

// Seed runs database seeding (for development/testing)
func (m *Migrator) Seed() error {
	m.logger.Info("Starting database seeding")

	// Implement seeding logic here
	// This is typically used for development and testing environments
	// to populate the database with initial data

	m.logger.Info("Database seeding completed successfully")
	return nil
}

// Reset drops all tables and recreates them (dangerous - use with caution)
func (m *Migrator) Reset(models ...interface{}) error {
	m.logger.Warn("Resetting database - this will drop all tables")

	// Drop all tables
	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		if err := m.db.Migrator().DropTable(model); err != nil {
			m.logger.WithError(err).WithField("model", modelName).Error("Failed to drop table")
			return fmt.Errorf("failed to drop table for model %s: %w", modelName, err)
		}
		m.logger.WithField("model", modelName).Info("Table dropped")
	}

	// Recreate tables
	if err := m.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to recreate tables: %w", err)
	}

	// Recreate indexes
	if err := m.CreateIndexes(); err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	m.logger.Info("Database reset completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status
func (m *Migrator) GetMigrationStatus(models ...interface{}) map[string]interface{} {
	status := make(map[string]interface{})

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
		}

		status[modelName] = tableInfo
	}

	return status
}