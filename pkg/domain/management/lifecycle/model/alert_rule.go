package model

import (
	"time"
)

// AlertSeverity 表示告警的严重级别
// AlertSeverity represents the severity level of an alert.
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "CRITICAL" // 严重 Critical
	SeverityError    AlertSeverity = "ERROR"    // 错误 Error (less than critical)
	SeverityWarning  AlertSeverity = "WARNING"  // 警告 Warning
	SeverityInfo     AlertSeverity = "INFO"     // 信息 Info (rarely used for alerts, more for notifications)
)

// AlertRule defines the structure for an alert rule.
// AlertRule 定义告警规则的结构。
type AlertRule struct {
	// ID 规则的唯一标识符。
	// ID Unique identifier for the rule.
	ID string `json:"id"`

	// Name 规则的可读名称。
	// Name Human-readable name for the rule.
	Name string `json:"name"`

	// Description (可选) 规则的详细描述。
	// Description (Optional) Detailed description of the rule.
	Description string `json:"description,omitempty"`

	// IsEnabled 规则是否启用。
	// IsEnabled Whether the rule is enabled.
	IsEnabled bool `json:"isEnabled"`

	// Expression 告警表达式或查询条件。
	// 具体格式取决于底层的监控系统 (例如 PromQL for Prometheus, SQL-like for metric stores)。
	// Expression The alert expression or query condition.
	// The specific format depends on the underlying monitoring system (e.g., PromQL for Prometheus, SQL-like for metric stores).
	Expression string `json:"expression"`

	// ForDuration (可选) 告警条件需要持续多久才触发告警 (例如 "5m", "1h")。
	// ForDuration (Optional) How long the alert condition must persist before firing (e.g., "5m", "1h").
	ForDuration time.Duration `json:"forDuration,omitempty"` // Stored as duration, might be string like "5m" in config

	// Severity 告警的严重级别。
	// Severity Severity level of the alert.
	Severity AlertSeverity `json:"severity"`

	// Labels (可选) 附加到告警上的标签，用于路由或分组。
	// Labels (Optional) Labels attached to the alert for routing or grouping.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations (可选) 告警的附加信息，通常包含更详细的描述或运行手册链接。
	// Annotations (Optional) Additional information for the alert, often containing more detailed descriptions or runbook links.
	Annotations map[string]string `json:"annotations,omitempty"`

	// NotificationChannels (可选) 告警触发时通知的渠道列表 (例如 "email", "slack_channel_id", "webhook_url")。
	// NotificationChannels (Optional) List of channels to notify when the alert fires (e.g., "email", "slack_channel_id", "webhook_url").
	NotificationChannels []string `json:"notificationChannels,omitempty"`

	// CreatedAt 规则创建时间。
	// CreatedAt Timestamp when the rule was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt 规则最后更新时间。
	// UpdatedAt Timestamp when the rule was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
}

// ActiveAlert represents an active instance of an alert based on a rule.
// ActiveAlert 代表基于规则的活动告警实例。
type ActiveAlert struct {
	ID          string            `json:"id"`                   // 告警实例的唯一ID Unique ID for the alert instance
	RuleID      string            `json:"ruleId"`               // 触发此告警的规则ID ID of the rule that triggered this alert
	RuleName    string            `json:"ruleName"`             // 规则名称 Rule name
	State       string            `json:"state"`                // 告警状态 (例如 "PENDING", "FIRING", "RESOLVED") Alert state
	Severity    AlertSeverity     `json:"severity"`             // 严重级别 Severity
	ActiveAt    time.Time         `json:"activeAt"`             // 告警开始时间 Alert start time
	ResolvedAt  *time.Time        `json:"resolvedAt,omitempty"` // (可选) 告警解决时间 (Optional) Alert resolution time
	Labels      map[string]string `json:"labels"`               // 从规则和表达式中继承/生成的标签 Labels inherited/generated from rule and expression
	Annotations map[string]string `json:"annotations"`          // 告警的注解 Annotations for the alert
	Value       string            `json:"value"`                // 触发告警的指标值 Value that triggered the alert
	Summary     string            `json:"summary"`              // 告警摘要 Summary of the alert
	Description string            `json:"description"`          // 告警描述 Description of the alert
}

// Validate performs basic validation on the AlertRule.
// Validate 对 AlertRule 执行基本验证。
func (ar *AlertRule) Validate() error {
	if ar.Name == "" {
		return NewDomainError("AlertRule name cannot be empty")
	}
	if ar.Expression == "" {
		return NewDomainError("AlertRule expression cannot be empty")
	}
	if ar.Severity == "" { // Basic check, could validate against enum values
		return NewDomainError("AlertRule severity cannot be empty")
	}
	// Further validation for ForDuration format, labels, annotations, etc.
	return nil
}

// DomainError for lifecycle models
type DomainError struct{ message string }

func NewDomainError(message string) *DomainError { return &DomainError{message: message} }
func (e *DomainError) Error() string             { return e.message }
