package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already present in headers
		requestID := c.GetHeader(RequestIDHeader)

		// Generate a new request ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context
		c.Set(RequestIDKey, requestID)

		// Set request ID in response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestIDMiddlewareWithGenerator creates a request ID middleware with custom ID generator
func RequestIDMiddlewareWithGenerator(generator func() string) gin.HandlerFunc {
	if generator == nil {
		generator = func() string {
			return uuid.New().String()
		}
	}

	return func(c *gin.Context) {
		// Check if request ID is already present in headers
		requestID := c.GetHeader(RequestIDHeader)

		// Generate a new request ID if not present
		if requestID == "" {
			requestID = generator()
		}

		// Set request ID in context
		c.Set(RequestIDKey, requestID)

		// Set request ID in response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}
