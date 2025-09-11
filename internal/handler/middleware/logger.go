package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-web-starter/internal/infrastructure/logger"
)

// LoggerMiddleware creates a logging middleware for Gin
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log the request
		logger.LogHTTPRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.UserAgent(),
			c.ClientIP(),
			c.Writer.Status(),
			duration,
		)

		// Log any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.WithRequestID(requestID).WithError(err.Err).Error("Request processing error")
			}
		}
	}
}
