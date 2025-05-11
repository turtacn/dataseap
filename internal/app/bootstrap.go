package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger"
	// 导入其他需要的包，例如 transport, domain, adapter 等
	// Import other necessary packages like transport, domain, adapter etc.
	// "github.com/turtacn/dataseap/pkg/transport/grpc"
	// "github.com/turtacn/dataseap/pkg/transport/http"
)

// Application 应用结构体，包含所有主要组件
// Application struct holds all major components of the application.
type Application struct {
	Cfg    *config.Config
	Logger *logger.Config // Logger config, actual logger is global via logger.L()
	// GRPCServer *grpc.Server // gRPC 服务器实例 gRPC server instance
	// HTTPServer *http.Server // HTTP 服务器实例 HTTP server instance
	// ... 其他组件，如数据库连接、消息队列客户端等
	// ... other components like database connections, message queue clients etc.
	shutdownFuncs []func(ctx context.Context) error // 优雅关闭时需要执行的函数列表 List of functions to execute on graceful shutdown
}

// NewApplication 创建并初始化一个新的 Application 实例
// NewApplication creates and initializes a new Application instance.
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		Cfg:    cfg,
		Logger: &cfg.Logger, // Store logger config for reference
	}

	// 1. 初始化 Logger (已经通过 config 加载时完成，这里确认或重新应用)
	// 1. Initialize Logger (already done during config loading, confirm or re-apply here)
	if err := logger.InitGlobalLogger(&cfg.Logger); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	l := logger.L()
	l.Info("Logger initialized successfully.")

	// 2. 初始化数据库连接 (例如 StarRocks)
	// 2. Initialize database connections (e.g., StarRocks)
	// starrocksClient, err := adapter.NewStarRocksClient(cfg.StarRocks)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to initialize StarRocks client: %w", err)
	// }
	// app.AddShutdownFunc(func(ctx context.Context) error { return starrocksClient.Close() })
	// l.Info("StarRocks client initialized (placeholder).")

	// 3. 初始化消息队列 (例如 Pulsar)
	// 3. Initialize message queue (e.g., Pulsar)
	// pulsarClient, err := adapter.NewPulsarClient(cfg.Pulsar)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to initialize Pulsar client: %w", err)
	// }
	// app.AddShutdownFunc(func(ctx context.Context) error { pulsarClient.Close(); return nil })
	// l.Info("Pulsar client initialized (placeholder).")

	// 4. 初始化领域服务 (Domain Services)
	// 4. Initialize Domain Services
	// ingestionService := ingestion.NewService(...)
	// queryService := query.NewService(...)
	// l.Info("Domain services initialized (placeholder).")

	// 5. 初始化传输层 (gRPC, HTTP 服务器)
	// 5. Initialize Transport Layer (gRPC, HTTP servers)
	// grpcServer, err := grpc.NewServer(cfg.Server, queryService, ingestionService, ...)
	// if err != nil {
	//    return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	// }
	// app.GRPCServer = grpcServer
	// app.AddShutdownFunc(grpcServer.GracefulStop)
	// l.Info("gRPC server initialized (placeholder).")

	// httpServer, err := http.NewServer(cfg.Server, queryService, ingestionService, ...)
	// if err != nil {
	//    return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	// }
	// app.HTTPServer = httpServer
	// app.AddShutdownFunc(httpServer.Shutdown) // Assuming http server has a Shutdown method
	// l.Info("HTTP server initialized (placeholder).")

	l.Info("Application bootstrapped successfully (placeholders for actual components).")
	return app, nil
}

// AddShutdownFunc 添加一个在程序关闭时需要调用的清理函数
// AddShutdownFunc adds a cleanup function to be called when the application shuts down.
func (app *Application) AddShutdownFunc(f func(ctx context.Context) error) {
	app.shutdownFuncs = append(app.shutdownFuncs, f)
}

// Start 启动应用程序的所有服务
// Start starts all services of the application.
func (app *Application) Start() error {
	l := logger.L()
	// 启动 gRPC 服务器
	// Start gRPC server
	// go func() {
	// 	l.Infof("gRPC server starting on port %d", app.Cfg.Server.GRPCPort)
	// 	if err := app.GRPCServer.ListenAndServe(); err != nil {
	// 		l.Fatalf("Failed to start gRPC server: %v", err)
	// 	}
	// }()

	// 启动 HTTP 服务器
	// Start HTTP server
	// go func() {
	//  l.Infof("HTTP server starting on port %d", app.Cfg.Server.Port)
	// 	if err := app.HTTPServer.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
	// 		l.Fatalf("Failed to start HTTP server: %v", err)
	// 	}
	// }()

	l.Info("Application services started (placeholders). Waiting for interrupt signal.")
	// 等待中断信号
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down application...")
	return app.Shutdown()
}

// Shutdown 优雅地关闭应用程序
// Shutdown gracefully shuts down the application.
func (app *Application) Shutdown() error {
	l := logger.L()
	l.Info("Executing shutdown functions...")

	// 创建一个带有超时的上下文用于关闭操作
	// Create a context with timeout for shutdown operations
	// TODO: Make timeout configurable
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 逆序执行关闭函数，确保依赖关系正确处理
	// Execute shutdown functions in reverse order to handle dependencies correctly
	for i := len(app.shutdownFuncs) - 1; i >= 0; i-- {
		f := app.shutdownFuncs[i]
		if err := f(shutdownCtx); err != nil {
			l.Errorf("Error during shutdown function: %v", err)
			// 可以选择继续执行其他关闭函数，或者直接返回错误
			// Optionally continue with other shutdown functions or return error directly
		}
	}

	// 确保日志被刷新
	// Ensure logs are flushed
	if err := logger.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Error syncing logger: %v\n", err) // Use fmt if logger itself fails
	}

	l.Info("Application shutdown complete.")
	return nil
}
