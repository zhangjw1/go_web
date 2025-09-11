# Go Web Starter

一个开箱即用的Go web应用程序模板，集成了现代web开发所需的核心组件和工具。

## 功能特性

- 🚀 **Gin Web框架** - 高性能HTTP web框架
- 🗄️ **GORM + MySQL** - 强大的ORM和数据库支持
- 📨 **Kafka消息队列** - 异步消息处理
- ⚡ **Redis缓存** - 高性能缓存系统
- ⚙️ **Viper配置管理** - 灵活的配置管理
- 📝 **结构化日志** - 完善的日志记录系统
- 🧪 **完整测试支持** - 单元测试和集成测试
- 🐳 **Docker支持** - 容器化部署

## 项目结构

```
go-web-starter/
├── cmd/                    # 应用程序入口
├── internal/               # 私有应用程序代码
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP处理器
│   ├── service/           # 业务逻辑层
│   ├── repository/        # 数据访问层
│   ├── model/             # 数据模型
│   ├── infrastructure/    # 基础设施层
│   └── app/               # 应用程序初始化
├── pkg/                   # 公共库代码
├── test/                  # 测试文件
├── configs/               # 配置文件
├── deployments/           # 部署配置
├── scripts/               # 脚本文件
└── docs/                  # 文档
```

## 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Kafka 2.8+

### 安装和运行

1. 克隆项目
```bash
git clone <repository-url>
cd go-web-starter
```

2. 安装依赖
```bash
make deps
```

3. 配置环境
```bash
cp configs/config.yaml.example configs/config.yaml
# 编辑配置文件
```

4. 运行应用
```bash
make run
```

### 开发环境

设置开发环境：
```bash
make setup
```

使用热重载运行：
```bash
make dev
```

## 可用命令

```bash
make help          # 显示帮助信息
make build          # 构建应用程序
make test           # 运行测试
make coverage       # 运行测试并生成覆盖率报告
make run            # 运行应用程序
make clean          # 清理构建文件
make deps           # 下载依赖
make fmt            # 格式化代码
make lint           # 代码检查
make docker-build   # 构建Docker镜像
make docker-run     # 运行Docker容器
```

## 配置

应用程序支持多种配置方式：

1. 配置文件 (`configs/config.yaml`)
2. 环境变量
3. 命令行参数

详细配置说明请参考 [配置文档](docs/configuration.md)。

## API文档

API文档请参考 [API文档](docs/api.md)。

## 贡献

欢迎贡献代码！请阅读 [贡献指南](CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证。详情请参考 [LICENSE](LICENSE) 文件。# go_web
