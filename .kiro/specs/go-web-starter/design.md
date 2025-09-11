# 设计文档

## 概述

go-web-starter是一个基于Go语言的现代web应用程序模板，采用分层架构设计，集成了企业级应用所需的核心组件。项目遵循Clean Architecture原则，确保代码的可维护性、可测试性和可扩展性。

## 架构

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   Kafka Client  │    │   Admin Tools   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway (Gin)                        │
├─────────────────────────────────────────────────────────────────┤
│                         Middleware Layer                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Logging   │ │    CORS     │ │ Rate Limit  │ │    Auth     ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                        Handler Layer                            │
├─────────────────────────────────────────────────────────────────┤
│                        Service Layer                            │
├─────────────────────────────────────────────────────────────────┤
│                      Repository Layer                           │
├─────────────────────────────────────────────────────────────────┤
│                       Infrastructure                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   MySQL     │ │    Redis    │ │    Kafka    │ │   Logger    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### 目录结构

```
go-web-starter/
├── cmd/
│   └── server/
│       └── main.go                 # 应用程序入口
├── internal/
│   ├── config/
│   │   └── config.go              # 配置管理
│   ├── handler/
│   │   ├── health.go              # 健康检查处理器
│   │   ├── user.go                # 用户相关处理器
│   │   └── middleware/
│   │       ├── cors.go            # CORS中间件
│   │       ├── logger.go          # 日志中间件
│   │       └── recovery.go        # 恢复中间件
│   ├── service/
│   │   ├── user.go                # 用户业务逻辑
│   │   └── message.go             # 消息处理服务
│   ├── repository/
│   │   ├── user.go                # 用户数据访问
│   │   └── interfaces.go          # 仓储接口定义
│   ├── model/
│   │   ├── user.go                # 用户模型
│   │   └── base.go                # 基础模型
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── mysql.go           # MySQL连接管理
│   │   ├── cache/
│   │   │   └── redis.go           # Redis连接管理
│   │   ├── message/
│   │   │   └── kafka.go           # Kafka连接管理
│   │   └── logger/
│   │       └── logger.go          # 日志配置
│   └── app/
│       └── app.go                 # 应用程序初始化
├── pkg/
│   ├── response/
│   │   └── response.go            # 统一响应格式
│   └── utils/
│       └── validator.go           # 验证工具
├── test/
│   ├── integration/
│   │   └── api_test.go            # API集成测试
│   └── unit/
│       └── service_test.go        # 单元测试
├── configs/
│   ├── config.yaml                # 默认配置
│   ├── config.dev.yaml            # 开发环境配置
│   └── config.prod.yaml           # 生产环境配置
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       └── deployment.yaml
├── scripts/
│   ├── build.sh                   # 构建脚本
│   └── migrate.sh                 # 数据库迁移脚本
├── docs/
│   └── api.md                     # API文档
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 组件和接口

### 1. 配置管理 (Viper)

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Kafka    KafkaConfig    `mapstructure:"kafka"`
    Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
    Port         string `mapstructure:"port"`
    Mode         string `mapstructure:"mode"`
    ReadTimeout  int    `mapstructure:"read_timeout"`
    WriteTimeout int    `mapstructure:"write_timeout"`
}
```

### 2. 数据库层 (GORM + MySQL)

```go
type Repository interface {
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id uint, entity interface{}) error
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id uint, entity interface{}) error
    List(ctx context.Context, entities interface{}, filters map[string]interface{}) error
}

type UserRepository interface {
    Repository
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    GetActiveUsers(ctx context.Context) ([]*model.User, error)
}
```

### 3. 缓存层 (Redis)

```go
type CacheService interface {
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    SetHash(ctx context.Context, key string, fields map[string]interface{}) error
    GetHash(ctx context.Context, key string) (map[string]string, error)
}
```

### 4. 消息队列 (Kafka)

```go
type MessageProducer interface {
    SendMessage(ctx context.Context, topic string, message []byte) error
    SendMessageWithKey(ctx context.Context, topic string, key string, message []byte) error
}

type MessageConsumer interface {
    Subscribe(topics []string, handler MessageHandler) error
    Close() error
}

type MessageHandler func(ctx context.Context, message *Message) error

type Message struct {
    Topic     string
    Key       string
    Value     []byte
    Timestamp time.Time
    Headers   map[string]string
}
```

### 5. 服务层

```go
type UserService interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
    GetUser(ctx context.Context, id uint) (*UserResponse, error)
    UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) (*UserResponse, error)
    DeleteUser(ctx context.Context, id uint) error
    ListUsers(ctx context.Context, filters *UserFilters) ([]*UserResponse, error)
}

type MessageService interface {
    PublishUserEvent(ctx context.Context, event *UserEvent) error
    ProcessUserEvent(ctx context.Context, event *UserEvent) error
}
```

### 6. HTTP处理层 (Gin)

```go
type UserHandler struct {
    userService UserService
    logger      *logrus.Logger
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    // 请求验证、业务逻辑调用、响应处理
}
```

## 数据模型

### 基础模型

```go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 用户模型

```go
type User struct {
    BaseModel
    Email     string `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
    Username  string `gorm:"uniqueIndex;not null" json:"username" validate:"required,min=3,max=50"`
    Password  string `gorm:"not null" json:"-" validate:"required,min=6"`
    FirstName string `gorm:"not null" json:"first_name" validate:"required"`
    LastName  string `gorm:"not null" json:"last_name" validate:"required"`
    Status    string `gorm:"default:active" json:"status"`
    Profile   Profile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
}

type Profile struct {
    BaseModel
    UserID      uint   `gorm:"not null" json:"user_id"`
    Avatar      string `json:"avatar"`
    Bio         string `json:"bio"`
    PhoneNumber string `json:"phone_number"`
}
```

## 错误处理

### 错误类型定义

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

const (
    ErrCodeValidation    = "VALIDATION_ERROR"
    ErrCodeNotFound      = "NOT_FOUND"
    ErrCodeUnauthorized  = "UNAUTHORIZED"
    ErrCodeInternal      = "INTERNAL_ERROR"
    ErrCodeDatabase      = "DATABASE_ERROR"
    ErrCodeCache         = "CACHE_ERROR"
    ErrCodeMessage       = "MESSAGE_ERROR"
)
```

### 统一响应格式

```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *AppError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Page       int `json:"page,omitempty"`
    PageSize   int `json:"page_size,omitempty"`
    TotalCount int `json:"total_count,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}
```

### 错误处理中间件

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            // 根据错误类型返回相应的HTTP状态码和错误信息
        }
    }
}
```

## 测试策略

### 1. 单元测试

- **覆盖范围**: 所有service层、repository层和utility函数
- **测试工具**: Go标准testing包 + testify/assert
- **Mock策略**: 使用testify/mock或gomock生成mock对象
- **测试数据**: 使用工厂模式创建测试数据

### 2. 集成测试

- **API测试**: 使用httptest包测试HTTP端点
- **数据库测试**: 使用测试数据库或内存数据库
- **外部服务**: 使用testcontainers进行真实环境测试

### 3. 测试配置

```go
type TestConfig struct {
    Database TestDatabaseConfig `mapstructure:"database"`
    Redis    TestRedisConfig    `mapstructure:"redis"`
    Kafka    TestKafkaConfig    `mapstructure:"kafka"`
}
```

### 4. 测试工具

- **数据库迁移**: 每个测试前后自动迁移和清理
- **测试数据**: 使用fixtures或工厂模式
- **并发测试**: 确保测试可以并发运行
- **覆盖率**: 目标覆盖率80%以上

## 日志策略

### 日志级别和格式

```go
type LogConfig struct {
    Level      string `mapstructure:"level"`      // debug, info, warn, error
    Format     string `mapstructure:"format"`     // json, text
    Output     string `mapstructure:"output"`     // stdout, file
    Filename   string `mapstructure:"filename"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxBackups int    `mapstructure:"max_backups"`
    MaxAge     int    `mapstructure:"max_age"`
    Compress   bool   `mapstructure:"compress"`
}
```

### 结构化日志

```go
logger.WithFields(logrus.Fields{
    "user_id":    userID,
    "action":     "create_user",
    "request_id": requestID,
    "duration":   duration,
}).Info("User created successfully")
```

### 日志内容

- **HTTP请求**: 请求方法、路径、状态码、响应时间
- **数据库操作**: SQL查询、执行时间、影响行数
- **外部服务调用**: 服务名称、请求参数、响应时间
- **错误信息**: 错误堆栈、上下文信息
- **业务事件**: 用户操作、系统状态变更

## 部署和运维

### Docker配置

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
CMD ["./main"]
```

### 健康检查

```go
func (h *HealthHandler) Check(c *gin.Context) {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Services: map[string]string{
            "database": h.checkDatabase(),
            "redis":    h.checkRedis(),
            "kafka":    h.checkKafka(),
        },
    }
    c.JSON(http.StatusOK, status)
}
```

### 监控指标

- **应用指标**: 请求数量、响应时间、错误率
- **系统指标**: CPU使用率、内存使用率、磁盘空间
- **业务指标**: 用户注册数、活跃用户数、业务操作数量