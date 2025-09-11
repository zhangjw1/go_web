# 配置管理

本项目使用 [Viper](https://github.com/spf13/viper) 进行配置管理，支持多种配置源和环境。

## 配置文件

### 配置文件位置

配置文件按以下优先级查找：
1. 通过 `Load(configPath)` 指定的路径
2. `./configs/config.yaml`
3. `./config.yaml`

### 环境配置文件

- `configs/config.yaml` - 默认配置
- `configs/config.dev.yaml` - 开发环境配置
- `configs/config.prod.yaml.example` - 生产环境配置模板

## 环境变量

所有配置项都可以通过环境变量覆盖，环境变量名规则：
- 将配置路径中的 `.` 替换为 `_`
- 转换为大写

例如：
- `server.port` → `SERVER_PORT`
- `database.host` → `DATABASE_HOST`
- `logger.max_size` → `LOGGER_MAX_SIZE`

## 配置结构

### 服务器配置 (server)

```yaml
server:
  port: "8080"              # 服务端口
  mode: "debug"             # 运行模式: debug, release, test
  read_timeout: 30          # 读取超时时间(秒)
  write_timeout: 30         # 写入超时时间(秒)
```

### 数据库配置 (database)

```yaml
database:
  host: "localhost"         # 数据库主机
  port: 3306               # 数据库端口
  username: "root"         # 用户名
  password: ""             # 密码
  database: "go_web_starter" # 数据库名
  charset: "utf8mb4"       # 字符集
```

### Redis配置 (redis)

```yaml
redis:
  host: "localhost"        # Redis主机
  port: 6379              # Redis端口
  password: ""            # Redis密码
  database: 0             # Redis数据库编号
```

### Kafka配置 (kafka)

```yaml
kafka:
  brokers:                # Kafka代理列表
    - "localhost:9092"
  group_id: "go-web-starter" # 消费者组ID
```

### 日志配置 (logger)

```yaml
logger:
  level: "info"           # 日志级别: debug, info, warn, error
  format: "json"          # 日志格式: json, text
  output: "stdout"        # 输出方式: stdout, file
  filename: "logs/app.log" # 日志文件路径
  max_size: 100           # 单个日志文件最大大小(MB)
  max_backups: 3          # 保留的日志文件数量
  max_age: 28             # 日志文件保留天数
  compress: true          # 是否压缩旧日志文件
```

## 使用方法

### 基本用法

```go
package main

import (
    "log"
    "go-web-starter/internal/config"
)

func main() {
    // 加载默认配置
    cfg, err := config.Load("")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // 使用配置
    fmt.Printf("Server will run on port: %s\n", cfg.Server.Port)
}
```

### 指定配置文件

```go
// 加载特定配置文件
cfg, err := config.Load("configs/config.dev.yaml")
```

### 配置热重载

```go
// 启用配置热重载
err := config.WatchConfig(func(newConfig *config.Config) {
    fmt.Println("Configuration reloaded!")
    // 处理配置更新
})
```

## 环境变量示例

创建 `.env` 文件：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DATABASE_HOST=localhost
DATABASE_USERNAME=root
DATABASE_PASSWORD=password
DATABASE_DATABASE=go_web_starter

# Redis配置
REDIS_HOST=localhost
REDIS_PASSWORD=

# Kafka配置
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=go-web-starter

# 日志配置
LOGGER_LEVEL=info
LOGGER_FORMAT=json
```

## 配置验证

配置加载时会自动进行验证，确保：
- 必需字段不为空
- 枚举值在有效范围内
- 数值在合理范围内

验证失败时会返回详细的错误信息。

## 最佳实践

1. **开发环境**: 使用 `config.dev.yaml` 和本地环境变量
2. **生产环境**: 使用环境变量覆盖敏感配置（密码、密钥等）
3. **配置文件**: 不要将包含敏感信息的配置文件提交到版本控制
4. **默认值**: 为所有配置项设置合理的默认值
5. **文档**: 及时更新配置文档，说明各配置项的作用