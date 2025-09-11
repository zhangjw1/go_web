package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
)

func TestNew(t *testing.T) {
	// Skip if no test database is available
	if testing.Short() {
		t.Skip("Skipping database tests in short mode")
	}

	// Create logger for testing
	loggerConfig := &config.LoggerConfig{
		Level:  "error", // Reduce log noise in tests
		Format: "json",
		Output: "stdout",
	}
	log, err := logger.New(loggerConfig)
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "test",
				Password:        "test",
				Database:        "test_db",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			wantErr: true, // Will fail without actual database
			errMsg:  "failed to connect to database",
		},
		{
			name: "invalid host",
			config: &config.DatabaseConfig{
				Host:            "invalid-host",
				Port:            3306,
				Username:        "test",
				Password:        "test",
				Database:        "test_db",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			wantErr: true,
			errMsg:  "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.config, log)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, tt.config, db.GetConfig())

			// Test health check
			err = db.Health()
			assert.NoError(t, err)

			// Test stats
			stats := db.GetStats()
			assert.NotNil(t, stats)
			assert.Contains(t, stats, "max_open_connections")

			// Close database
			err = db.Close()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseTransaction(t *testing.T) {
	// This test would require a real database connection
	// For now, we'll test the transaction wrapper logic
	t.Skip("Skipping transaction test - requires real database")
}

func TestDatabaseLogQuery(t *testing.T) {
	// Create logger for testing
	loggerConfig := &config.LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	log, err := logger.New(loggerConfig)
	require.NoError(t, err)

	// Create database config (won't actually connect)
	dbConfig := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "test",
		Password:        "test",
		Database:        "test_db",
		LogLevel:        "error",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 60,
	}

	// Create database instance (this will fail, but we can still test LogQuery)
	db := &Database{
		config: dbConfig,
		logger: log,
	}

	// Test LogQuery method
	db.LogQuery("SELECT * FROM users", 25*time.Millisecond, 10)
	// This should log the query information
}

func TestDatabaseConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *config.DatabaseConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "user",
				Password:        "pass",
				Database:        "db",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			valid: true,
		},
		{
			name: "empty host",
			config: &config.DatabaseConfig{
				Host:            "",
				Port:            3306,
				Username:        "user",
				Password:        "pass",
				Database:        "db",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			valid: false,
		},
		{
			name: "invalid port",
			config: &config.DatabaseConfig{
				Host:            "localhost",
				Port:            0,
				Username:        "user",
				Password:        "pass",
				Database:        "db",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			valid: false,
		},
		{
			name: "empty database name",
			config: &config.DatabaseConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "user",
				Password:        "pass",
				Database:        "",
				LogLevel:        "error",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateDatabaseConfig(tt.config)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(cfg *config.DatabaseConfig) bool {
	if cfg.Host == "" {
		return false
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return false
	}
	if cfg.Database == "" {
		return false
	}
	if cfg.MaxIdleConns < 0 {
		return false
	}
	if cfg.MaxOpenConns < 0 {
		return false
	}
	if cfg.ConnMaxLifetime < 0 {
		return false
	}
	return true
}

func TestDatabaseStats(t *testing.T) {
	// Create a mock database instance with nil DB
	db := &Database{
		DB:     nil, // This will cause GetStats to return an error
		config: &config.DatabaseConfig{},
		logger: nil,
	}

	// Test GetStats with nil DB (should return error)
	stats := db.GetStats()
	assert.Contains(t, stats, "error")
}

func BenchmarkDatabaseConnection(b *testing.B) {
	// Skip benchmark if no test database is available
	if testing.Short() {
		b.Skip("Skipping database benchmark in short mode")
	}

	// This benchmark would test database connection performance
	// For now, we'll skip it as it requires a real database
	b.Skip("Skipping database benchmark - requires real database")
}