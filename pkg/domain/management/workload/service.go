package workload

import (
	"context"

	"github.com/turtacn/dataseap/pkg/adapter/starrocks"
	"github.com/turtacn/dataseap/pkg/common/errors"
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	"github.com/turtacn/dataseap/pkg/domain/management/workload/model"
	"github.com/turtacn/dataseap/pkg/logger"
)

type serviceImpl struct {
	srDDLExecutor starrocks.DDLExecutor
}

// NewService creates a new instance of the workload management service.
// NewService 创建一个新的工作负载管理服务实例。
func NewService(ddlExecutor starrocks.DDLExecutor) Service {
	return &serviceImpl{
		srDDLExecutor: ddlExecutor,
	}
}

// CreateWorkloadGroup creates a new workload group in StarRocks.
// CreateWorkloadGroup 在StarRocks中创建一个新的工作负载组。
func (s *serviceImpl) CreateWorkloadGroup(ctx context.Context, group *model.WorkloadGroup) error {
	l := logger.L().Ctx(ctx).With("method", "CreateWorkloadGroup", "group_name", group.Name)
	l.Info("Attempting to create workload group")

	if err := group.Validate(); err != nil {
		l.Warnw("WorkloadGroup validation failed", "error", err)
		return errors.Wrap(err, errors.InvalidArgument, "invalid workload group configuration")
	}

	// Transform domain model to adapter model if they are different
	// For now, assume starrocks.WorkloadGroupDef is compatible or mapping is direct.
	adapterGroup := &starrocks.WorkloadGroupDef{
		Name:             group.Name,
		CPUShare:         group.CPUShare,
		MemoryLimit:      group.MemoryLimit,
		ConcurrencyLimit: group.ConcurrencyLimit,
		Properties:       group.Properties,
		// Map other fields
	}

	if err := s.srDDLExecutor.CreateWorkloadGroup(ctx, adapterGroup); err != nil {
		l.Errorw("Failed to create workload group via DDL executor", "error", err)
		// Specific error mapping can be done here (e.g., if StarRocks returns "already exists")
		return errors.Wrap(err, errors.DatabaseError, "failed to create workload group")
	}

	l.Info("Workload group created successfully")
	return nil
}

// GetWorkloadGroup retrieves the configuration of a specific workload group.
// GetWorkloadGroup 检索特定工作负载组的配置。
func (s *serviceImpl) GetWorkloadGroup(ctx context.Context, name string) (*model.WorkloadGroup, error) {
	l := logger.L().Ctx(ctx).With("method", "GetWorkloadGroup", "group_name", name)
	l.Info("Attempting to get workload group")

	if name == "" {
		return nil, errors.New(errors.InvalidArgument, "workload group name cannot be empty")
	}

	adapterGroup, err := s.srDDLExecutor.GetWorkloadGroup(ctx, name)
	if err != nil {
		l.Errorw("Failed to get workload group via DDL executor", "error", err)
		return nil, errors.Wrap(err, errors.DatabaseError, "failed to get workload group")
	}

	domainGroup := &model.WorkloadGroup{
		Name:             adapterGroup.Name,
		CPUShare:         adapterGroup.CPUShare,
		MemoryLimit:      adapterGroup.MemoryLimit,
		ConcurrencyLimit: adapterGroup.ConcurrencyLimit,
		Properties:       adapterGroup.Properties,
		// Map other fields
	}

	l.Info("Workload group retrieved successfully")
	return domainGroup, nil
}

// ListWorkloadGroups lists all configured workload groups.
// ListWorkloadGroups 列出所有已配置的工作负载组。
func (s *serviceImpl) ListWorkloadGroups(ctx context.Context, pagination *commontypes.PaginationRequest) (groups []*model.WorkloadGroup, total int64, err error) {
	l := logger.L().Ctx(ctx).With("method", "ListWorkloadGroups")
	l.Info("Attempting to list workload groups")

	// TODO: The starrocks.DDLExecutor interface doesn't currently have ListWorkloadGroups.
	// This would typically be implemented by "SHOW WORKLOAD GROUPS;" and then parsing.
	// For skeleton:
	l.Warn("ListWorkloadGroups skeleton: returning empty list and not implemented error")
	// As a placeholder, one could call GetWorkloadGroup for known/default groups if DDLExecutor supported it.
	// Or, a "SHOW WORKLOAD GROUPS" equivalent would be needed in the adapter.
	// Example:
	// adapterGroups, adapterTotal, err := s.srDDLExecutor.ListWorkloadGroups(ctx, pagination)
	// if err != nil { ... }
	// transform to domain model...
	return []*model.WorkloadGroup{}, 0, errors.New(errors.UnknownError, "ListWorkloadGroups not fully implemented")
}

// UpdateWorkloadGroup updates the configuration of an existing workload group.
// UpdateWorkloadGroup 更新现有工作负载组的配置。
func (s *serviceImpl) UpdateWorkloadGroup(ctx context.Context, group *model.WorkloadGroup) error {
	l := logger.L().Ctx(ctx).With("method", "UpdateWorkloadGroup", "group_name", group.Name)
	l.Info("Attempting to update workload group")

	if err := group.Validate(); err != nil {
		l.Warnw("WorkloadGroup validation failed for update", "error", err)
		return errors.Wrap(err, errors.InvalidArgument, "invalid workload group configuration for update")
	}

	adapterGroup := &starrocks.WorkloadGroupDef{
		Name:             group.Name,
		CPUShare:         group.CPUShare,
		MemoryLimit:      group.MemoryLimit,
		ConcurrencyLimit: group.ConcurrencyLimit,
		Properties:       group.Properties,
	}

	if err := s.srDDLExecutor.AlterWorkloadGroup(ctx, adapterGroup); err != nil {
		l.Errorw("Failed to update workload group via DDL executor", "error", err)
		return errors.Wrap(err, errors.DatabaseError, "failed to update workload group")
	}

	l.Info("Workload group updated successfully")
	return nil
}

// DeleteWorkloadGroup removes a workload group from StarRocks.
// DeleteWorkloadGroup 从StarRocks中移除一个工作负载组。
func (s *serviceImpl) DeleteWorkloadGroup(ctx context.Context, name string) error {
	l := logger.L().Ctx(ctx).With("method", "DeleteWorkloadGroup", "group_name", name)
	l.Info("Attempting to delete workload group")

	if name == "" {
		return errors.New(errors.InvalidArgument, "workload group name cannot be empty for deletion")
	}

	if err := s.srDDLExecutor.DropWorkloadGroup(ctx, name); err != nil {
		l.Errorw("Failed to delete workload group via DDL executor", "error", err)
		return errors.Wrap(err, errors.DatabaseError, "failed to delete workload group")
	}

	l.Info("Workload group deleted successfully")
	return nil
}
