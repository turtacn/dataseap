package workload

import (
	"context"

	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	"github.com/turtacn/dataseap/pkg/domain/management/workload/model"
)

// Service defines the interface for managing StarRocks workload groups.
// Service 定义了管理StarRocks工作负载组的接口。
type Service interface {
	// CreateWorkloadGroup creates a new workload group in StarRocks.
	// CreateWorkloadGroup 在StarRocks中创建一个新的工作负载组。
	CreateWorkloadGroup(ctx context.Context, group *model.WorkloadGroup) error

	// GetWorkloadGroup retrieves the configuration of a specific workload group.
	// GetWorkloadGroup 检索特定工作负载组的配置。
	GetWorkloadGroup(ctx context.Context, name string) (*model.WorkloadGroup, error)

	// ListWorkloadGroups lists all configured workload groups, with optional pagination.
	// ListWorkloadGroups 列出所有已配置的工作负载组，可选分页。
	ListWorkloadGroups(ctx context.Context, pagination *commontypes.PaginationRequest) (groups []*model.WorkloadGroup, total int64, err error)

	// UpdateWorkloadGroup updates the configuration of an existing workload group.
	// UpdateWorkloadGroup 更新现有工作负载组的配置。
	UpdateWorkloadGroup(ctx context.Context, group *model.WorkloadGroup) error

	// DeleteWorkloadGroup removes a workload group from StarRocks.
	// DeleteWorkloadGroup 从StarRocks中移除一个工作负载组。
	DeleteWorkloadGroup(ctx context.Context, name string) error
}
