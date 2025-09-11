package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: Enable XSS filtering
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: Prevent XSS and data injection attacks
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'none'; frame-src 'none'; worker-src 'none'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; manifest-src 'self'")

		// Strict-Transport-Security: Enforce HTTPS (only add in production with HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Permissions-Policy: Control browser features
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		c.Next()
	}
}

// SecurityConfig represents security headers configuration
type SecurityConfig struct {
	ContentTypeOptions    string
	FrameOptions          string
	XSSProtection         string
	ReferrerPolicy        string
	ContentSecurityPolicy string
	StrictTransportSecurity string
	PermissionsPolicy     string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentTypeOptions:      "nosniff",
		FrameOptions:           "DENY",
		XSSProtection:          "1; mode=block",
		ReferrerPolicy:         "strict-origin-when-cross-origin",
		ContentSecurityPolicy:  "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'none'; frame-src 'none'; worker-src 'none'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; manifest-src 'self'",
		PermissionsPolicy:      "camera=(), microphone=(), geolocation=(), interest-cohort=()",
		HSTSMaxAge:             31536000, // 1 year
		HSTSIncludeSubdomains:  true,
		HSTSPreload:            true,
	}
}

// SecurityHeadersMiddlewareWithConfig creates a security headers middleware with custom configuration
func SecurityHeadersMiddlewareWithConfig(config *SecurityConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(c *gin.Context) {
		// X-Content-Type-Options
		if config.ContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", config.ContentTypeOptions)
		}

		// X-Frame-Options
		if config.FrameOptions != "" {
			c.Header("X-Frame-Options", config.FrameOptions)
		}

		// X-XSS-Protection
		if config.XSSProtection != "" {
			c.Header("X-XSS-Protection", config.XSSProtection)
		}

		// Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", config.ReferrerPolicy)
		}

		// Content-Security-Policy
		if config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Permissions-Policy
		if config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", config.PermissionsPolicy)
		}

		// Strict-Transport-Security (only for HTTPS)
		if c.Request.TLS != nil {
			hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			if config.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}
			if config.HSTSPreload {
				hstsValue += "; preload"
			}
			c.Header("Strict-Transport-Security", hstsValue)
		}

		c.Next()
	}
}