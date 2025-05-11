package model

import (
	"time"
)

// RawEvent represents a single data event to be ingested into the system.
// RawEvent 代表一条将被采集到系统中的原始数据事件。
type RawEvent struct {
	// ID (可选) 事件的唯一标识符，如果数据源提供
	// ID (Optional) Unique identifier for the event, if provided by the data source.
	ID string `json:"id,omitempty"`

	// DataSourceID 数据来源标识，例如探针ID, 日志文件名等
	// DataSourceID Identifier for the data source, e.g., probe ID, log file name.
	DataSourceID string `json:"dataSourceId"`

	// DataType 数据类型，用于指导解析和存储，例如 "firewall_log", "edr_event"
	// DataType Type of data, used to guide parsing and storage, e.g., "firewall_log", "edr_event".
	DataType string `json:"dataType"`

	// Timestamp 事件发生的时间戳
	// Timestamp Timestamp of when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Data 事件的具体内容，通常是map[string]interface{}形式的结构化数据
	// Data The actual content of the event, typically structured data in map[string]interface{} form.
	Data map[string]interface{} `json:"data"`

	// RawPayload (可选) 原始的字节负载，如果Data字段无法完全表达或需要原始数据时使用
	// RawPayload (Optional) The original byte payload, used if the Data field cannot fully represent it or if raw data is needed.
	RawPayload []byte `json:"rawPayload,omitempty"`

	// Tags 附加的标签，用于分类或路由
	// Tags Additional tags for classification or routing.
	Tags map[string]string `json:"tags,omitempty"`

	// ReceivedAt (可选) DataSeaP平台接收到此事件的时间戳
	// ReceivedAt (Optional) Timestamp when DataSeaP platform received this event.
	ReceivedAt time.Time `json:"receivedAt,omitempty"`
}

// Validate performs basic validation on the RawEvent.
// Validate 对 RawEvent 执行基本验证。
func (re *RawEvent) Validate() error {
	if re.DataSourceID == "" {
		return NewDomainError("DataSourceID cannot be empty")
	}
	if re.DataType == "" {
		return NewDomainError("DataType cannot be empty")
	}
	if re.Timestamp.IsZero() {
		return NewDomainError("Timestamp cannot be zero")
	}
	if re.Data == nil && len(re.RawPayload) == 0 {
		return NewDomainError("Either Data or RawPayload must be provided")
	}
	return nil
}

// DomainError represents an error specific to the domain logic.
// DomainError 代表领域逻辑相关的错误。
type DomainError struct {
	message string
}

// NewDomainError creates a new DomainError.
// NewDomainError 创建一个新的 DomainError。
func NewDomainError(message string) *DomainError {
	return &DomainError{message: message}
}

// Error implements the error interface.
// Error 实现 error 接口。
func (e *DomainError) Error() string {
	return e.message
}
