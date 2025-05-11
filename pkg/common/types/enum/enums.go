package enum

import "strings"

// LogLevel 日志级别枚举
// LogLevel enum for logging levels.
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG" // 调试级别 Debug level
	LogLevelInfo  LogLevel = "INFO"  // 信息级别 Info level
	LogLevelWarn  LogLevel = "WARN"  // 警告级别 Warning level
	LogLevelError LogLevel = "ERROR" // 错误级别 Error level
	LogLevelFatal LogLevel = "FATAL" // 致命级别 Fatal level
)

// String 实现 Stringer 接口，方便打印
// String implements the Stringer interface for easy printing.
func (ll LogLevel) String() string {
	return string(ll)
}

// ParseLogLevel 从字符串解析 LogLevel，如果无效则返回默认值 LogLevelInfo
// ParseLogLevel parses LogLevel from a string, returns LogLevelInfo if invalid.
func ParseLogLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "FATAL":
		return LogLevelFatal
	default:
		return LogLevelInfo // 默认级别 Default level
	}
}

// WorkloadPriority 工作负载优先级枚举
// WorkloadPriority enum for workload priorities.
type WorkloadPriority string

const (
	WorkloadPriorityLowest  WorkloadPriority = "LOWEST"  // 最低优先级 Lowest priority
	WorkloadPriorityLow     WorkloadPriority = "LOW"     // 低优先级 Low priority
	WorkloadPriorityNormal  WorkloadPriority = "NORMAL"  // 普通优先级 Normal priority
	WorkloadPriorityHigh    WorkloadPriority = "HIGH"    // 高优先级 High priority
	WorkloadPriorityHighest WorkloadPriority = "HIGHEST" // 最高优先级 Highest priority
)

// String 实现 Stringer 接口
// String implements the Stringer interface.
func (wp WorkloadPriority) String() string {
	return string(wp)
}

// IndexType 索引类型枚举 (StarRocks相关)
// IndexType enum for index types (StarRocks related).
type IndexType string

const (
	IndexTypeBitmap   IndexType = "BITMAP"   // 位图索引 Bitmap index
	IndexTypeInverted IndexType = "INVERTED" // 倒排索引 Inverted index
	// 可根据StarRocks支持情况添加其他索引类型
	// Other index types can be added based on StarRocks support
)

// String 实现 Stringer 接口
// String implements the Stringer interface.
func (it IndexType) String() string {
	return string(it)
}

// TokenizerType 分词器类型枚举 (StarRocks倒排索引相关)
// TokenizerType enum for tokenizer types (StarRocks inverted index related).
type TokenizerType string

const (
	TokenizerTypeDefault  TokenizerType = "default"  // 默认分词器 Default tokenizer (usually standard)
	TokenizerTypeStandard TokenizerType = "standard" // 标准分词器 Standard tokenizer
	TokenizerTypeEnglish  TokenizerType = "english"  // 英文分词器 English tokenizer
	TokenizerTypeChinese  TokenizerType = "chinese"  // 中文分词器 (例如
)

// String 实现 Stringer 接口
// String implements the Stringer interface.
func (tt TokenizerType) String() string {
	return string(tt)
}

// TaskStatus 任务状态枚举
// TaskStatus enum for task statuses.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"   // 待处理 Pending
	TaskStatusRunning   TaskStatus = "RUNNING"   // 运行中 Running
	TaskStatusSuccess   TaskStatus = "SUCCESS"   // 成功 Success
	TaskStatusFailed    TaskStatus = "FAILED"    // 失败 Failed
	TaskStatusCancelled TaskStatus = "CANCELLED" // 已取消 Cancelled
)

// String 实现 Stringer 接口
// String implements the Stringer interface.
func (ts TaskStatus) String() string {
	return string(ts)
}

// DataType 数据类型枚举 (用于元数据管理)
// DataType enum for data types (for metadata management).
type DataType string

const (
	DataTypeUnknown  DataType = "UNKNOWN"  // 未知类型 Unknown type
	DataTypeBoolean  DataType = "BOOLEAN"  //布尔类型 Boolean type
	DataTypeTinyInt  DataType = "TINYINT"  // TinyInt 类型 TinyInt type
	DataTypeSmallInt DataType = "SMALLINT" // SmallInt 类型 SmallInt type
	DataTypeInt      DataType = "INT"      // Int 类型 Int type
	DataTypeBigInt   DataType = "BIGINT"   // BigInt 类型 BigInt type
	DataTypeLargeInt DataType = "LARGEINT" // LargeInt 类型 LargeInt type
	DataTypeFloat    DataType = "FLOAT"    // Float 类型 Float type
	DataTypeDouble   DataType = "DOUBLE"   // Double 类型 Double type
	DataTypeDecimal  DataType = "DECIMAL"  // Decimal 类型 Decimal type
	DataTypeDate     DataType = "DATE"     // Date 类型 Date type
	DataTypeDateTime DataType = "DATETIME" // DateTime 类型 DateTime type
	DataTypeChar     DataType = "CHAR"     // Char 类型 Char type
	DataTypeVarchar  DataType = "VARCHAR"  // Varchar 类型 Varchar type
	DataTypeString   DataType = "STRING"   // String 类型 String type
	DataTypeJSON     DataType = "JSON"     // JSON类型 JSON type
	DataTypeArray    DataType = "ARRAY"    // 数组类型 Array type
	DataTypeMap      DataType = "MAP"      // Map类型 Map type
	DataTypeStruct   DataType = "STRUCT"   // 结构体类型 Struct type
)

// String 实现 Stringer 接口
// String implements the Stringer interface.
func (dt DataType) String() string {
	return string(dt)
}
