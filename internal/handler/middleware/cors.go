package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", getAllowedOrigin(origin))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getAllowedOrigin returns the allowed origin based on the request origin
func getAllowedOrigin(origin string) string {
	// In production, you should maintain a whitelist of allowed origins
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:8080",
		"https://yourdomain.com",
	}

	// Check if the origin is in the allowed list
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return origin
		}
	}

	// Default to the first allowed origin or "*" for development
	if len(allowedOrigins) > 0 {
		return allowedOrigins[0]
	}
	
	return "*"
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Accept",
			"Origin",
			"Cache-Control",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddlewareWithConfig creates a CORS middleware with custom configuration
func CORSMiddlewareWithConfig(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowedOrigin := "*"
		if len(config.AllowedOrigins) > 0 {
			allowedOrigin = ""
			for _, allowedOrig := range config.AllowedOrigins {
				if origin == allowedOrig {
					allowedOrigin = origin
					break
				}
			}
			// If no match found, use the first allowed origin
			if allowedOrigin == "" {
				allowedOrigin = config.AllowedOrigins[0]
			}
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowedHeaders) > 0 {
			headers := ""
			for i, header := range config.AllowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
		}

		if len(config.AllowedMethods) > 0 {
			methods := ""
			for i, method := range config.AllowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)
		}

		if len(config.ExposedHeaders) > 0 {
			exposedHeaders := ""
			for i, header := range config.ExposedHeaders {
				if i > 0 {
					exposedHeaders += ", "
				}
				exposedHeaders += header
			}
			c.Header("Access-Control-Expose-Headers", exposedHeaders)
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}