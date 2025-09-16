package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-web-starter/internal/config"
	docs "go-web-starter/internal/docs"
	"go-web-starter/internal/handler"
	"go-web-starter/internal/handler/middleware"
	"go-web-starter/internal/infrastructure/logger"
)

// RouteManager manages all application routes (lightweight)
type RouteManager struct {
	config *config.Config
	logger *logger.Logger

	// Handlers
	healthHandler     *handler.HealthHandler
	userHandler       *handler.UserHandler
	blockchainHandler *handler.BlockChainHandler
}

// NewRouteManager creates a new route manager with all dependencies
func NewRouteManager(
	cfg *config.Config,
	log *logger.Logger,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	blockchainHandler *handler.BlockChainHandler,
) *RouteManager {
	rm := &RouteManager{
		config:            cfg,
		logger:            log,
		healthHandler:     healthHandler,
		userHandler:       userHandler,
		blockchainHandler: blockchainHandler,
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

// RegisterRoutes 注册所有应用程序路由
// 该方法负责配置中间件并注册健康检查、API接口、文档和静态文件等所有路由
//
// 参数:
//   - router: *gin.Engine - Gin路由器实例，用于注册所有路由
func (rm *RouteManager) RegisterRoutes(router *gin.Engine) {
	// 设置全局中间件
	rm.setupMiddleware(router)

	// 注册健康检查路由（不需要中间件）
	rm.registerHealthRoutes(router)

	// 注册API路由
	rm.registerAPIRoutes(router)

	// 注册文档路由
	rm.registerDocumentationRoutes(router)

	// 注册静态文件路由（如需要）
	rm.registerStaticRoutes(router)

	rm.logger.Info("所有路由注册成功")
}

// setupMiddleware 设置全局中间件
// 为Gin路由器配置必要的中间件组件，包括恢复、日志、CORS、请求ID生成、
// 响应时间跟踪和安全头部等核心功能
//
// 参数:
//   - router: *gin.Engine - 需要配置中间件的Gin路由器引擎
func (rm *RouteManager) setupMiddleware(router *gin.Engine) {
	// 恢复中间件 - 捕获并处理panic，防止服务器崩溃
	router.Use(gin.Recovery())

	// 日志中间件 - 记录每个请求的详细信息
	router.Use(middleware.LoggerMiddleware(rm.logger))

	// CORS中间件 - 处理跨域资源共享
	router.Use(middleware.CORSMiddleware())

	// 请求ID中间件 - 为每个请求生成唯一标识符
	router.Use(middleware.RequestIDMiddleware())

	// 响应时间中间件 - 记录和追踪API响应时间
	router.Use(middleware.ResponseTimeMiddleware())

	// 安全头部中间件 - 添加安全相关的HTTP头部
	router.Use(middleware.SecurityHeadersMiddleware())

	rm.logger.Info("全局中间件配置完成")
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
	// 为了同时兼容 /health/ 和 /health 两种写法
	router.GET("/health", rm.healthHandler.Health)
}

// registerAPIRoutes registers API routes
func (rm *RouteManager) registerAPIRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		//版本控制点，便于未来并行提供 /v2、灰度发布、逐步迁移。
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

			// Blockchain routes (only if blockchain handler is available)
			if rm.blockchainHandler != nil {
				rm.registerBlockchainRoutes(v1)
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

// registerBlockchainRoutes registers blockchain-related routes
func (rm *RouteManager) registerBlockchainRoutes(v1 *gin.RouterGroup) {
	blockchain := v1.Group("/blockchain")
	{
		// 获取最新区块号
		blockchain.GET("/block/latest", rm.blockchainHandler.GetLatestBlockNumber)

		// 根据区块号获取区块信息
		blockchain.GET("/block/:number", rm.blockchainHandler.GetBlockByNumber)

		// 根据交易哈希获取交易信息
		blockchain.GET("/transaction/:hash", rm.blockchainHandler.GetTransactionByHash)

		// 获取地址余额
		blockchain.GET("/balance/:address", rm.blockchainHandler.GetBalance)
	}
}

// registerDocumentationRoutes registers API documentation routes
func (rm *RouteManager) registerDocumentationRoutes(router *gin.Engine) {
	// 配置 Swagger 元信息（按需覆盖）
	docs.SwaggerInfo.BasePath = "/"
	// 动态设置 Host 与 Schemes，避免端口或协议不一致导致“Try it out”失败
	docs.SwaggerInfo.Host = "localhost:" + rm.config.Server.Port
	if rm.config.Server.Mode == "release" {
		docs.SwaggerInfo.Schemes = []string{"https", "http"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// Swagger UI 路由（为 swagger 放宽 CSP）
	swaggerGroup := router.Group("/swagger")
	//设置宽松的 CSP 策略, 允许内联样式和脚本,确保Swagger UI 能够正常加载和运行
	swaggerGroup.Use(func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'self';")
		c.Next()
	})
	//将所有对 /swagger/* 路径的 GET 请求都交给 Swagger UI 的静态文件处理器来处理，从而提供完整的 API 文档界面
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// /docs 重定向到 Swagger UI
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	rm.logger.Info("Swagger documentation routes registered")
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
