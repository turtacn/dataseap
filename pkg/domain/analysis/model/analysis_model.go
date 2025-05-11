package model

import (
	"time"

	"github.com/turtacn/dataseap/pkg/common/types/enum"
)

// AnalysisTask represents a long-running analysis task in the system.
// AnalysisTask 代表系统中的一个长时间运行的分析任务。
type AnalysisTask struct {
	// ID 任务的唯一标识符
	// ID Unique identifier for the task.
	ID string `json:"id"`

	// Name (可选) 任务的可读名称
	// Name (Optional) Human-readable name for the task.
	Name string `json:"name,omitempty"`

	// TaskType 任务类型，例如 "historical_scan", "correlation_analysis", "model_training"
	// TaskType Type of the task, e.g., "historical_scan", "correlation_analysis", "model_training".
	TaskType string `json:"taskType"`

	// Status 任务的当前状态
	// Status Current status of the task.
	Status enum.TaskStatus `json:"status"`

	// Parameters 任务执行所需的参数，键值对形式
	// Parameters Parameters required for task execution, in key-value format.
	Parameters map[string]interface{} `json:"parameters"`

	// Priority (可选) 任务优先级
	// Priority (Optional) Task priority.
	Priority int `json:"priority,omitempty"` // e.g., 0 (lowest) to 10 (highest)

	// SubmittedBy (可选) 提交任务的用户或系统组件
	// SubmittedBy (Optional) User or system component that submitted the task.
	SubmittedBy string `json:"submittedBy,omitempty"`

	// CreatedAt 任务创建时间
	// CreatedAt Timestamp when the task was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt 任务状态最后更新时间
	// UpdatedAt Timestamp when the task status was last updated.
	UpdatedAt time.Time `json:"updatedAt"`

	// StartedAt (可选) 任务开始执行时间
	// StartedAt (Optional) Timestamp when the task started execution.
	StartedAt *time.Time `json:"startedAt,omitempty"`

	// CompletedAt (可选) 任务完成时间
	// CompletedAt (Optional) Timestamp when the task completed.
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	// Progress (可选) 任务执行进度 (例如 0.0 到 1.0)
	// Progress (Optional) Execution progress of the task (e.g., 0.0 to 1.0).
	Progress float32 `json:"progress,omitempty"`

	// Message (可选) 关于任务状态或结果的附加信息
	// Message (Optional) Additional message regarding the task's status or result.
	Message string `json:"message,omitempty"`

	// ResultURI (可选) 如果任务产生持久化结果，则为结果的存储位置标识符 (例如 S3 URI)
	// ResultURI (Optional) If the task produces persistent results, this is the identifier for their storage location (e.g., S3 URI).
	ResultURI string `json:"resultUri,omitempty"`

	// ErrorDetails (可选) 如果任务失败，记录错误详情
	// ErrorDetails (Optional) If the task failed, records error details.
	ErrorDetails string `json:"errorDetails,omitempty"`
}

// Validate performs basic validation on the AnalysisTask.
// Validate 对 AnalysisTask 执行基本验证。
func (at *AnalysisTask) ValidateForCreation() error {
	if at.TaskType == "" {
		return NewDomainError("TaskType cannot be empty")
	}
	// Parameters might be task-type specific, validation can be delegated.
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
