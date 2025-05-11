package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/logger"
)

// Config 应用配置结构体
// Config is the application configuration structure.
type Config struct {
	Server    ServerConfig    `mapstructure:"server" json:"server" yaml:"server"`
	Logger    logger.Config   `mapstructure:"logger" json:"logger" yaml:"logger"`
	StarRocks StarRocksConfig `mapstructure:"starrocks" json:"starrocks" yaml:"starrocks"`
	Pulsar    PulsarConfig    `mapstructure:"pulsar" json:"pulsar" yaml:"pulsar"`
	// 可以添加其他配置项，例如数据库、缓存等
	// Other configurations like database, cache can be added here
}

// ServerConfig 服务器相关配置
// ServerConfig holds server-related configurations.
type ServerConfig struct {
	Host           string `mapstructure:"host" json:"host" yaml:"host"`
	Port           int    `mapstructure:"port" json:"port" yaml:"port"`
	GRPCPort       int    `mapstructure:"grpcPort" json:"grpcPort" yaml:"grpcPort"`
	Mode           string `mapstructure:"mode" json:"mode" yaml:"mode"`                         // "debug", "release", "test"
	ReadTimeout    int    `mapstructure:"readTimeout" json:"readTimeout" yaml:"readTimeout"`    // 秒 seconds
	WriteTimeout   int    `mapstructure:"writeTimeout" json:"writeTimeout" yaml:"writeTimeout"` // 秒 seconds
	MaxHeaderBytes int    `mapstructure:"maxHeaderBytes" json:"maxHeaderBytes" yaml:"maxHeaderBytes"`
}

// StarRocksConfig StarRocks数据库配置
// StarRocksConfig holds StarRocks database configurations.
type StarRocksConfig struct {
	Hosts          []string `mapstructure:"hosts" json:"hosts" yaml:"hosts"`             // FE hosts, e.g., ["fe_host1:http_port", "fe_host2:http_port"]
	QueryPort      int      `mapstructure:"queryPort" json:"queryPort" yaml:"queryPort"` // 通常是FE的HTTP端口 Typically FE's HTTP port
	User           string   `mapstructure:"user" json:"user" yaml:"user"`
	Password       string   `mapstructure:"password" json:"password" yaml:"password"`
	Database       string   `mapstructure:"database" json:"database" yaml:"database"`
	ConnectTimeout int      `mapstructure:"connectTimeout" json:"connectTimeout" yaml:"connectTimeout"` // 秒 seconds
	QueryTimeout   int      `mapstructure:"queryTimeout" json:"queryTimeout" yaml:"queryTimeout"`       // 秒 seconds
	LoadURL        string   `mapstructure:"loadUrl" json:"loadUrl" yaml:"loadUrl"`                      // e.g. "fe_host1:http_port;fe_host2:http_port" for stream load
}

// PulsarConfig Pulsar消息队列配置
// PulsarConfig holds Pulsar message queue configurations.
type PulsarConfig struct {
	ServiceURL       string `mapstructure:"serviceUrl" json:"serviceUrl" yaml:"serviceUrl"`                   // e.g., "pulsar://localhost:6650"
	OperationTimeout int    `mapstructure:"operationTimeout" json:"operationTimeout" yaml:"operationTimeout"` // 秒 seconds
	// 可以添加更多Pulsar特定的配置，如TLS, Auth等
	// More Pulsar specific configurations like TLS, Auth can be added
}

var (
	globalConfig *Config
	configOnce   sync.Once
)

// LoadConfig 加载配置信息
// LoadConfig loads configuration from file and environment variables.
// filePath: 配置文件路径 (可选，如果为空则尝试默认路径或只从环境变量加载)
// filePath: Path to the config file (optional, if empty, tries default path or loads only from env).
func LoadConfig(filePath ...string) (*Config, error) {
	var err error
	configOnce.Do(func() {
		v := viper.New()

		// 设置默认值
		// Set default values
		v.SetDefault("server.host", "0.0.0.0")
		v.SetDefault("server.port", constants.DefaultAPIPort)
		v.SetDefault("server.grpcPort", constants.DefaultGRPCPort)
		v.SetDefault("server.mode", "debug")
		v.SetDefault("server.readTimeout", 30)       // 30 seconds
		v.SetDefault("server.writeTimeout", 30)      // 30 seconds
		v.SetDefault("server.maxHeaderBytes", 1<<20) // 1MB

		defaultLoggerCfg := logger.DefaultConfig()
		v.SetDefault("logger.level", defaultLoggerCfg.Level)
		v.SetDefault("logger.format", defaultLoggerCfg.Format)
		v.SetDefault("logger.outputPaths", defaultLoggerCfg.OutputPaths)
		v.SetDefault("logger.errorPaths", defaultLoggerCfg.ErrorPaths)
		v.SetDefault("logger.development", defaultLoggerCfg.Development)

		v.SetDefault("starrocks.queryPort", 8030) // Common StarRocks FE HTTP port
		v.SetDefault("starrocks.user", "root")
		v.SetDefault("starrocks.password", "")
		v.SetDefault("starrocks.connectTimeout", constants.StarRocksDefaultConnectTimeout)
		v.SetDefault("starrocks.queryTimeout", constants.StarRocksDefaultQueryTimeout)

		v.SetDefault("pulsar.operationTimeout", constants.PulsarDefaultOperationTimeout)

		// 设置配置文件路径和类型
		// Set config file path and type
		if len(filePath) > 0 && filePath[0] != "" {
			v.SetConfigFile(filePath[0])
		} else {
			v.SetConfigFile(constants.DefaultConfigPath) // 使用常量中定义的默认路径 Use default path defined in constants
		}
		v.SetConfigType("yaml") // 或者 "json", "toml" etc. or "json", "toml" etc.

		// 读取配置文件
		// Read config file
		if errRead := v.ReadInConfig(); errRead != nil {
			if _, ok := errRead.(viper.ConfigFileNotFoundError); ok {
				// 配置文件未找到，可以忽略，因为我们有环境变量和默认值
				// Config file not found; ignore error if we have env vars and defaults
				fmt.Printf("Config file not found or not specified, using defaults and environment variables. Path tried: %s\n", v.ConfigFileUsed())
			} else {
				// 配置文件被找到但解析错误
				// Config file was found but another error was produced
				err = fmt.Errorf("failed to read config file: %s, error: %w", v.ConfigFileUsed(), errRead)
				return
			}
		} else {
			fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
		}

		// 启用环境变量覆盖 (前缀为 DATASERP, 例如 DATASERP_SERVER_PORT)
		// Enable environment variable overriding (prefix DATASERP, e.g., DATASERP_SERVER_PORT)
		v.SetEnvPrefix(strings.ToUpper(constants.ServiceName))
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // SERVER_PORT instead of SERVER.PORT

		// 反序列化配置到结构体
		// Unmarshal config into struct
		var cfg Config
		if errUnmarshal := v.Unmarshal(&cfg); errUnmarshal != nil {
			err = fmt.Errorf("failed to unmarshal config: %w", errUnmarshal)
			return
		}
		globalConfig = &cfg
	})

	if err != nil {
		return nil, err
	}
	if globalConfig == nil && err == nil { // Should not happen with once.Do if no error occurred.
		return nil, fmt.Errorf("configuration was not loaded but no error reported")
	}
	return globalConfig, nil
}

// GetConfig 返回已加载的全局配置实例
// GetConfig returns the loaded global configuration instance.
// 如果配置未加载，它会尝试使用默认路径加载。
// If config is not loaded, it will try to load with default path.
func GetConfig() *Config {
	if globalConfig == nil {
		// 尝试加载默认配置，忽略错误以简化调用，但实际应用中可能需要处理
		// Try to load default config, ignore error for simple call, but real app might need to handle it
		_, _ = LoadConfig()
		if globalConfig == nil {
			// 如果仍然为nil（例如，默认配置文件也没有，且LoadConfig内部发生错误被忽略了）
			// If still nil (e.g. default config also not present and LoadConfig internal error was suppressed)
			// 返回一个包含默认值的最小配置，以防止nil指针
			// Return a minimal config with defaults to prevent nil pointers
			return &Config{
				Server: ServerConfig{
					Port: constants.DefaultAPIPort, GRPCPort: constants.DefaultGRPCPort, Mode: "debug",
				},
				Logger: *logger.DefaultConfig(),
			}
		}
	}
	return globalConfig
}
