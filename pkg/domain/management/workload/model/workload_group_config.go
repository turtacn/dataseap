package model

// WorkloadGroup represents the configuration of a StarRocks workload group.
// WorkloadGroup 代表StarRocks工作负载组的配置。
type WorkloadGroup struct {
	// Name 工作负载组的唯一名称。
	// Name Unique name of the workload group.
	Name string `json:"name"`

	// CPUShare CPU资源权重。一个整数，表示相对CPU时间片。
	// CPUShare CPU resource weight. An integer representing relative CPU time slices.
	CPUShare int32 `json:"cpuShare"`

	// MemoryLimit 内存限制。例如 "10G", "20%" (相对于查询可用的总内存或BE节点内存)。
	// MemoryLimit Memory limit. E.g., "10G", "20%" (relative to total query-available memory or BE node memory).
	MemoryLimit string `json:"memoryLimit"`

	// ConcurrencyLimit (可选) 该工作负载组允许的最大并发查询数。
	// ConcurrencyLimit (Optional) Maximum number of concurrent queries allowed for this workload group.
	ConcurrencyLimit int32 `json:"concurrencyLimit,omitempty"`

	// MaxQueueSize (可选) 如果并发达到上限，允许排队等待执行的最大任务数。
	// MaxQueueSize (Optional) Maximum number of tasks allowed to queue if concurrency limit is reached.
	MaxQueueSize int32 `json:"maxQueueSize,omitempty"`

	// Properties (可选) 其他StarRocks特定的属性，键值对形式。
	// 例如: "spill_mem_limit_threshold": "0.8", "enable_memory_overcommit": "false"
	// Properties (Optional) Other StarRocks-specific properties in key-value format.
	// E.g., "spill_mem_limit_threshold": "0.8", "enable_memory_overcommit": "false"
	Properties map[string]string `json:"properties,omitempty"`
}

// Validate performs basic validation on the WorkloadGroup.
// Validate 对 WorkloadGroup 执行基本验证。
func (wg *WorkloadGroup) Validate() error {
	if wg.Name == "" {
		return NewDomainError("WorkloadGroup name cannot be empty")
	}
	if wg.CPUShare <= 0 {
		return NewDomainError("WorkloadGroup CPUShare must be positive")
	}
	if wg.MemoryLimit == "" { // Basic check, could be more sophisticated (e.g., parse percentage or size)
		return NewDomainError("WorkloadGroup MemoryLimit cannot be empty")
	}
	// Further validation for properties, limits, etc. can be added.
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
