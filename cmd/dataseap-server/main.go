package main

import (
	"fmt"
	"os"

	"github.com/turtacn/dataseap/internal/app"
	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger" // 确保logger被导入以便其init或全局变量能工作 Ensure logger is imported so its init or globals can work
)

func main() {
	// 0. 打印版本信息 (可选)
	// 0. Print version information (optional)
	fmt.Printf("%s version %s\n", constants.ServiceName, constants.ServiceVersion)

	// 1. 加载配置
	// 1. Load configuration
	// 可以通过命令行参数、环境变量等方式指定配置文件路径
	// Config file path can be specified via command line args, environment variables, etc.
	// configFilePath := os.Getenv("DATASEAP_CONFIG_PATH")
	cfg, err := config.LoadConfig() // LoadConfig will use default path if no arg given
	if err != nil {
		// 如果配置加载失败，无法初始化logger, 使用fmt打印到stderr
		// If config loading fails, logger cannot be initialized, use fmt to print to stderr
		fmt.Fprintf(os.Stderr, "FATAL: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化Logger (config.LoadConfig 内部已包含logger.InitGlobalLogger的调用，如果配置存在logger部分)
	// 2. Initialize Logger (config.LoadConfig internally calls logger.InitGlobalLogger if logger section exists in config)
	// 此处调用 L() 确保logger被激活，并获取实例
	// Calling L() here ensures logger is active and gets an instance
	l := logger.L()
	l.Info("Configuration loaded successfully.")

	// 3. 创建和引导应用程序
	// 3. Create and bootstrap the application
	application, err := app.NewApplication(cfg)
	if err != nil {
		l.Fatalf("Failed to bootstrap application: %v", err)
	}

	// 4. 启动应用程序
	// 4. Start the application
	// Start方法内部会阻塞，直到接收到关闭信号，并执行优雅关闭
	// The Start method will block until a shutdown signal is received and then perform graceful shutdown.
	if err := application.Start(); err != nil {
		l.Fatalf("Application exited with error: %v", err)
	}

	l.Info("Application has shut down gracefully.")
}
