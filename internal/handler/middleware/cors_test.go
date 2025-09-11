package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name:           "allowed origin",
			origin:         "http://localhost:3000",
			method:         "GET",
			expectedOrigin: "http://localhost:3000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unknown origin",
			origin:         "http://unknown.com",
			method:         "GET",
			expectedOrigin: "http://localhost:3000", // fallback to first allowed
			expectedStatus: http.StatusOK,
		},
		{
			name:           "preflight request",
			origin:         "http://localhost:3000",
			method:         "OPTIONS",
			expectedOrigin: "http://localhost:3000",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORSMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
			assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
		})
	}
}

func TestCORSMiddlewareWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(CORSMiddlewareWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials")) // Should be empty when false
	assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, POST", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()
	
	assert.NotNil(t, config)
	assert.True(t, len(config.AllowedOrigins) > 0)
	assert.True(t, len(config.AllowedMethods) > 0)
	assert.True(t, len(config.AllowedHeaders) > 0)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}