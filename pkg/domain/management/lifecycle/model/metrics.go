package model

import (
	"time"
)

// ComponentType 表示受管组件的类型
// ComponentType represents the type of a managed component.
type ComponentType string

const (
	ComponentTypeStarRocksFE  ComponentType = "StarRocksFE"  // StarRocks Frontend
	ComponentTypeStarRocksBE  ComponentType = "StarRocksBE"  // StarRocks Backend
	ComponentTypePulsarBroker ComponentType = "PulsarBroker" // Pulsar Broker
	ComponentTypePulsarBookie ComponentType = "PulsarBookie" // Pulsar Bookie
	ComponentTypeDataSeaP     ComponentType = "DataSeaP"     // DataSeaP Service itself
	ComponentTypeUnknown      ComponentType = "Unknown"      // 未知类型 Unknown type
)

// ComponentStatus represents the operational status of a managed component.
// ComponentStatus 代表受管组件的运行状态。
type ComponentStatus struct {
	// ComponentName 组件的唯一名称或ID (例如 "StarRocksFE-1", "dataseap-ingestion-pod-xyz")
	// ComponentName Unique name or ID of the component (e.g., "StarRocksFE-1", "dataseap-ingestion-pod-xyz").
	ComponentName string `json:"componentName"`

	// ComponentType 组件类型。
	// ComponentType Type of the component.
	ComponentType ComponentType `json:"componentType"`

	// Status 组件的健康状态 (例如 "HEALTHY", "UNHEALTHY", "DEGRADED", "INITIALIZING", "UNKNOWN")。
	// Status Health status of the component (e.g., "HEALTHY", "UNHEALTHY", "DEGRADED", "INITIALIZING", "UNKNOWN").
	// Consider using an enum for fixed status values.
	Status string `json:"status"`

	// Message (可选) 关于状态的附加可读信息。
	// Message (Optional) Additional human-readable information about the status.
	Message string `json:"message,omitempty"`

	// Details (可选) 更详细的状态信息或指标，键值对形式。
	// Details (Optional) More detailed status information or metrics, in key-value format.
	Details map[string]interface{} `json:"details,omitempty"`

	// LastCheckedAt 上次检查状态的时间戳。
	// LastCheckedAt Timestamp of the last status check.
	LastCheckedAt time.Time `json:"lastCheckedAt"`

	// Version (可选) 组件的版本信息。
	// Version (Optional) Version information of the component.
	Version string `json:"version,omitempty"`
}

// SystemMetric represents a single metric point for a system or component.
// SystemMetric 代表系统或组件的单个指标点。
type SystemMetric struct {
	// Name 指标的名称 (例如 "cpu_usage_percent", "memory_free_bytes", "query_latency_p99_ms")。
	// Name Name of the metric (e.g., "cpu_usage_percent", "memory_free_bytes", "query_latency_p99_ms").
	Name string `json:"name"`

	// Value 指标的当前值。
	// Value Current value of the metric.
	Value float64 `json:"value"`

	// Labels (可选) 指标的标签，用于区分不同的维度 (例如 {"host":"server1", "component":"FE"})。
	// Labels (Optional) Labels for the metric to distinguish different dimensions (e.g., {"host":"server1", "component":"FE"}).
	Labels map[string]string `json:"labels,omitempty"`

	// Timestamp 指标记录的时间戳。
	// Timestamp Timestamp when the metric was recorded.
	Timestamp time.Time `json:"timestamp"`

	// Unit (可选) 指标的单位 (例如 "%", "bytes", "ms")。
	// Unit (Optional) Unit of the metric (e.g., "%", "bytes", "ms").
	Unit string `json:"unit,omitempty"`
}

// MetricsQueryRequest represents a request to query system or component metrics.
// MetricsQueryRequest 代表查询系统或组件指标的请求。
type MetricsQueryRequest struct {
	// MetricNames (可选) 要查询的指标名称列表。如果为空，可能表示查询所有可用指标或特定类型的指标。
	// MetricNames (Optional) List of metric names to query. If empty, might mean all available or specific type.
	MetricNames []string `json:"metricNames,omitempty"`

	// ComponentName (可选) 筛选特定组件的指标。
	// ComponentName (Optional) Filter metrics for a specific component.
	ComponentName string `json:"componentName,omitempty"`

	// ComponentType (可选) 筛选特定类型组件的指标。
	// ComponentType (Optional) Filter metrics for a specific type of component.
	ComponentType ComponentType `json:"componentType,omitempty"`

	// Labels (可选) 进一步根据标签筛选指标。
	// Labels (Optional) Further filter metrics by labels.
	Labels map[string]string `json:"labels,omitempty"`

	// TimeRange (可选) 查询的时间范围。
	// TimeRange (Optional) Time range for the query.
	TimeRange *TimeRange `json:"timeRange,omitempty"` // Assuming TimeRange is defined in common types or here
}

// TimeRange defines a time range. (Duplicated from common/types for clarity if not imported)
// TimeRange 定义时间范围。 (为清晰起见，如果未导入，则从 common/types 复制)
type TimeRange struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
