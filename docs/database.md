# 数据库系统

本项目使用 [GORM](https://gorm.io/) 作为ORM框架，支持MySQL数据库，提供完整的数据库连接管理、模型定义和迁移功能。

## 功能特性

- 🗄️ **MySQL支持**: 使用GORM MySQL驱动
- 🔗 **连接池管理**: 自动连接池配置和管理
- 📊 **基础模型**: 统一的BaseModel包含常用字段
- 🔄 **自动迁移**: 数据库结构自动迁移
- 📝 **软删除**: 内置软删除支持
- 📄 **分页查询**: 内置分页、排序、过滤功能
- 🔍 **查询日志**: 数据库查询日志记录
- 🏥 **健康检查**: 数据库连接健康检查
- 📈 **连接统计**: 连接池统计信息
- 🔒 **事务支持**: 数据库事务管理

## 配置

### 数据库配置结构

```yaml
database:
  host: "localhost"              # 数据库主机
  port: 3306                     # 数据库端口
  username: "root"               # 用户名
  password: "password"           # 密码
  database: "go_web_starter"     # 数据库名
  charset: "utf8mb4"             # 字符集
  log_level: "error"             # 日志级别: silent, error, warn, info
  max_idle_conns: 10             # 最大空闲连接数
  max_open_conns: 100            # 最大打开连接数
  conn_max_lifetime: 60          # 连接最大生存时间(分钟)
```

### 环境变量

```bash
DATABASE_HOST=localhost
DATABASE_PORT=3306
DATABASE_USERNAME=root
DATABASE_PASSWORD=password
DATABASE_DATABASE=go_web_starter
DATABASE_CHARSET=utf8mb4
DATABASE_LOG_LEVEL=error
DATABASE_MAX_IDLE_CONNS=10
DATABASE_MAX_OPEN_CONNS=100
DATABASE_CONN_MAX_LIFETIME=60
```

## 使用方法

### 基本连接

```go
package main

import (
    \"go-web-starter/internal/config\"
    \"go-web-starter/internal/infrastructure/database\"
    \"go-web-starter/internal/infrastructure/logger\"
)

func main() {
    // 创建日志器
    log, _ := logger.New(loggerConfig)
    
    // 数据库配置
    dbConfig := &config.DatabaseConfig{
        Host:            \"localhost\",
        Port:            3306,
        Username:        \"root\",
        Password:        \"password\",
        Database:        \"go_web_starter\",
        LogLevel:        \"error\",
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: 60,
    }
    
    // 创建数据库连接
    db, err := database.New(dbConfig, log)
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // 使用数据库
    // ...
}
```

## 基础模型

### BaseModel

所有模型都应该嵌入BaseModel：

```go
package model

import (
    \"go-web-starter/internal/domain/model\"
)

type User struct {
    model.BaseModel
    Name  string `json:\"name\" gorm:\"not null;size:100\"`
    Email string `json:\"email\" gorm:\"uniqueIndex;not null;size:255\"`
}
```

BaseModel包含以下字段：

- `ID`: 主键，自增
- `CreatedAt`: 创建时间，自动设置
- `UpdatedAt`: 更新时间，自动更新
- `DeletedAt`: 删除时间，软删除标记

### 模型接口

```go
// Model 所有模型都应实现的接口
type Model interface {
    GetID() uint
    GetCreatedAt() time.Time
    GetUpdatedAt() time.Time
    GetDeletedAt() *time.Time
    IsDeleted() bool
}

// TableNamer 自定义表名
type TableNamer interface {
    TableName() string
}

// Validator 自定义验证
type Validator interface {
    Validate() error
}
```

### 模型钩子

```go
// 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 调用父类钩子
    if err := u.BaseModel.BeforeCreate(tx); err != nil {
        return err
    }
    
    // 自定义逻辑
    if u.Status == \"\" {
        u.Status = \"active\"
    }
    
    return nil
}

// 更新前钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    return u.BaseModel.BeforeUpdate(tx)
}
```

## 数据库迁移

### 自动迁移

```go
// 创建迁移器
migrator := database.NewMigrator(db.DB, log)

// 自动迁移模型
err := migrator.AutoMigrate(&User{}, &Profile{})
if err != nil {
    log.Fatal(\"Migration failed:\", err)
}

// 创建自定义索引
err = migrator.CreateIndexes()
if err != nil {
    log.Fatal(\"Index creation failed:\", err)
}
```

### 迁移状态

```go
// 获取迁移状态
status := migrator.GetMigrationStatus(&User{}, &Profile{})
fmt.Printf(\"Migration status: %+v\\n\", status)
```

### 数据库重置（危险操作）

```go
// 重置数据库（删除所有表并重新创建）
err := migrator.Reset(&User{}, &Profile{})
if err != nil {
    log.Fatal(\"Database reset failed:\", err)
}
```

## CRUD操作

### 创建

```go
user := &User{
    Name:  \"John Doe\",
    Email: \"john@example.com\",
}

// 验证（如果实现了Validator接口）
if err := user.Validate(); err != nil {
    return err
}

// 创建记录
result := db.Create(user)
if result.Error != nil {
    return result.Error
}

fmt.Printf(\"Created user with ID: %d\\n\", user.ID)
```

### 查询

```go
// 根据ID查询
var user User
result := db.First(&user, 1)
if result.Error != nil {
    return result.Error
}

// 根据条件查询
var users []User
result = db.Where(\"status = ?\", \"active\").Find(&users)
if result.Error != nil {
    return result.Error
}

// 查询单条记录
var user User
result = db.Where(\"email = ?\", \"john@example.com\").First(&user)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // 记录不存在
        return nil
    }
    return result.Error
}
```

### 更新

```go
// 更新单个字段
result := db.Model(&user).Update(\"name\", \"Jane Doe\")
if result.Error != nil {
    return result.Error
}

// 更新多个字段
result = db.Model(&user).Updates(User{
    Name:   \"Jane Doe\",
    Status: \"inactive\",
})
if result.Error != nil {
    return result.Error
}

// 批量更新
result = db.Model(&User{}).Where(\"status = ?\", \"pending\").Update(\"status\", \"active\")
if result.Error != nil {
    return result.Error
}
```

### 删除

```go
// 软删除
result := db.Delete(&user)
if result.Error != nil {
    return result.Error
}

// 硬删除
result = db.Unscoped().Delete(&user)
if result.Error != nil {
    return result.Error
}

// 批量删除
result = db.Where(\"status = ?\", \"inactive\").Delete(&User{})
if result.Error != nil {
    return result.Error
}
```

## 分页查询

### 基本分页

```go
// 分页参数
params := &model.PaginationParams{
    Page:     1,
    PageSize: 10,
}
params.SetDefaults()

var users []User
var total int64

// 计算总数
db.Model(&User{}).Count(&total)

// 分页查询
result := db.Offset(params.GetOffset()).Limit(params.GetLimit()).Find(&users)
if result.Error != nil {
    return result.Error
}

// 创建分页结果
paginationResult := model.NewPaginationResult(params, total, users)
```

### 排序查询

```go
// 排序参数
sortParams := &model.SortParams{
    SortBy:    \"created_at\",
    SortOrder: \"desc\",
}
sortParams.SetDefaults()

var users []User
result := db.Order(sortParams.GetOrderBy()).Find(&users)
if result.Error != nil {
    return result.Error
}
```

### 过滤查询

```go
// 过滤参数
filterParams := &model.FilterParams{
    Search: \"john\",
    Status: \"active\",
}

query := db.Model(&User{})

// 搜索过滤
if filterParams.Search != \"\" {
    query = query.Where(\"name LIKE ? OR email LIKE ?\", 
        \"%\"+filterParams.Search+\"%\", 
        \"%\"+filterParams.Search+\"%\")
}

// 状态过滤
if filterParams.Status != \"\" {
    query = query.Where(\"status = ?\", filterParams.Status)
}

var users []User
result := query.Find(&users)
if result.Error != nil {
    return result.Error
}
```

## 事务管理

### 基本事务

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&user1).Error; err != nil {
        return err // 自动回滚
    }
    
    if err := tx.Create(&user2).Error; err != nil {
        return err // 自动回滚
    }
    
    return nil // 提交事务
})

if err != nil {
    log.Error(\"Transaction failed:\", err)
}
```

### 手动事务

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user1).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&user2).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error
```

## 关联关系

### 一对一

```go
type User struct {
    model.BaseModel
    Name      string   `json:\"name\"`
    Profile   Profile  `json:\"profile\" gorm:\"foreignKey:UserID\"`
    ProfileID uint     `json:\"profile_id\"`
}

type Profile struct {
    model.BaseModel
    UserID uint   `json:\"user_id\"`
    Bio    string `json:\"bio\"`
}

// 预加载关联
var user User
db.Preload(\"Profile\").First(&user, 1)
```

### 一对多

```go
type User struct {
    model.BaseModel
    Name  string  `json:\"name\"`
    Posts []Post  `json:\"posts\" gorm:\"foreignKey:UserID\"`
}

type Post struct {
    model.BaseModel
    UserID uint   `json:\"user_id\"`
    Title  string `json:\"title\"`
}

// 预加载关联
var user User
db.Preload(\"Posts\").First(&user, 1)
```

### 多对多

```go
type User struct {
    model.BaseModel
    Name  string `json:\"name\"`
    Roles []Role `json:\"roles\" gorm:\"many2many:user_roles;\"`
}

type Role struct {
    model.BaseModel
    Name  string `json:\"name\"`
    Users []User `json:\"users\" gorm:\"many2many:user_roles;\"`
}

// 预加载关联
var user User
db.Preload(\"Roles\").First(&user, 1)
```

## 健康检查

```go
// 检查数据库连接健康状态
if err := db.Health(); err != nil {
    log.Error(\"Database health check failed:\", err)
} else {
    log.Info(\"Database is healthy\")
}

// 获取连接统计信息
stats := db.GetStats()
log.Info(\"Database stats:\", stats)
```

## 查询日志

```go
// 记录查询日志
start := time.Now()
var count int64
db.Model(&User{}).Count(&count)
duration := time.Since(start)

db.LogQuery(\"SELECT COUNT(*) FROM users\", duration, count)
```

## 最佳实践

### 1. 模型设计

```go
// 好的做法
type User struct {
    model.BaseModel
    Name     string `json:\"name\" gorm:\"not null;size:100\"`
    Email    string `json:\"email\" gorm:\"uniqueIndex;not null;size:255\"`
    Status   string `json:\"status\" gorm:\"default:'active';size:20\"`
    Metadata string `json:\"metadata\" gorm:\"type:json\"`
}

// 实现TableName方法
func (User) TableName() string {
    return \"users\"
}

// 实现Validate方法
func (u *User) Validate() error {
    if u.Name == \"\" {
        return errors.New(\"name is required\")
    }
    if u.Email == \"\" {
        return errors.New(\"email is required\")
    }
    return nil
}
```

### 2. 查询优化

```go
// 使用索引
db.Where(\"email = ?\", email).First(&user) // email有索引

// 预加载关联
db.Preload(\"Profile\").Preload(\"Posts\").Find(&users)

// 选择特定字段
db.Select(\"id\", \"name\", \"email\").Find(&users)

// 批量操作
db.CreateInBatches(users, 100)
```

### 3. 错误处理

```go
result := db.First(&user, id)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf(\"failed to find user: %w\", result.Error)
}
```

### 4. 连接池配置

```go
// 生产环境推荐配置
database:
  max_idle_conns: 10      # 空闲连接数
  max_open_conns: 100     # 最大连接数
  conn_max_lifetime: 60   # 连接生存时间(分钟)
```

### 5. 安全考虑

```go
// 使用参数化查询防止SQL注入
db.Where(\"name = ?\", name).Find(&users) // ✅

// 避免字符串拼接
db.Where(fmt.Sprintf(\"name = '%s'\", name)).Find(&users) // ❌
```

## 性能监控

### 连接池监控

```go
stats := db.GetStats()
log.WithFields(logrus.Fields{
    \"max_open\":     stats[\"max_open_connections\"],
    \"open\":         stats[\"open_connections\"],
    \"in_use\":       stats[\"in_use\"],
    \"idle\":         stats[\"idle\"],
    \"wait_count\":   stats[\"wait_count\"],
    \"wait_duration\": stats[\"wait_duration\"],
}).Info(\"Database connection stats\")
```

### 慢查询监控

```go
// 在GORM配置中启用慢查询日志
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // 记录所有SQL
})

// 或者只记录慢查询
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Warn).
        SlowThreshold(200 * time.Millisecond), // 200ms以上的查询
})
```

## 故障排除

### 常见问题

1. **连接超时**
   ```
   Error: dial tcp: i/o timeout
   ```
   - 检查数据库服务是否运行
   - 检查网络连接
   - 检查防火墙设置

2. **连接被拒绝**
   ```
   Error: connection refused
   ```
   - 检查数据库端口是否正确
   - 检查数据库是否监听指定端口

3. **认证失败**
   ```
   Error: Access denied for user
   ```
   - 检查用户名和密码
   - 检查用户权限

4. **数据库不存在**
   ```
   Error: Unknown database
   ```
   - 创建数据库
   - 检查数据库名称拼写

### 调试技巧

```go
// 启用详细日志
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// 打印SQL语句
db.Debug().Where(\"name = ?\", \"john\").First(&user)

// 获取生成的SQL
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Model(&User{}).Where(\"name = ?\", \"john\").Find(&users)
})
fmt.Println(sql)
```