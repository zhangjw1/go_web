package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// ResponseTimeHeader is the header name for response time
	ResponseTimeHeader = "X-Response-Time"
	// ResponseTimeKey is the context key for response time
	ResponseTimeKey = "response_time"
)

// ResponseTimeMiddleware adds response time to headers and context
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(start)
		responseTime := duration.Milliseconds()

		// Set response time in context
		c.Set(ResponseTimeKey, responseTime)

		// Set response time in header (in milliseconds)
		c.Header(ResponseTimeHeader, strconv.FormatInt(responseTime, 10)+"ms")
	}
}

// GetResponseTime retrieves the response time from the context
func GetResponseTime(c *gin.Context) int64 {
	if responseTime, exists := c.Get(ResponseTimeKey); exists {
		if rt, ok := responseTime.(int64); ok {
			return rt
		}
	}
	return 0
}

// ResponseTimeMiddlewareWithCallback creates a response time middleware with callback
func ResponseTimeMiddlewareWithCallback(callback func(duration time.Duration, c *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(start)
		responseTime := duration.Milliseconds()

		// Set response time in context
		c.Set(ResponseTimeKey, responseTime)

		// Set response time in header (in milliseconds)
		c.Header(ResponseTimeHeader, strconv.FormatInt(responseTime, 10)+"ms")

		// Call callback if provided
		if callback != nil {
			callback(duration, c)
		}
	}
}
