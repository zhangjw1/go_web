package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-web-starter/internal/config"
	appLogger "go-web-starter/internal/infrastructure/logger"
)

// Database wraps gorm.DB with additional functionality
type Database struct {
	*gorm.DB
	config *config.DatabaseConfig
	logger *appLogger.Logger
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig, log *appLogger.Logger) (*Database, error) {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Configure GORM logger
	var gormLogger logger.Interface
	switch cfg.LogLevel {
	case "silent":
		gormLogger = logger.Default.LogMode(logger.Silent)
	case "error":
		gormLogger = logger.Default.LogMode(logger.Error)
	case "warn":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "info":
		gormLogger = logger.Default.LogMode(logger.Info)
	default:
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Improve performance
		PrepareStmt:            true, // Cache prepared statements
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	}).Info("Database connected successfully")

	return &Database{
		DB:     db,
		config: cfg,
		logger: log,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	d.logger.Info("Database connection closed")
	return nil
}

// Health checks the database connection health
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func (d *Database) GetStats() map[string]interface{} {
	if d.DB == nil {
		return map[string]interface{}{
			"error": "database connection is nil",
		}
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// LogQuery logs database query information
func (d *Database) LogQuery(query string, duration time.Duration, rowsAffected int64) {
	d.logger.LogDatabaseQuery(query, duration.Milliseconds(), rowsAffected)
}

// GetConfig returns the database configuration
func (d *Database) GetConfig() *config.DatabaseConfig {
	return d.config
}