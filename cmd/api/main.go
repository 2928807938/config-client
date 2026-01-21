package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"config-client/api/config-api/converter"
	configHttp "config-client/api/config-api/http"
	"config-client/api/config-api/service"
	domainService "config-client/config/domain/service"
	infraListener "config-client/config/infrastructure/listener"
	infraRepository "config-client/config/infrastructure/repository"
	"config-client/share/config"
	"config-client/share/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	cfg                *config.Config
	db                 *gorm.DB
	rdb                *redis.Client
	ctx                = context.Background()
	hertzH             *server.Hertz
	longPollingService *domainService.LongPollingService
	configListener     *infraListener.RedisConfigListener
)

func main() {
	// 1. 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	hlog.Infof("配置加载成功")

	// 2. 初始化数据库
	if err := initDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	hlog.Infof("数据库连接成功")

	// 3. 初始化Redis
	if err := initRedis(); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	hlog.Infof("Redis连接成功")

	// 4. 初始化长轮询管理器
	if err := initLongPolling(); err != nil {
		log.Fatalf("初始化长轮询管理器失败: %v", err)
	}
	hlog.Infof("长轮询管理器初始化成功")

	// 5. 初始化HTTP服务器
	initServer()
	hlog.Infof("HTTP服务器初始化完成，监听端口: %d", cfg.Server.Port)

	// 6. 启动服务器（非阻塞）
	go func() {
		if err := hertzH.Run(); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 6. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	hlog.Info("正在关闭服务器...")

	// 7. 优雅关闭
	gracefulShutdown()

	hlog.Info("服务器已关闭")
}

// loadConfig 加载配置文件
func loadConfig() error {
	// 尝试多个可能的配置文件路径
	configPaths := []string{
		"config.yaml",       // 当前目录
		"../../config.yaml", // 项目根目录（从 cmd/api 向上两级）
		"../config.yaml",    // 上级目录
	}

	var err error
	var lastErr error

	for _, path := range configPaths {
		cfg, err = config.LoadConfig(path)
		if err == nil {
			hlog.Infof("从 %s 加载配置文件成功", path)
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("无法找到配置文件，已尝试路径: %v, 最后错误: %w", configPaths, lastErr)
}

// initDatabase 初始化数据库连接
func initDatabase() error {
	// 设置日志级别
	var logLevel logger.LogLevel
	switch cfg.Database.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	// 创建数据库连接
	var err error
	db, err = gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层的 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.GetConnMaxLifetime())

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移（如果配置启用）
	if cfg.Database.AutoMigrate {
		hlog.Info("执行数据库自动迁移...")
		// 在这里添加需要迁移的模型
		// if err := db.AutoMigrate(&models.YourModel{}); err != nil {
		// 	return fmt.Errorf("数据库迁移失败: %w", err)
		// }
	}

	return nil
}

// initRedis 初始化Redis连接
func initRedis() error {
	rdb = redis.NewClient(&redis.Options{
		Addr:            cfg.Redis.GetAddr(),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		DialTimeout:     cfg.Redis.GetDialTimeout(),
		ReadTimeout:     cfg.Redis.GetReadTimeout(),
		WriteTimeout:    cfg.Redis.GetWriteTimeout(),
		PoolTimeout:     cfg.Redis.GetPoolTimeout(),
		ConnMaxIdleTime: cfg.Redis.GetIdleTimeout(),
		ConnMaxLifetime: cfg.Redis.GetMaxConnAge(),
	})

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	return nil
}

// initLongPolling 初始化长轮询服务
func initLongPolling() error {
	// 1. 创建Redis配置监听器
	configListener = infraListener.NewRedisConfigListener(rdb)

	// 2. 创建配置仓储（用于版本查询）
	configRepo := infraRepository.NewConfigRepository(db)

	// 3. 创建长轮询领域服务（超时60秒）
	longPollingService = domainService.NewLongPollingService(
		configListener,
		configRepo,
		60*time.Second,
	)

	// 4. 启动长轮询服务
	if err := longPollingService.Start(); err != nil {
		return fmt.Errorf("启动长轮询服务失败: %w", err)
	}

	return nil
}

// initServer 初始化HTTP服务器
func initServer() {
	// 创建Hertz实例
	hertzH = server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", cfg.Server.Port)),
	)

	// 注册全局中间件
	hertzH.Use(middleware.Recovery())

	// 注册路由
	registerRoutes()
}

// registerRoutes 注册路由
func registerRoutes() {
	// 健康检查
	hertzH.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "服务运行正常",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 数据库健康检查
	hertzH.GET("/health/db", func(c context.Context, ctx *app.RequestContext) {
		sqlDB, err := db.DB()
		if err != nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "获取数据库连接失败",
				"error":   err.Error(),
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "数据库连接失败",
				"error":   err.Error(),
			})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "数据库连接正常",
		})
	})

	// Redis健康检查
	hertzH.GET("/health/redis", func(c context.Context, ctx *app.RequestContext) {
		if err := rdb.Ping(c).Err(); err != nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Redis连接失败",
				"error":   err.Error(),
			})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "Redis连接正常",
		})
	})

	// 根路径
	hertzH.GET("/", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"message": "欢迎使用配置中心 API",
			"version": "1.0.0",
		})
	})

	// 注册配置管理路由（包含变更历史路由）
	registerConfigRoutes()
	hlog.Info("配置管理路由和变更历史路由注册成功")

	// 注册命名空间管理路由
	registerNamespaceRoutes()
	hlog.Info("命名空间管理路由注册成功")
}

// registerConfigRoutes 注册配置管理路由
func registerConfigRoutes() {
	// 初始化依赖层级：Repository -> DomainService -> AppService -> Handler

	// 1. 创建仓储层实例
	configRepo := infraRepository.NewConfigRepository(db)
	changeHistoryRepo := infraRepository.NewChangeHistoryRepository(db)

	// 2. 创建变更历史领域服务
	changeHistoryService := domainService.NewChangeHistoryService(changeHistoryRepo, configRepo, nil)

	// 3. 创建配置领域服务实例（传入配置监听器和变更历史服务）
	configDomainService := domainService.NewConfigService(configRepo, configListener, changeHistoryService)

	// 4. 更新变更历史服务的配置服务引用（用于回滚）
	changeHistoryService = domainService.NewChangeHistoryService(changeHistoryRepo, configRepo, configDomainService)

	// 5. 创建转换器实例
	configConverter := converter.NewConfigConverter()

	// 6. 创建应用服务实例
	configAppService := service.NewConfigAppService(configDomainService, configConverter)
	changeHistoryAppService := service.NewChangeHistoryAppService(changeHistoryService)

	// 7. 创建HTTP处理器实例
	configHandler := configHttp.NewConfigHandler(configAppService)
	changeHistoryHandler := configHttp.NewChangeHistoryHandler(changeHistoryAppService)

	// 8. 创建长轮询应用服务
	longPollingAppService := service.NewLongPollingAppService(longPollingService, configRepo)
	longPollingHandler := configHttp.NewLongPollingHandler(longPollingAppService)

	// 9. 注册路由
	api := hertzH.Group("/api/v1")
	{
		configs := api.Group("/configs")
		{
			configs.POST("", configHandler.CreateConfig)      // 创建配置
			configs.PUT("", configHandler.UpdateConfig)       // 更新配置（ID在请求体中）
			configs.GET("", configHandler.QueryConfigs)       // 分页查询配置
			configs.POST("/get", configHandler.GetConfigByID) // 根据ID获取配置（ID在请求体中）
			configs.DELETE("", configHandler.DeleteConfig)    // 删除配置（ID在请求体中）
			configs.POST("/watch", longPollingHandler.Watch)  // 长轮询监听配置变更
		}

		history := api.Group("/history")
		{
			history.GET("", changeHistoryHandler.QueryHistory)             // 分页查询变更历史
			history.POST("/get", changeHistoryHandler.GetHistoryByID)      // 根据ID查询变更记录（ID在请求体中）
			history.GET("/statistics", changeHistoryHandler.GetStatistics) // 获取变更统计
			history.GET("/config", changeHistoryHandler.GetConfigHistory)  // 获取配置变更历史
			history.POST("/compare", changeHistoryHandler.CompareVersions) // 对比版本
			history.POST("/rollback", changeHistoryHandler.Rollback)       // 回滚配置
		}
	}
}

// registerNamespaceRoutes 注册命名空间管理路由
func registerNamespaceRoutes() {
	// 初始化依赖层级：Repository -> DomainService -> AppService -> Handler

	// 1. 创建仓储层实例
	namespaceRepo := infraRepository.NewNamespaceRepository(db)
	configRepo := infraRepository.NewConfigRepository(db)

	// 2. 创建领域服务实例（传入配置监听器）
	namespaceDomainService := domainService.NewNamespaceService(namespaceRepo, configRepo)

	// 3. 创建转换器实例
	namespaceConverter := converter.NewNamespaceConverter()

	// 4. 创建应用服务实例
	namespaceAppService := service.NewNamespaceAppService(namespaceDomainService, namespaceConverter, namespaceRepo)

	// 5. 创建HTTP处理器实例
	namespaceHandler := configHttp.NewNamespaceHandler(namespaceAppService)

	// 6. 注册路由
	api := hertzH.Group("/api/v1")
	{
		namespaces := api.Group("/namespaces")
		{
			namespaces.POST("", namespaceHandler.CreateNamespace)                // 创建命名空间
			namespaces.PUT("", namespaceHandler.UpdateNamespace)                 // 更新命名空间（ID在请求体中）
			namespaces.DELETE("", namespaceHandler.DeleteNamespace)              // 删除命名空间（ID在请求体中）
			namespaces.PUT("/activate", namespaceHandler.ActivateNamespace)      // 激活命名空间（ID在请求体中）
			namespaces.PUT("/deactivate", namespaceHandler.DeactivateNamespace)  // 停用命名空间（ID在请求体中）
			namespaces.GET("", namespaceHandler.QueryNamespaces)                 // 分页查询命名空间
			namespaces.POST("/get", namespaceHandler.GetNamespaceByID)           // 根据ID获取命名空间（ID在请求体中）
			namespaces.GET("/name", namespaceHandler.GetNamespaceByName)         // 根据名称获取命名空间
			namespaces.GET("/active", namespaceHandler.GetActiveNamespace)       // 获取激活的命名空间
			namespaces.GET("/all", namespaceHandler.ListAllNamespaces)           // 获取所有命名空间（不分页）
			namespaces.GET("/active/all", namespaceHandler.ListActiveNamespaces) // 获取所有激活的命名空间（不分页）
		}
	}
}

// gracefulShutdown 优雅关闭
func gracefulShutdown() {
	// 关闭长轮询服务
	if longPollingService != nil {
		hlog.Info("正在关闭长轮询服务...")
		if err := longPollingService.Stop(); err != nil {
			hlog.Errorf("关闭长轮询服务失败: %v", err)
		}
	}

	// 关闭HTTP服务器
	if hertzH != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := hertzH.Shutdown(shutdownCtx); err != nil {
			hlog.Errorf("关闭HTTP服务器失败: %v", err)
		}
	}

	// 关闭数据库连接
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				hlog.Errorf("关闭数据库连接失败: %v", err)
			}
		}
	}

	// 关闭Redis连接
	if rdb != nil {
		if err := rdb.Close(); err != nil {
			hlog.Errorf("关闭Redis连接失败: %v", err)
		}
	}
}
