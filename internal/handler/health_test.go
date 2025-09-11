package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-web-starter/internal/config"
	"go-web-starter/internal/handler/response"
	"go-web-starter/internal/infrastructure/logger"
)

func setupHealthHandler() *HealthHandler {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
	}

	loggerConfig := &config.LoggerConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}
	log, _ := logger.New(loggerConfig)

	return NewHealthHandler(cfg, log)
}

func setupTestRouter(handler *HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add request ID middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	health := router.Group("/health")
	{
		health.GET("/", handler.Health)
		health.GET("/ready", handler.Readiness)
		health.GET("/live", handler.Liveness)
	}

	return router
}

func TestNewHealthHandler(t *testing.T) {
	handler := setupHealthHandler()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.config)
	assert.NotNil(t, handler.logger)
	assert.NotZero(t, handler.startTime)
}

func TestHealthCheck(t *testing.T) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response response.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Parse the health check data
	healthData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "healthy", healthData["status"])
	assert.Equal(t, "go-web-starter", healthData["service"])
	assert.Equal(t, "1.0.0", healthData["version"])
	assert.Equal(t, "test", healthData["environment"])
	assert.NotEmpty(t, healthData["timestamp"])
	assert.NotEmpty(t, healthData["uptime"])
	assert.NotNil(t, healthData["checks"])

	// Check that database check is included
	checks, ok := healthData["checks"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, checks, "database")

	// Verify database check structure
	dbCheck, ok := checks["database"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "healthy", dbCheck["status"])
	assert.Equal(t, "mysql", dbCheck["type"])
}

func TestReadinessCheck(t *testing.T) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response response.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Parse the readiness check data
	readinessData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "ready", readinessData["status"])
	assert.Equal(t, "go-web-starter", readinessData["service"])
	assert.NotEmpty(t, readinessData["timestamp"])
	assert.NotNil(t, readinessData["checks"])

	// Check that database check is included
	checks, ok := readinessData["checks"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, checks, "database")
}

func TestLivenessCheck(t *testing.T) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/live", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response response.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Parse the liveness check data
	livenessData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "alive", livenessData["status"])
	assert.Equal(t, "go-web-starter", livenessData["service"])
	assert.NotEmpty(t, livenessData["timestamp"])
	assert.NotEmpty(t, livenessData["uptime"])
}

func TestHealthCheckStructures(t *testing.T) {
	// Test HealthCheck structure
	healthCheck := HealthCheck{
		Status:      "healthy",
		Service:     "test-service",
		Version:     "1.0.0",
		Environment: "test",
		Timestamp:   "2023-01-01T00:00:00Z",
		Uptime:      "1h30m",
		Checks:      map[string]interface{}{"db": "healthy"},
	}

	assert.Equal(t, "healthy", healthCheck.Status)
	assert.Equal(t, "test-service", healthCheck.Service)
	assert.NotNil(t, healthCheck.Checks)

	// Test ReadinessCheck structure
	readinessCheck := ReadinessCheck{
		Status:    "ready",
		Service:   "test-service",
		Timestamp: "2023-01-01T00:00:00Z",
		Checks:    map[string]interface{}{"db": "ready"},
	}

	assert.Equal(t, "ready", readinessCheck.Status)
	assert.Equal(t, "test-service", readinessCheck.Service)
	assert.NotNil(t, readinessCheck.Checks)

	// Test LivenessCheck structure
	livenessCheck := LivenessCheck{
		Status:    "alive",
		Service:   "test-service",
		Timestamp: "2023-01-01T00:00:00Z",
		Uptime:    "1h30m",
	}

	assert.Equal(t, "alive", livenessCheck.Status)
	assert.Equal(t, "test-service", livenessCheck.Service)
	assert.Equal(t, "1h30m", livenessCheck.Uptime)
}

func TestCheckDatabase(t *testing.T) {
	handler := setupHealthHandler()

	dbStatus := handler.checkDatabase()

	assert.NotNil(t, dbStatus)

	dbMap, ok := dbStatus.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "healthy", dbMap["status"])
	assert.Equal(t, "mysql", dbMap["type"])
	assert.Contains(t, dbMap, "response_time")
	assert.Contains(t, dbMap, "message")
}

func TestCheckRedis(t *testing.T) {
	handler := setupHealthHandler()

	redisStatus := handler.checkRedis()

	// Should return nil since Redis is not configured in test
	assert.Nil(t, redisStatus)
}

func TestCheckKafka(t *testing.T) {
	handler := setupHealthHandler()

	kafkaStatus := handler.checkKafka()

	// Should return nil since Kafka is not configured in test
	assert.Nil(t, kafkaStatus)
}

func BenchmarkHealthCheck(b *testing.B) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkReadinessCheck(b *testing.B) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/ready", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLivenessCheck(b *testing.B) {
	handler := setupHealthHandler()
	router := setupTestRouter(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/live", nil)
		router.ServeHTTP(w, req)
	}
}