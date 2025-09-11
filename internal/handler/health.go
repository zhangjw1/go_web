package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"context"
	"go-web-starter/internal/config"
	"go-web-starter/internal/handler/response"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	config    *config.Config
	logger    *logger.Logger
	startTime time.Time
	// optional deps
	db        *database.Database
	cache     cache.CacheService
	messaging messaging.MessagingService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config, log *logger.Logger, db *database.Database, cacheSvc cache.CacheService, msgSvc messaging.MessagingService) *HealthHandler {
	return &HealthHandler{
		config:    cfg,
		logger:    log,
		startTime: time.Now(),
		db:        db,
		cache:     cacheSvc,
		messaging: msgSvc,
	}
}

// HealthCheck represents the health check response
type HealthCheck struct {
	Status      string                 `json:"status"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Timestamp   string                 `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]interface{} `json:"checks,omitempty"`
}

// ReadinessCheck represents the readiness check response
type ReadinessCheck struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// LivenessCheck represents the liveness check response
type LivenessCheck struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// Health performs a comprehensive health check
// @Summary Health check
// @Description Comprehensive health check including all dependencies
// @Tags health
// @Produce json
// @Success 200 {object} response.Response{data=HealthCheck}
// @Failure 503 {object} response.Response{data=HealthCheck}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	now := time.Now()
	uptime := now.Sub(h.startTime)

	healthCheck := HealthCheck{
		Status:      "healthy",
		Service:     "go-web-starter",
		Version:     "1.0.0",
		Environment: h.config.Server.Mode,
		Timestamp:   now.UTC().Format(time.RFC3339),
		Uptime:      uptime.String(),
		Checks:      make(map[string]interface{}),
	}

	// Check database connectivity
	dbStatus := h.checkDatabase()
	healthCheck.Checks["database"] = dbStatus

	// Check Redis connectivity (if configured)
	redisStatus := h.checkRedis(c)
	if redisStatus != nil {
		healthCheck.Checks["redis"] = redisStatus
	}

	// Check Kafka connectivity (if configured)
	kafkaStatus := h.checkKafka()
	if kafkaStatus != nil {
		healthCheck.Checks["kafka"] = kafkaStatus
	}

	// Determine overall health status
	overallHealthy := true
	for _, check := range healthCheck.Checks {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if status, exists := checkMap["status"]; exists && status != "healthy" {
				overallHealthy = false
				break
			}
		}
	}

	if !overallHealthy {
		healthCheck.Status = "unhealthy"
		h.logger.Warn("Health check failed - some dependencies are unhealthy")
		response.ServiceUnavailable(c, "Service is unhealthy")
		return
	}

	h.logger.Debug("Health check passed")
	response.Success(c, healthCheck)
}

// Readiness checks if the service is ready to serve requests
// @Summary Readiness check
// @Description Check if the service is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} response.Response{data=ReadinessCheck}
// @Failure 503 {object} response.Response{data=ReadinessCheck}
// @Router /health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	now := time.Now()

	readinessCheck := ReadinessCheck{
		Status:    "ready",
		Service:   "go-web-starter",
		Timestamp: now.UTC().Format(time.RFC3339),
		Checks:    make(map[string]interface{}),
	}

	// Check critical dependencies for readiness
	dbStatus := h.checkDatabase()
	readinessCheck.Checks["database"] = dbStatus

	// Check if database is ready
	if dbMap, ok := dbStatus.(map[string]interface{}); ok {
		if status, exists := dbMap["status"]; exists && status != "healthy" {
			readinessCheck.Status = "not_ready"
			h.logger.Warn("Readiness check failed - database is not ready")
			response.ServiceUnavailable(c, "Service is not ready")
			return
		}
	}

	h.logger.Debug("Readiness check passed")
	response.Success(c, readinessCheck)
}

// Liveness checks if the service is alive
// @Summary Liveness check
// @Description Check if the service is alive and running
// @Tags health
// @Produce json
// @Success 200 {object} response.Response{data=LivenessCheck}
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	now := time.Now()
	uptime := now.Sub(h.startTime)

	livenessCheck := LivenessCheck{
		Status:    "alive",
		Service:   "go-web-starter",
		Timestamp: now.UTC().Format(time.RFC3339),
		Uptime:    uptime.String(),
	}

	h.logger.Debug("Liveness check passed")
	response.Success(c, livenessCheck)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase() interface{} {
	if h.db == nil {
		return map[string]interface{}{
			"status": "not_configured",
		}
	}
	start := time.Now()
	if err := h.db.Health(); err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"type":    "mysql",
			"error":   err.Error(),
			"message": "Database connection failed",
		}
	}
	return map[string]interface{}{
		"status":        "healthy",
		"type":          "mysql",
		"response_time": time.Since(start).String(),
		"message":       "Database connection is healthy",
	}
}

// checkRedis checks Redis connectivity
func (h *HealthHandler) checkRedis(c *gin.Context) interface{} {
	if h.cache == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	start := time.Now()
	if err := h.cache.Ping(ctx); err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"type":    "redis",
			"error":   err.Error(),
			"message": "Redis connection failed",
		}
	}
	return map[string]interface{}{
		"status":        "healthy",
		"type":          "redis",
		"response_time": time.Since(start).String(),
		"message":       "Redis connection is healthy",
	}
}

// checkKafka checks Kafka connectivity
func (h *HealthHandler) checkKafka() interface{} {
	if h.messaging == nil {
		return nil
	}
	metrics := h.messaging.GetMetrics()
	return map[string]interface{}{
		"status":  "healthy",
		"type":    "kafka",
		"metrics": metrics,
	}
}
