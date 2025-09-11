package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-web-starter/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.LoggerConfig
		wantErr bool
	}{
		{
			name: "valid json stdout config",
			config: &config.LoggerConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "valid text stdout config",
			config: &config.LoggerConfig{
				Level:  "debug",
				Format: "text",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "valid file config",
			config: &config.LoggerConfig{
				Level:      "warn",
				Format:     "json",
				Output:     "file",
				Filename:   "test_logs/test.log",
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   true,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &config.LoggerConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: &config.LoggerConfig{
				Level:  "info",
				Format: "invalid",
				Output: "stdout",
			},
			wantErr: true,
		},
		{
			name: "invalid output",
			config: &config.LoggerConfig{
				Level:  "info",
				Format: "json",
				Output: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				
				// Clean up test log files
				if tt.config.Output == "file" {
					os.RemoveAll(filepath.Dir(tt.config.Filename))
				}
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a logger that outputs to a buffer for testing
	var buf bytes.Buffer
	
	config := &config.LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := New(config)
	require.NoError(t, err)
	
	// Redirect output to buffer
	logger.SetOutput(&buf)

	t.Run("WithRequestID", func(t *testing.T) {
		buf.Reset()
		logger.WithRequestID("test-request-123").Info("test message")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "test-request-123", logEntry["request_id"])
		assert.Equal(t, "test message", logEntry["msg"])
	})

	t.Run("WithUserID", func(t *testing.T) {
		buf.Reset()
		logger.WithUserID("user-456").Info("user action")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "user-456", logEntry["user_id"])
	})

	t.Run("WithDuration", func(t *testing.T) {
		buf.Reset()
		duration := "150ms"
		logger.WithDuration(duration).Info("operation completed")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, duration, logEntry["duration"])
	})

	t.Run("LogHTTPRequest", func(t *testing.T) {
		buf.Reset()
		logger.LogHTTPRequest("GET", "/api/users", "test-agent", "127.0.0.1", 200, "150ms")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/api/users", logEntry["path"])
		assert.Equal(t, float64(200), logEntry["status_code"])
		assert.Equal(t, "http_request", logEntry["type"])
	})

	t.Run("LogDatabaseQuery", func(t *testing.T) {
		buf.Reset()
		logger.LogDatabaseQuery("SELECT * FROM users", "50ms", 10)
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "SELECT * FROM users", logEntry["query"])
		assert.Equal(t, "50ms", logEntry["duration"])
		assert.Equal(t, float64(10), logEntry["rows_affected"])
		assert.Equal(t, "database_query", logEntry["type"])
	})

	t.Run("LogCacheOperation", func(t *testing.T) {
		buf.Reset()
		logger.LogCacheOperation("GET", "user:123", true, "5ms")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "GET", logEntry["operation"])
		assert.Equal(t, "user:123", logEntry["key"])
		assert.Equal(t, true, logEntry["hit"])
		assert.Equal(t, "cache_operation", logEntry["type"])
	})

	t.Run("LogKafkaMessage", func(t *testing.T) {
		buf.Reset()
		logger.LogKafkaMessage("user-events", "produce", 1024, "10ms")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "user-events", logEntry["topic"])
		assert.Equal(t, "produce", logEntry["operation"])
		assert.Equal(t, float64(1024), logEntry["message_size"])
		assert.Equal(t, "kafka_message", logEntry["type"])
	})

	t.Run("LogBusinessEvent", func(t *testing.T) {
		buf.Reset()
		details := map[string]interface{}{
			"email": "test@example.com",
			"role":  "admin",
		}
		logger.LogBusinessEvent("user_created", "user", 123, details)
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "user_created", logEntry["event"])
		assert.Equal(t, "user", logEntry["entity"])
		assert.Equal(t, float64(123), logEntry["entity_id"])
		assert.Equal(t, "test@example.com", logEntry["email"])
		assert.Equal(t, "admin", logEntry["role"])
		assert.Equal(t, "business_event", logEntry["type"])
	})
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	
	config := &config.LoggerConfig{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := New(config)
	require.NoError(t, err)
	logger.SetOutput(&buf)

	// Debug and Info should not be logged when level is warn
	logger.Debug("debug message")
	logger.Info("info message")
	assert.Empty(t, buf.String())

	// Warn and Error should be logged
	logger.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")
	
	buf.Reset()
	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestTextFormatter(t *testing.T) {
	var buf bytes.Buffer
	
	config := &config.LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	
	logger, err := New(config)
	require.NoError(t, err)
	logger.SetOutput(&buf)

	logger.Info("test message")
	
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "level=info")
	
	// Should not be JSON format
	var jsonTest map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &jsonTest)
	assert.Error(t, err) // Should fail to parse as JSON
}