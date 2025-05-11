package analysis

import (
	"context"

	"github.com/turtacn/dataseap/pkg/common/types/enum"
	"github.com/turtacn/dataseap/pkg/domain/analysis/model"
)

// Service defines the interface for the data analysis service.
// Service 定义了数据分析服务的接口。
type Service interface {
	// SubmitTask submits a new analysis task for execution.
	// SubmitTask 提交一个新的分析任务以供执行。
	// It returns the ID of the submitted task.
	// 它返回已提交任务的ID。
	SubmitTask(ctx context.Context, taskDetails *model.AnalysisTask) (taskID string, err error)

	// GetTask retrieves the details of a specific analysis task.
	// GetTask 检索特定分析任务的详细信息。
	GetTask(ctx context.Context, taskID string) (*model.AnalysisTask, error)

	// GetTaskStatus retrieves the current status of a specific analysis task.
	// GetTaskStatus 检索特定分析任务的当前状态。
	GetTaskStatus(ctx context.Context, taskID string) (enum.TaskStatus, error)

	// CancelTask requests cancellation of a running or pending analysis task.
	// CancelTask 请求取消正在运行或待处理的分析任务。
	CancelTask(ctx context.Context, taskID string) error

	// ListTasks lists analysis tasks, possibly with filtering and pagination.
	// ListTasks 列出分析任务，可能带有过滤和分页。
	// TODO: Define ListTasksRequest and ListTasksResponse models if pagination/filtering is needed.
	// ListTasks(ctx context.Context, request *model.ListTasksRequest) (*model.ListTasksResponse, error)

	// GetTaskResult retrieves the result of a completed analysis task.
	// GetTaskResult 检索已完成分析任务的结果。
	// The nature of the result is task-dependent and might require type assertion or a generic wrapper.
	// 结果的性质取决于具体任务，可能需要类型断言或通用包装器。
	// For tasks with large results, this might return a URI or stream.
	// 对于结果较大的任务，此方法可能返回URI或流。
	GetTaskResult(ctx context.Context, taskID string) (result interface{}, err error)
}
