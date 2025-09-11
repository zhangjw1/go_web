package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-web-starter/internal/config"
	"go-web-starter/internal/handler"
	"go-web-starter/internal/handler/middleware"
	"go-web-starter/internal/infrastructure/logger"
)

// RouteManager manages all application routes (lightweight)
type RouteManager struct {
	config *config.Config
	logger *logger.Logger

	// Handlers
	healthHandler *handler.HealthHandler
	userHandler   *handler.UserHandler
}

// NewRouteManager creates a new route manager with all dependencies
func NewRouteManager(
	cfg *config.Config,
	log *logger.Logger,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
) *RouteManager {
	rm := &RouteManager{
		config:        cfg,
		logger:        log,
		healthHandler: healthHandler,
		userHandler:   userHandler,
	}

	// Initialize handlers if not provided
	rm.initHandlers()

	return rm
}

// initHandlers initializes all handlers
func (rm *RouteManager) initHandlers() {
	if rm.healthHandler == nil {
		rm.healthHandler = handler.NewHealthHandler(rm.config, rm.logger, nil, nil, nil)
	}
}

// SetUserHandler sets the user handler (for dependency injection)
func (rm *RouteManager) SetUserHandler(userHandler *handler.UserHandler) {
	rm.userHandler = userHandler
}

// RegisterRoutes registers all application routes
func (rm *RouteManager) RegisterRoutes(router *gin.Engine) {
	// Setup global middleware
	rm.setupMiddleware(router)

	// Health check routes (no middleware needed)
	rm.registerHealthRoutes(router)

	// API routes
	rm.registerAPIRoutes(router)

	// Documentation routes
	rm.registerDocumentationRoutes(router)

	// Static file routes (if needed)
	rm.registerStaticRoutes(router)

	rm.logger.Info("All routes registered successfully")
}

// setupMiddleware sets up global middleware
func (rm *RouteManager) setupMiddleware(router *gin.Engine) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Logger middleware
	router.Use(middleware.LoggerMiddleware(rm.logger))

	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Request ID middleware
	router.Use(middleware.RequestIDMiddleware())

	// Response time middleware
	router.Use(middleware.ResponseTimeMiddleware())

	// Security headers middleware
	router.Use(middleware.SecurityHeadersMiddleware())

	rm.logger.Info("Global middleware configured")
}

// registerHealthRoutes registers health check routes
func (rm *RouteManager) registerHealthRoutes(router *gin.Engine) {
	// Detailed health endpoints
	health := router.Group("/health")
	{
		health.GET("/", rm.healthHandler.Health)
		health.GET("/ready", rm.healthHandler.Readiness)
		health.GET("/live", rm.healthHandler.Liveness)
	}
	// Backward compatible simple endpoint
	router.GET("/health", rm.healthHandler.Health)
}

// registerAPIRoutes registers API routes
func (rm *RouteManager) registerAPIRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Welcome endpoint
			v1.GET("/", rm.welcome)

			// Server info endpoint
			v1.GET("/info", rm.serverInfo)

			// Time endpoint
			v1.GET("/time", rm.currentTime)

			// Echo endpoint for testing
			v1.POST("/echo", rm.echo)

			// User routes (only if user handler is available)
			if rm.userHandler != nil {
				rm.registerUserRoutes(v1)
			}
		}
	}
}

// registerUserRoutes registers user-related routes
func (rm *RouteManager) registerUserRoutes(v1 *gin.RouterGroup) {
	users := v1.Group("/users")
	{
		users.POST("/", rm.userHandler.CreateUser)
		users.GET("/", rm.userHandler.ListUsers)
		users.GET("/:id", rm.userHandler.GetUser)
		users.PUT("/:id", rm.userHandler.UpdateUser)
		users.DELETE("/:id", rm.userHandler.DeleteUser)
		users.GET("/username/:username", rm.userHandler.GetUserByUsername)
		users.PUT("/:id/password", rm.userHandler.ChangePassword)
	}
}

// registerDocumentationRoutes registers API documentation routes
func (rm *RouteManager) registerDocumentationRoutes(router *gin.Engine) {
	// TODO: Implement Swagger middleware
	// router.Use(middleware.SwaggerRedirectMiddleware())

	// Swagger documentation routes
	docs := router.Group("/swagger")
	{
		// TODO: Apply conditional middleware (disable in production)
		// docs.Use(middleware.ConditionalSwaggerMiddleware(rm.config, rm.logger))

		// TODO: Swagger UI
		// docs.GET("/*any", middleware.SwaggerMiddleware(rm.config, rm.logger))
		docs.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Swagger UI - Coming Soon"})
		})
	}

	// API documentation redirect
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API Documentation - Coming Soon"})
	})

	rm.logger.Info("Documentation routes registered")
}

// registerStaticRoutes registers static file routes
func (rm *RouteManager) registerStaticRoutes(router *gin.Engine) {
	// Serve static files (uncomment if needed)
	// router.Static("/static", "./web/static")
	// router.StaticFile("/favicon.ico", "./web/static/favicon.ico")

	// Serve uploaded files (uncomment if needed)
	// router.Static("/uploads", "./uploads")
}

// Health check handlers delegated to rm.healthHandler

// API handlers

// welcome returns a welcome message
func (rm *RouteManager) welcome(c *gin.Context) {
	response := gin.H{
		"message":     "Welcome to Go Web Starter!",
		"service":     "go-web-starter",
		"version":     "1.0.0",
		"environment": rm.config.Server.Mode,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"request_id":  c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// serverInfo returns server information
func (rm *RouteManager) serverInfo(c *gin.Context) {
	response := gin.H{
		"service":     "go-web-starter",
		"version":     "1.0.0",
		"environment": rm.config.Server.Mode,
		"server": gin.H{
			"port":          rm.config.Server.Port,
			"read_timeout":  rm.config.Server.ReadTimeout,
			"write_timeout": rm.config.Server.WriteTimeout,
		},
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"request_id": c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// currentTime returns the current server time
func (rm *RouteManager) currentTime(c *gin.Context) {
	now := time.Now()
	response := gin.H{
		"timestamp":  now.UTC().Format(time.RFC3339),
		"unix":       now.Unix(),
		"timezone":   now.Location().String(),
		"request_id": c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// echo returns the request body as response (for testing)
func (rm *RouteManager) echo(c *gin.Context) {
	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid JSON",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
		})
		return
	}

	response := gin.H{
		"echo":       body,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"headers":    c.Request.Header,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"request_id": c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}
