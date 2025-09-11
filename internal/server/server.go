package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"go-web-starter/internal/config"
	"go-web-starter/internal/handler/middleware"
	"go-web-starter/internal/infrastructure/logger"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	router     *gin.Engine
	httpServer *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, log *logger.Logger) *Server {
	// Set Gin mode based on configuration
	switch cfg.Server.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return &Server{
		config:     cfg,
		logger:     log,
		router:     router,
		httpServer: httpServer,
	}
}

// SetupMiddleware sets up all middleware
func (s *Server) SetupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(middleware.LoggerMiddleware(s.logger))

	// CORS middleware
	s.router.Use(middleware.CORSMiddleware())

	// Request ID middleware
	s.router.Use(middleware.RequestIDMiddleware())

	// Response time middleware
	s.router.Use(middleware.ResponseTimeMiddleware())

	// Security headers middleware
	s.router.Use(middleware.SecurityHeadersMiddleware())
}

// SetupRoutes sets up all routes
func (s *Server) SetupRoutes() {
	// Health check route
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", s.healthCheck)
		
		// System info
		v1.GET("/info", s.systemInfo)
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// systemInfo handles system info requests
func (s *Server) systemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":     "go-web-starter",
		"version":     "1.0.0",
		"environment": s.config.Server.Mode,
		"timestamp":   time.Now().UTC(),
		"uptime":      time.Since(startTime).String(),
	})
}

var startTime = time.Now()

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.WithField("port", s.config.Server.Port).WithField("mode", s.config.Server.Mode).Info("Starting HTTP server")

	// Setup middleware and routes
	s.SetupMiddleware()
	s.SetupRoutes()

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	s.logger.WithField("port", s.config.Server.Port).Info("HTTP server started successfully")
	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown HTTP server gracefully")
		return err
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// Run starts the server and waits for shutdown signal
func (s *Server) Run() error {
	// Start the server
	if err := s.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutdown signal received")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	return s.Stop(ctx)
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetHTTPServer returns the HTTP server (useful for testing)
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}