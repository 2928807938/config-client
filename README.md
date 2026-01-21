# 配置中心 Go SDK

Go 语言实现的配置中心客户端 SDK，提供配置管理和实时监听能力。

## 快速开始

### 安装

```bash
go get github.com/yourorg/config-client-go
```

**环境要求:** Go 1.18+

### 基础使用

```go
package main

import (
    "fmt"
    configsdk "github.com/yourorg/config-client-go"
)

func main() {
    // 1. 创建客户端
    client, err := configsdk.New(
        configsdk.WithServerURL("http://localhost:8080"),
        configsdk.WithNamespace("default"),
    )
    if err != nil {
        panic(err)
    }
    defer client.Stop()

    // 2. 获取配置
    dbURL := client.GetString("database.url")
    port := client.GetIntOrDefault("server.port", 8080)

    fmt.Printf("数据库: %s, 端口: %d\n", dbURL, port)

    // 3. 监听配置变更
    client.GetAndWatch("database.url", func(key, value string) {
        fmt.Printf("配置已更新: %s\n", value)
    })

    select {} // 保持运行
}
```

---

## 核心功能

### 1. 获取配置

#### 基本类型
```go
// 字符串
appName := client.GetString("app.name")
appName := client.GetOrDefault("app.name", "default-app")

// 整数
port := client.GetIntOrDefault("server.port", 8080)

// 浮点数
ratio := client.GetFloat64OrDefault("rate.limit", 0.8)

// 布尔值（支持 true/false, 1/0, yes/no, on/off）
enabled := client.GetBoolOrDefault("feature.enabled", false)

// 字符串数组（逗号分隔）
hosts := client.GetStringSliceOrDefault("redis.hosts", []string{"localhost"})
```

#### 批量操作
```go
// 获取所有配置
allConfigs := client.GetAll()

// 按前缀获取
dbConfigs := client.GetByPrefix("database.")
// 返回: {"url": "...", "port": "5432", "username": "..."}

// 检查是否存在
if client.Has("feature.new") {
    // ...
}
```

### 2. 监听配置变更

```go
// 方式 1: 仅监听变更
client.Watch("database.url", func(key, value string) {
    fmt.Printf("配置变更: %s = %s\n", key, value)
})

// 方式 2: 获取当前值 + 监听变更（推荐）
client.GetAndWatch("database.url", func(key, value string) {
    fmt.Printf("数据库地址: %s\n", value)
    // 首次调用返回当前值，后续配置变更时再次调用
})

// 取消监听
client.Unwatch("database.url")
```

### 3. 刷新配置

```go
// 刷新单个配置
client.Refresh("database.url")

// 刷新所有配置
client.RefreshAll()
```

---

## 配置选项

### 基础配置

```go
client, err := configsdk.New(
    // 必填: 服务器地址
    configsdk.WithServerURL("http://localhost:8080"),

    // 命名空间（二选一）
    configsdk.WithNamespace("production"),      // 使用名称
    configsdk.WithNamespaceID(1),              // 使用 ID

    // 自动启动（默认 true）
    configsdk.WithAutoStart(true),

    // 启用本地缓存（默认 true）
    configsdk.WithCache(true),

    // 初始化时拉取所有配置（默认 true）
    configsdk.WithFetchOnInit(true),
)
```

### 降级配置

当配置中心不可用时使用的默认值:

```go
configsdk.WithFallback(map[string]string{
    "database.url": "postgres://localhost:5432/mydb",
    "redis.host":   "localhost:6379",
    "timeout":      "30",
})
```

### 监听器模式

#### HTTP 长轮询（默认，推荐）

适合大多数场景，无需额外依赖:

```go
client, err := configsdk.New(
    configsdk.WithServerURL("http://localhost:8080"),
    configsdk.WithHTTPWatcher(60 * time.Second), // 可选: 轮询超时时间
)
```

**特点:**
- ✅ 部署简单，无需 Redis
- ✅ 适合大规模客户端
- ⏱️ 延迟: 1-3 秒

#### Redis 订阅模式

适合对实时性要求高的场景:

```go
import "github.com/redis/go-redis/v9"

// 方式 1: 使用已有 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
client, err := configsdk.New(
    configsdk.WithServerURL("http://localhost:8080"),
    configsdk.WithRedisWatcher(redisClient),
)

// 方式 2: 使用 Redis 配置
client, err := configsdk.New(
    configsdk.WithServerURL("http://localhost:8080"),
    configsdk.WithRedisOptions(&redis.Options{
        Addr:     "localhost:6379",
        Password: "secret",
        DB:       0,
    }),
)
```

**特点:**
- ✅ 毫秒级延迟
- ✅ 实时推送
- ⚠️ 需要 Redis 依赖

---

## 完整示例

### 示例 1: Web 应用配置管理

```go
package main

import (
    "database/sql"
    "log"
    "sync"
    configsdk "github.com/yourorg/config-client-go"
)

type AppConfig struct {
    mu          sync.RWMutex
    DatabaseURL string
    RedisHost   string
    LogLevel    string
}

func (c *AppConfig) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    switch key {
    case "database.url":
        c.DatabaseURL = value
    case "redis.host":
        c.RedisHost = value
    case "log.level":
        c.LogLevel = value
    }
}

func (c *AppConfig) Get() (string, string, string) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.DatabaseURL, c.RedisHost, c.LogLevel
}

var (
    appConfig = &AppConfig{}
    db        *sql.DB
)

func main() {
    // 创建配置客户端
    client, err := configsdk.New(
        configsdk.WithServerURL("http://localhost:8080"),
        configsdk.WithNamespace("production"),
        configsdk.WithFallback(map[string]string{
            "database.url": "postgres://localhost:5432/mydb",
            "redis.host":   "localhost:6379",
            "log.level":    "info",
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // 加载初始配置
    appConfig.Set("database.url", client.GetString("database.url"))
    appConfig.Set("redis.host", client.GetString("redis.host"))
    appConfig.Set("log.level", client.GetOrDefault("log.level", "info"))

    dbURL, redisHost, logLevel := appConfig.Get()
    log.Printf("配置已加载 - 数据库: %s, Redis: %s, 日志: %s",
        dbURL, redisHost, logLevel)

    // 初始化数据库
    db, _ = sql.Open("postgres", dbURL)
    defer db.Close()

    // 监听配置变更
    client.Watch("database.url", func(key, value string) {
        appConfig.Set(key, value)
        log.Printf("数据库配置变更，重新连接: %s", value)
        db.Close()
        db, _ = sql.Open("postgres", value)
    })

    client.Watch("log.level", func(key, value string) {
        appConfig.Set(key, value)
        log.Printf("日志级别变更: %s", value)
        // 更新日志级别...
    })

    // 运行应用
    log.Println("应用已启动")
    select {}
}
```

### 示例 2: 简单配置监听

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    configsdk "github.com/yourorg/config-client-go"
)

func main() {
    client, err := configsdk.New(
        configsdk.WithServerURL("http://localhost:8080"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // 获取并监听多个配置
    client.GetAndWatch("database.url", func(key, value string) {
        fmt.Printf("数据库地址: %s\n", value)
    })

    client.GetAndWatch("redis.host", func(key, value string) {
        fmt.Printf("Redis 地址: %s\n", value)
    })

    fmt.Println("监听器已启动，按 Ctrl+C 退出...")

    // 优雅退出
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    fmt.Println("\n正在关闭...")
}
```

---

## 最佳实践

### ✅ 推荐做法

#### 1. 使用 `GetAndWatch` 简化代码

```go
// 一行代码搞定获取 + 监听
client.GetAndWatch("database.url", func(key, value string) {
    if db == nil {
        initDB(value)
    } else {
        reconnectDB(value)
    }
})
```

#### 2. 使用带默认值的方法

```go
// 简洁且安全
port := client.GetIntOrDefault("server.port", 8080)
enabled := client.GetBoolOrDefault("feature.enabled", false)
```

#### 3. 配置降级保障

```go
client, _ := configsdk.New(
    configsdk.WithServerURL("http://localhost:8080"),
    configsdk.WithFallback(map[string]string{
        "database.url": "postgres://localhost:5432/mydb",
        "timeout":      "30",
    }),
)
// 配置中心故障时自动使用降级配置
```

#### 4. 优雅关闭

```go
func main() {
    client, _ := configsdk.New(/*...*/)
    defer client.Stop()  // 确保资源释放

    // 或监听信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        client.Stop()
        os.Exit(0)
    }()
}
```

### ❌ 避免做法

#### 1. 重复获取和监听

```go
// ❌ 不推荐
dbURL := client.GetString("database.url")
initDB(dbURL)
client.Watch("database.url", func(key, value string) {
    reconnectDB(value)
})

// ✅ 推荐
client.GetAndWatch("database.url", func(key, value string) {
    // 处理逻辑
})
```

#### 2. 忘记处理错误

```go
// ❌ 不推荐
port, _ := client.GetInt("server.port")

// ✅ 推荐
port := client.GetIntOrDefault("server.port", 8080)
```

---

## 常见问题

### Q1: 初始化时会自动拉取配置吗？

**答:** 是的。默认 `WithFetchOnInit(true)`，SDK 初始化时会自动拉取所有配置到本地缓存。

---

### Q2: 配置变更的延迟是多少？

**答:**
- **HTTP 长轮询模式**: 1-3 秒
- **Redis 订阅模式**: 毫秒级

---

### Q3: 网络断开后会自动重连吗？

**答:** 会。SDK 内置自动重连机制，网络恢复后会自动重新连接。

---

### Q4: 监听回调是同步还是异步？

**答:** 异步。回调函数在独立的 goroutine 中执行，不会阻塞主流程。

---

### Q5: 可以同时监听多个配置吗？

**答:** 可以。每个配置独立监听：

```go
client.Watch("config1", callback1)
client.Watch("config2", callback2)
client.Watch("config3", callback3)
```

---

### Q6: 如何选择监听模式？

**答:**

| 场景 | 推荐模式 | 原因 |
|------|---------|------|
| 一般业务系统 | HTTP 长轮询 | 部署简单，1-3秒延迟可接受 |
| 高实时性要求 | Redis 订阅 | 毫秒级延迟 |
| 大规模客户端 | HTTP 长轮询 | 无需维护 Redis |
| 已有 Redis 基础设施 | Redis 订阅 | 充分利用现有资源 |

---

### Q7: 缓存会自动更新吗？

**答:** 会。监听到配置变更时会自动更新本地缓存，无需手动刷新。

---

### Q8: 配置不存在时会怎样？

**答:**
- 使用 `Get()` / `GetInt()` 等方法会返回错误
- 使用 `GetOrDefault()` / `GetIntOrDefault()` 等方法会返回默认值
- 如果配置了 `Fallback`，会尝试使用降级配置

---

## API 参考

### 客户端创建

| 方法 | 说明 |
|------|------|
| `New(opts ...Option) (*Client, error)` | 创建并初始化客户端 |

### 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithServerURL(url)` | 配置中心服务地址 | - |
| `WithNamespace(name)` | 命名空间名称 | default |
| `WithNamespaceID(id)` | 命名空间 ID | - |
| `WithHTTPWatcher(timeout)` | HTTP 长轮询（推荐） | 60s |
| `WithRedisWatcher(client)` | Redis 订阅 | - |
| `WithRedisOptions(opts)` | Redis 连接配置 | - |
| `WithAutoStart(enabled)` | 自动启动 | true |
| `WithCache(enabled)` | 启用缓存 | true |
| `WithFetchOnInit(enabled)` | 初始化时拉取配置 | true |
| `WithFallback(configs)` | 降级配置 | - |

### 获取配置

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `Get(key)` | (string, error) | 获取配置 |
| `GetString(key)` | string | 获取字符串（空值返回 ""） |
| `GetInt(key)` | (int, error) | 获取整数 |
| `GetFloat64(key)` | (float64, error) | 获取浮点数 |
| `GetBool(key)` | (bool, error) | 获取布尔值 |
| `GetStringSlice(key)` | ([]string, error) | 获取字符串数组 |
| `GetOrDefault(key, def)` | string | 带默认值的字符串 |
| `GetIntOrDefault(key, def)` | int | 带默认值的整数 |
| `GetFloat64OrDefault(key, def)` | float64 | 带默认值的浮点数 |
| `GetBoolOrDefault(key, def)` | bool | 带默认值的布尔值 |
| `GetStringSliceOrDefault(key, def)` | []string | 带默认值的数组 |

### 批量操作

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `GetAll()` | map[string]string | 获取所有配置 |
| `GetByPrefix(prefix)` | map[string]string | 根据前缀获取配置 |
| `Has(key)` | bool | 检查配置是否存在 |

### 监听与刷新

| 方法 | 说明 |
|------|------|
| `Watch(key, callback)` | 监听配置变更 |
| `GetAndWatch(key, callback)` | 获取当前值并监听变更 |
| `Unwatch(key)` | 取消监听 |
| `Refresh(key)` | 刷新单个配置 |
| `RefreshAll()` | 刷新所有配置 |

### 生命周期

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动客户端（通常自动启动） |
| `Stop()` | 停止客户端并释放资源 |
| `IsRunning()` | 检查客户端运行状态 |

---

## 技术支持

- **问题反馈**: 提交 Issue
- **功能建议**: 联系开发团队

## 许可证

MIT License
