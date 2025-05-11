package lifecycle

import (
	"context"

	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	"github.com/turtacn/dataseap/pkg/domain/management/lifecycle/model"
)

// Service defines the interface for managing the lifecycle of system components,
// including monitoring, alerting, and potentially other operational tasks.
// Service 定义了管理系统组件生命周期的接口，
// 包括监控、告警以及可能的其他运维任务。
type Service interface {
	// --- Component Status & Metrics ---

	// GetComponentStatus retrieves the operational status of one or more components.
	// GetComponentStatus 检索一个或多个组件的运行状态。
	// If componentName is empty, it might return status for all components of componentType,
	// or all managed components if componentType is also empty/unknown.
	// 如果 componentName 为空，则可能返回 componentType 类型的所有组件的状态，
	// 或者在 componentType 也为空/未知时返回所有受管组件的状态。
	GetComponentStatus(ctx context.Context, componentName string, componentType model.ComponentType) ([]*model.ComponentStatus, error)

	// GetSystemMetrics retrieves system or component metrics based on a query.
	// GetSystemMetrics 根据查询检索系统或组件指标。
	GetSystemMetrics(ctx context.Context, query *model.MetricsQueryRequest) ([]*model.SystemMetric, error)

	// --- Alerting ---

	// ListAlertRules lists all configured alert rules, with optional pagination.
	// ListAlertRules 列出所有已配置的告警规则，可选分页。
	ListAlertRules(ctx context.Context, pagination *commontypes.PaginationRequest) (rules []*model.AlertRule, total int64, err error)

	// GetAlertRule retrieves a specific alert rule by its ID.
	// GetAlertRule 通过ID检索特定的告警规则。
	GetAlertRule(ctx context.Context, ruleID string) (*model.AlertRule, error)

	// CreateAlertRule creates a new alert rule.
	// CreateAlertRule 创建一个新的告警规则。
	CreateAlertRule(ctx context.Context, rule *model.AlertRule) (createdRuleID string, err error)

	// UpdateAlertRule updates an existing alert rule.
	// UpdateAlertRule 更新一个已存在的告警规则。
	UpdateAlertRule(ctx context.Context, rule *model.AlertRule) error

	// DeleteAlertRule deletes an alert rule by its ID.
	// DeleteAlertRule 通过ID删除一个告警规则。
	DeleteAlertRule(ctx context.Context, ruleID string) error

	// ListActiveAlerts lists currently active (firing or pending) alerts.
	// ListActiveAlerts 列出当前活动的（触发中或待处理的）告警。
	// TODO: Define ListActiveAlertsRequest for filtering (e.g., by severity, ruleID) and pagination.
	ListActiveAlerts(ctx context.Context, pagination *commontypes.PaginationRequest) (alerts []*model.ActiveAlert, total int64, err error)

	// --- Cluster Operations (Examples, might be separate services) ---

	// TriggerUpgrade (示例) 触发组件或集群的升级流程。
	// TriggerUpgrade (Example) Triggers an upgrade process for a component or cluster.
	// TriggerUpgrade(ctx context.Context, componentType model.ComponentType, targetVersion string) (jobID string, err error)

	// ScaleComponent (示例) 调整组件的实例数量或资源。
	// ScaleComponent (Example) Adjusts the number of instances or resources for a component.
	// ScaleComponent(ctx context.Context, componentName string, componentType model.ComponentType, scaleRequest model.ScaleConfig) error

	// PerformBackup (示例) 执行备份操作。
	// PerformBackup(ctx context.Context, target string, backupOptions model.BackupOptions) (backupID string, err error)
}
