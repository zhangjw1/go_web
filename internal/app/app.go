package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"go-web-starter/internal/config"
	"go-web-starter/internal/handler"
	"go-web-starter/internal/handler/routes"
	"go-web-starter/internal/handler/validator"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
	"go-web-starter/internal/infrastructure/repository"
)

// App represents the main application
type App struct {
	config *config.Config
	logger *logger.Logger
	server *http.Server
	router *gin.Engine

	// Infrastructure components
	db     *database.Database
	cache  cache.CacheService
	msgSvc messaging.MessagingService

	// Route manager
	routeManager *routes.RouteManager
}

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
	}

	if err := app.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	return app, nil
}

// initialize initializes all application components
func (a *App) initialize() error {
	// Initialize logger first
	if err := a.initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	a.logger.Info("Initializing Go Web Starter application...")

	// Initialize infrastructure components
	if err := a.initInfrastructure(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize HTTP server
	if err := a.initHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	a.logger.Info("Application initialized successfully")
	return nil
}

// initLogger initializes the logger
func (a *App) initLogger() error {
	appLogger, err := logger.New(&a.config.Logger)
	if err != nil {
		return err
	}

	a.logger = appLogger
	a.logger.Info("Logger initialized")
	return nil
}

// initInfrastructure initializes infrastructure components
func (a *App) initInfrastructure() error {
	// Initialize database
	if a.config.Database.Enabled {
		if err := a.initDatabase(); err != nil {
			return err
		}
	}

	// Initialize cache (optional - continue if fails)
	if a.config.Redis.Enabled {
		if err := a.initCache(); err != nil {
			a.logger.WithError(err).Warn("Failed to initialize cache, continuing without cache")
			a.cache = nil
		}
	}

	// Initialize messaging (optional - continue if fails)
	if a.config.Kafka.Enabled {
		if err := a.initMessaging(); err != nil {
			a.logger.WithError(err).Warn("Failed to initialize messaging, continuing without messaging")
			a.msgSvc = nil
		}
	}

	return nil
}

// initDatabase initializes the database connection
func (a *App) initDatabase() error {
	a.logger.Info("Initializing database connection...")

	db, err := database.New(&a.config.Database, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	// Test database connection
	if err := db.Health(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	migrationManager := database.NewMigrationManager(db.DB, a.logger)
	if err := migrationManager.MigrateAll(); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	a.db = db
	a.logger.Info("Database initialized successfully")
	return nil
}

// initCache initializes the cache service
func (a *App) initCache() error {
	a.logger.Info("Initializing cache...")

	redisClient, err := cache.NewRedisClient(&a.config.Redis, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create cache client: %w", err)
	}

	a.cache = cache.NewRedisCacheService(redisClient)

	// Test cache connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.cache.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping cache: %w", err)
	}

	a.logger.Info("Cache initialized successfully")
	return nil
}

// initMessaging initializes the messaging service
func (a *App) initMessaging() error {
	a.logger.Info("Initializing messaging...")

	kafkaClient, err := messaging.NewKafkaClient(&a.config.Kafka, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create messaging client: %w", err)
	}
	a.msgSvc = messaging.NewKafkaMessagingService(kafkaClient)

	a.logger.Info("Messaging initialized successfully")
	return nil
}

// initHTTPServer 初始化HTTP服务器和路由
// 该函数负责设置Gin框架、配置路由、初始化处理器、注册路由，
// 并根据配置创建HTTP服务器实例。
//
// 执行步骤包括：
// 1. 根据配置设置Gin运行模式（release/test/debug）
// 2. 创建新的Gin路由器
// 3. 初始化Gin自定义验证器（注册用户名/密码/手机号等验证规则）
// 4. 创建路由处理器（健康检查和用户处理器）
// 5. 设置路由管理器并注册所有路由
// 6. 创建HTTP服务器实例，配置超时时间和端口
//
// 返回值：
//   - error: 成功时返回nil，失败时返回错误信息
func (a *App) initHTTPServer() error {
	a.logger.Info("正在初始化HTTP服务器...")

	// 根据配置设置Gin模式
	switch a.config.Server.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin路由器
	a.router = gin.New()

	// 初始化Gin绑定的自定义验证器（注册用户名/密码/手机号等验证规则）
	validator.InitGinValidator()

	// 创建路由管理器
	healthHandler := handler.NewHealthHandler(a.config, a.logger, a.db, a.cache, a.msgSvc)
	var userHandler *handler.UserHandler
	// 如果数据库可用，则创建用户处理器
	if a.db != nil {
		userRepo := repository.NewUserRepository(a.db.DB, a.logger)
		userHandler = handler.NewUserHandler(userRepo, a.config, a.logger)
	}
	a.routeManager = routes.NewRouteManager(
		a.config,
		a.logger,
		healthHandler,
		userHandler,
	)

	// 注册所有路由
	a.routeManager.RegisterRoutes(a.router)

	// 创建HTTP服务器
	a.server = &http.Server{
		Addr:         ":" + a.config.Server.Port,                                // 服务器地址和端口
		Handler:      a.router,                                                  // 请求处理器
		ReadTimeout:  time.Duration(a.config.Server.ReadTimeout) * time.Second,  // 读取超时时间
		WriteTimeout: time.Duration(a.config.Server.WriteTimeout) * time.Second, // 写入超时时间
	}

	a.logger.WithField("port", a.config.Server.Port).Info("HTTP服务器初始化完成")
	return nil
}

// Run starts the application
// Run 启动应用程序的主要运行逻辑
// 该函数负责启动消息消费者、发布系统启动事件、启动HTTP服务器，并等待关闭信号
// 返回值: error - 如果在启动或运行过程中发生致命错误则返回错误信息
func (a *App) Run() error {
	a.logger.WithField("port", a.config.Server.Port).WithField("mode", a.config.Server.Mode).Info("Starting Go Web Starter")

	// 启动消息消费者（如果消息功能可用）
	if err := a.startMessageConsumers(); err != nil {
		a.logger.WithError(err).Error("Failed to start message consumers")
		// 即使消息消费者启动失败也继续运行应用程序
	}

	// 发布系统启动事件
	if err := a.publishSystemStartup(); err != nil {
		a.logger.WithError(err).Error("Failed to publish system startup event")
		// 即使事件发布失败也继续运行应用程序
	}

	// 在goroutine中启动HTTP服务器以避免阻塞主线程
	go func() {
		a.logger.WithField("port", a.config.Server.Port).Info("Starting HTTP server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	a.logger.WithField("port", a.config.Server.Port).Info("Application started successfully")

	// 等待中断信号以优雅地关闭应用程序
	return a.waitForShutdown()
}

// startMessageConsumers starts message consumers
func (a *App) startMessageConsumers() error {
	if a.msgSvc == nil {
		a.logger.Info("Messaging not available, skipping message consumers")
		return nil
	}

	ctx := context.Background()

	// Start user event consumer (no-op handler for minimal setup)
	userHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			return fmt.Errorf("failed to parse message: %w", err)
		}
		_ = envelope
		return nil
	})
	if err := a.msgSvc.ConsumeMessages(ctx, []string{messaging.TopicUserEvents}, userHandler); err != nil {
		return fmt.Errorf("failed to start user event consumer: %w", err)
	}

	// Start system event consumer (no-op)
	systemHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			return fmt.Errorf("failed to parse message: %w", err)
		}
		_ = envelope
		return nil
	})
	if err := a.msgSvc.ConsumeMessages(ctx, []string{messaging.TopicSystemEvents}, systemHandler); err != nil {
		return fmt.Errorf("failed to start system event consumer: %w", err)
	}

	a.logger.Info("Message consumers started successfully")
	return nil
}

// publishSystemStartup publishes system startup event
func (a *App) publishSystemStartup() error {
	if a.msgSvc == nil {
		return nil
	}

	env := messaging.NewEnvelopedMessage(messaging.EventSystemStartup, "go-web-starter", map[string]interface{}{
		"version":     "1.0.0",
		"environment": a.config.Server.Mode,
		"port":        a.config.Server.Port,
		"timestamp":   time.Now().UTC(),
	})
	return a.msgSvc.SendJSONMessage(messaging.TopicSystemEvents, "system", env)
}

// waitForShutdown waits for shutdown signal and performs graceful shutdown
func (a *App) waitForShutdown() error {
	// Create channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-quit
	a.logger.WithField("signal", sig.String()).Info("Shutdown signal received")

	// Perform graceful shutdown
	return a.shutdown()
}

// shutdown performs graceful shutdown of all components
func (a *App) shutdown() error {
	a.logger.Info("Starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Publish system shutdown event
	if a.msgSvc != nil {
		_ = a.msgSvc.SendJSONMessage(messaging.TopicSystemEvents, "system", messaging.NewEnvelopedMessage(messaging.EventSystemShutdown, "go-web-starter", nil))
	}

	// Shutdown HTTP server
	if a.server != nil {
		a.logger.Info("Shutting down HTTP server...")
		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.WithError(err).Error("Failed to shutdown HTTP server gracefully")
		} else {
			a.logger.Info("HTTP server shutdown completed")
		}
	}

	// Close messaging (Kafka client inside service)
	if a.msgSvc != nil {
		if err := a.msgSvc.Close(); err != nil {
			a.logger.WithError(err).Error("Failed to close messaging service")
		} else {
			a.logger.Info("Messaging service closed")
		}
	}

	// Close cache
	if a.cache != nil {
		if err := a.cache.Close(); err != nil {
			a.logger.WithError(err).Error("Failed to close cache")
		} else {
			a.logger.Info("Cache closed")
		}
	}

	// Close database
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.WithError(err).Error("Failed to close database")
		} else {
			a.logger.Info("Database closed")
		}
	}

	a.logger.Info("Graceful shutdown completed")
	return nil
}

// GetConfig returns the application configuration
func (a *App) GetConfig() *config.Config { return a.config }

// GetLogger returns the application logger
func (a *App) GetLogger() *logger.Logger { return a.logger }

// GetRouter returns the HTTP router (useful for testing)
func (a *App) GetRouter() *gin.Engine { return a.router }

// HealthCheck performs a comprehensive health check
func (a *App) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "go-web-starter",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"checks":    make(map[string]interface{}),
	}

	checks := health["checks"].(map[string]interface{})

	// Check database
	if a.db != nil {
		if err := a.db.Health(); err != nil {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "unhealthy"
		} else {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
				"type":   "mysql",
			}
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// Check cache
	if a.cache != nil {
		if err := a.cache.Ping(ctx); err != nil {
			checks["cache"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "unhealthy"
		} else {
			checks["cache"] = map[string]interface{}{
				"status": "healthy",
				"type":   "redis",
			}
		}
	} else {
		checks["cache"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// Check messaging
	if a.msgSvc != nil {
		metrics := a.msgSvc.GetMetrics()
		checks["messaging"] = map[string]interface{}{
			"status":  "healthy",
			"type":    "kafka",
			"metrics": metrics,
		}
	} else {
		checks["messaging"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	return health
}
