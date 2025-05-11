package constants

// ServiceName 服务名称
// ServiceName is the name of the service.
const ServiceName = "dataseap-server"

// ServiceVersion 服务版本 (可以使用 ldflags 在构建时注入)
// ServiceVersion is the version of the service (can be injected at build time using ldflags).
var ServiceVersion = "0.1.0-dev"

// DefaultAPIPort API服务的默认端口
// DefaultAPIPort is the default port for the API service.
const DefaultAPIPort = 8080

// DefaultGRPCPort gRPC服务的默认端口
// DefaultGRPCPort is the default port for the gRPC service.
const DefaultGRPCPort = 50051

// DefaultConfigPath 默认配置文件路径
// DefaultConfigPath is the default path for the configuration file.
const DefaultConfigPath = "./config/config.yaml"

// ContextKey 自定义Context键类型，以避免冲突
// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// ContextKeyRequestID 用于在Context中存储请求ID
	// ContextKeyRequestID is used to store the request ID in the Context.
	ContextKeyRequestID ContextKey = "request_id"

	// ContextKeyLogger 用于在Context中存储logger实例
	// ContextKeyLogger is used to store the logger instance in the Context.
	ContextKeyLogger ContextKey = "logger"

	// ContextKeyUser 用于在Context中存储用户信息
	// ContextKeyUser is used to store user information in the Context.
	ContextKeyUser ContextKey = "user"
)

// DefaultPageSize 默认分页大小
// DefaultPageSize is the default size for pagination.
const DefaultPageSize = 10

// MaxPageSize 最大分页大小
// MaxPageSize is the maximum size for pagination.
const MaxPageSize = 100

// DefaultTimeFormat 默认时间格式
// DefaultTimeFormat is the default time format used in the application.
const DefaultTimeFormat = "2006-01-02 15:04:05"

// DefaultTimeZone 默认时区
// DefaultTimeZone is the default time zone used in the application.
const DefaultTimeZone = "UTC"

// StarRocksDefaultQueryTimeoutStarRocks默认查询超时时间（秒）
// StarRocksDefaultQueryTimeout is the default query timeout for StarRocks in seconds.
const StarRocksDefaultQueryTimeout = 30

// StarRocksDefaultConnectTimeout StarRocks默认连接超时时间（秒）
// StarRocksDefaultConnectTimeout is the default connection timeout for StarRocks in seconds.
const StarRocksDefaultConnectTimeout = 5

// PulsarDefaultOperationTimeout Pulsar默认操作超时时间（秒）
// PulsarDefaultOperationTimeout is the default operation timeout for Pulsar in seconds.
const PulsarDefaultOperationTimeout = 10

// HeaderRequestID HTTP头部中用于追踪请求ID的键名
// HeaderRequestID is the key name in HTTP headers for tracing request ID.
const HeaderRequestID = "X-Request-ID"
