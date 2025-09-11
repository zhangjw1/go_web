package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
)

func TestLoggerMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a logger that outputs to a buffer
	var buf bytes.Buffer
	loggerConfig := &config.LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}

	testLogger, err := logger.New(loggerConfig)
	require.NoError(t, err)
	testLogger.SetOutput(&buf)

	// Create Gin router with logger middleware
	router := gin.New()
	router.Use(LoggerMiddleware(testLogger))

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "test-agent")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check log output
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput)

	// Parse log entry
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Verify log fields
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status_code"])
	assert.Equal(t, "test-agent", logEntry["user_agent"])
	assert.Equal(t, "http_request", logEntry["type"])
	assert.Contains(t, logEntry, "duration")
	assert.Contains(t, logEntry, "client_ip")
}

func TestLoggerMiddlewareWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	loggerConfig := &config.LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}

	testLogger, err := logger.New(loggerConfig)
	require.NoError(t, err)
	testLogger.SetOutput(&buf)

	router := gin.New()
	router.Use(LoggerMiddleware(testLogger))

	// Add a route that returns an error
	router.GET("/error", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	req, err := http.NewRequest("GET", "/error", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that error was logged
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Request processing error")
	assert.Contains(t, logOutput, "\"level\":\"error\"")
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": capturedRequestID})
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedRequestID)

	// Check that request ID is in response header
	assert.Equal(t, capturedRequestID, w.Header().Get("X-Request-ID"))
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("with request ID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(RequestIDKey, "test-request-id")

		requestID := GetRequestID(c)
		assert.Equal(t, "test-request-id", requestID)
	})

	t.Run("without request ID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		requestID := GetRequestID(c)
		assert.Empty(t, requestID)
	})

	t.Run("with invalid request ID type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(RequestIDKey, 123) // Invalid type

		requestID := GetRequestID(c)
		assert.Empty(t, requestID)
	})
}
