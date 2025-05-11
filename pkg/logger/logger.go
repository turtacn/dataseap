package logger

import (
	"os"
	"strings"
	"sync"

	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/common/types/enum"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger // 全局 logger 实例 Global logger instance
	once         sync.Once   // 用于确保 logger 初始化一次 Used to ensure logger initializes once
)

// Config 日志配置结构
// Config structure for logger configuration.
type Config struct {
	Level       string   `mapstructure:"level" json:"level" yaml:"level"`                   // 日志级别 (DEBUG, INFO, WARN, ERROR, FATAL) Log level
	Format      string   `mapstructure:"format" json:"format" yaml:"format"`                // 日志格式 (json, console) Log format
	OutputPaths []string `mapstructure:"outputPaths" json:"outputPaths" yaml:"outputPaths"` // 日志输出路径 (stdout, stderr, file paths) Log output paths
	ErrorPaths  []string `mapstructure:"errorPaths" json:"errorPaths" yaml:"errorPaths"`    // 错误日志输出路径 Error log output paths
	Development bool     `mapstructure:"development" json:"development" yaml:"development"` // 是否为开发模式 Whether it's development mode
}

// DefaultConfig 返回默认的日志配置
// DefaultConfig returns the default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:       "INFO",
		Format:      "console", // "json" or "console"
		OutputPaths: []string{"stdout"},
		ErrorPaths:  []string{"stderr"},
		Development: false,
	}
}

// InitGlobalLogger 初始化全局 logger
// InitGlobalLogger initializes the global logger.
// 它是线程安全的，且只会执行一次。
// It is thread-safe and will only execute once.
func InitGlobalLogger(cfg *Config) error {
	var err error
	once.Do(func() {
		var zapCfg zap.Config
		if cfg.Development {
			zapCfg = zap.NewDevelopmentConfig()
		} else {
			zapCfg = zap.NewProductionConfig()
		}

		// 设置日志级别
		// Set log level
		logLevel := enum.ParseLogLevel(cfg.Level)
		switch logLevel {
		case enum.LogLevelDebug:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		case enum.LogLevelInfo:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		case enum.LogLevelWarn:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		case enum.LogLevelError:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		case enum.LogLevelFatal:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
		default:
			zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		}

		// 设置日志格式
		// Set log format
		if strings.ToLower(cfg.Format) == "console" {
			zapCfg.Encoding = "console"
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 控制台彩色输出 Console color output
		} else {
			zapCfg.Encoding = "json"
		}

		// 设置输出路径
		// Set output paths
		if len(cfg.OutputPaths) > 0 {
			zapCfg.OutputPaths = cfg.OutputPaths
		} else {
			zapCfg.OutputPaths = []string{"stdout"}
		}

		if len(cfg.ErrorPaths) > 0 {
			zapCfg.ErrorOutputPaths = cfg.ErrorPaths
		} else {
			zapCfg.ErrorOutputPaths = []string{"stderr"}
		}

		// 构建 logger
		// Build logger
		l, buildErr := zapCfg.Build(zap.AddCallerSkip(1)) // AddCallerSkip(1) to show correct caller
		if buildErr != nil {
			err = buildErr
			// 如果构建失败，使用一个备用的简单 logger
			// If build fails, use a fallback simple logger
			l, _ = zap.NewProduction(zap.ErrorOutput(zapcore.Lock(os.Stderr)))
		}
		globalLogger = l.Named(constants.ServiceName) // 为 logger 添加服务名称前缀 Add service name prefix to logger
	})
	return err
}

// L 返回一个全局的 SugaredLogger 实例
// L returns a global SugaredLogger instance.
// 如果 logger 未初始化，它会使用默认配置进行初始化。
// If the logger is not initialized, it will be initialized with default configuration.
func L() *zap.SugaredLogger {
	if globalLogger == nil {
		// 在 logger 未被显式初始化时提供一个默认实例
		// Provide a default instance if logger is not explicitly initialized
		_ = InitGlobalLogger(DefaultConfig()) // 错误处理可以根据需要添加 Error handling can be added if needed
	}
	return globalLogger.Sugar()
}

// NamedL 返回一个带有指定名称的 SugaredLogger 实例
// NamedL returns a SugaredLogger instance with a specified name.
func NamedL(name string) *zap.SugaredLogger {
	if globalLogger == nil {
		_ = InitGlobalLogger(DefaultConfig())
	}
	return globalLogger.Named(name).Sugar()
}

// Sync刷新所有缓冲的日志条目。应用程序应在退出前调用此方法。
// Sync flushes any buffered log entries. Applications should call this before exiting.
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
