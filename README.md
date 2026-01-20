# config-center

基于 Go 语言的分布式配置中心，采用领域驱动设计（DDD）架构。

## 核心功能

- **长轮询**: 实现配置变更的实时推送
- **配置管理**: 配置的增删改查，支持多环境、多命名空间
- **订阅管理**: 客户端订阅配置，监听变更
- **变更管理**: 配置变更历史记录和审计

## 技术栈

- **语言**: Go 1.24.11
- **HTTP 框架**: Hertz (CloudWeGo)
- **ORM**: GORM
- **数据库**: PostgreSQL 16
- **架构模式**: DDD（领域驱动设计）
- **依赖管理**: Go Workspace + BOM

## 项目结构

```
config-center/
├── go.work                   # Go 工作区配置
├── bom/                      # BOM 依赖管理模块
├── share/                    # 公共组件模块
│   ├── errors/               # 错误定义
│   ├── types/                # 通用类型
│   └── repository/           # 仓储基类和GORM实现
├── config/                   # 配置聚合模块（待开发）
│   ├── domain/               # 领域层
│   │   ├── entity/           # 配置实体
│   │   ├── repository/       # 仓储接口
│   │   └── service/          # 领域服务
│   └── infrastructure/       # 基础设施层
│       ├── entity/           # 数据库实体
│       └── repository/       # 仓储实现
├── api/                      # API 聚合模块（待开发）
│   └── config-api/           # 配置 API
│       ├── dto/              # 数据传输对象
│       ├── service/          # 应用服务
│       └── http/             # HTTP 处理器
└── cmd/
    └── api/                  # 主程序入口
```

## 快速开始

### 1. 同步依赖

```bash
go work sync
```

### 2. 启动数据库服务

```bash
docker-compose up -d postgres
```

### 3. 运行应用

```bash
go run ./cmd/api/main.go
```

访问 http://localhost:8080/health 检查服务状态。

## 环境变量

- `DB_HOST`: PostgreSQL 主机（默认：localhost）
- `DB_PORT`: PostgreSQL 端口（默认：5432）
- `DB_USER`: 数据库用户（默认：postgres）
- `DB_PASSWORD`: 数据库密码（默认：postgres）
- `DB_NAME`: 数据库名称（默认：config-center）
- `PORT`: 服务端口（默认：8080）

## 常用命令

```bash
# 同步依赖
go work sync

# 启动 Docker 服务
docker-compose up -d

# 停止 Docker 服务
docker-compose down

# 运行应用
go run ./cmd/api/main.go
```

## 开发计划

- [ ] 配置模型设计（配置项、命名空间、环境）
- [ ] 配置CRUD接口
- [ ] 长轮询实现
- [ ] 订阅管理机制
- [ ] 变更历史记录
- [ ] 内存缓存层
- [ ] 配置监听器接口设计

## License

MIT
