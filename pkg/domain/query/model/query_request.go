package model

import (
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
)

// SQLQueryRequest represents a request to execute an SQL query.
// SQLQueryRequest 代表执行SQL查询的请求。
type SQLQueryRequest struct {
	// SQL is the raw SQL query string.
	// SQL 是原始的SQL查询字符串。
	SQL string `json:"sql"`

	// Params (可选) SQL查询的参数，用于防止SQL注入，键为参数名，值为参数值。
	// Params (Optional) Parameters for the SQL query to prevent SQL injection.
	// The key is the parameter name, and the value is the parameter value.
	Params map[string]interface{} `json:"params,omitempty"`

	// Pagination (可选) 分页参数。
	// Pagination (Optional) Pagination parameters.
	Pagination *commontypes.PaginationRequest `json:"pagination,omitempty"`

	// WorkloadGroup (可选) 指定查询使用的工作负载组。
	// WorkloadGroup (Optional) Specifies the workload group to be used for the query.
	WorkloadGroup string `json:"workloadGroup,omitempty"`

	// QueryTimeoutSecs (可选) 查询超时时间（秒）。如果为0或负数，则使用系统默认值。
	// QueryTimeoutSecs (Optional) Query timeout in seconds. If 0 or negative, system default is used.
	QueryTimeoutSecs int `json:"queryTimeoutSecs,omitempty"`

	// Database (可选) 指定查询的数据库。如果为空，则使用连接的默认数据库。
	// Database (Optional) Specifies the database for the query. If empty, uses the connection's default database.
	Database string `json:"database,omitempty"`
}

// FullTextSearchRequest represents a request for a full-text search operation.
// FullTextSearchRequest 代表全文检索操作的请求。
type FullTextSearchRequest struct {
	// Keywords 检索的关键字，可以是空格分隔的多个词。
	// Keywords Keywords for the search, can be multiple space-separated words.
	Keywords string `json:"keywords"`

	// TargetTables (可选) 指定要搜索的表名列表。如果为空，则可能搜索所有已配置全文检索的表。
	// TargetTables (Optional) List of table names to search. If empty, may search all tables configured for full-text search.
	TargetTables []string `json:"targetTables,omitempty"`

	// TargetFields (可选) 指定在表内要搜索的字段列表。如果为空，则搜索表内所有已配置索引的文本字段。
	// TargetFields (Optional) List of fields to search within tables. If empty, searches all indexed text fields in the table.
	TargetFields []string `json:"targetFields,omitempty"`

	// Tokenizer (可选) 指定分词器，如 "standard", "english", "chinese"。
	// Tokenizer (Optional) Specify the tokenizer, e.g., "standard", "english", "chinese".
	Tokenizer string `json:"tokenizer,omitempty"`

	// RecallPriority 召回优先模式。如果为 true，则更倾向于召回更多可能相关的结果 (例如 StarRocks 的 MATCH_ANY)。
	// 如果为 false，则更倾向于精确匹配 (例如 StarRocks 的 MATCH_ALL)。默认为 true。
	// RecallPriority Recall priority mode. If true, favors recalling more potentially relevant results (e.g., StarRocks' MATCH_ANY).
	// If false, favors precise matching (e.g., StarRocks' MATCH_ALL). Defaults to true.
	RecallPriority bool `json:"recallPriority"`

	// Pagination (可选) 分页参数。
	// Pagination (Optional) Pagination parameters.
	Pagination *commontypes.PaginationRequest `json:"pagination,omitempty"`

	// TimeRangeFilter (可选) 时间范围过滤。
	// TimeRangeFilter (Optional) Time range filter.
	TimeRangeFilter *commontypes.TimeRange `json:"timeRangeFilter,omitempty"`

	// SortBy (可选) 排序字段和顺序。
	// SortBy (Optional) Sort fields and order.
	SortBy []*commontypes.SortField `json:"sortBy,omitempty"`

	// AdditionalFilters (可选) 额外的过滤条件，例如 map["status"] = "active"
	// AdditionalFilters (Optional) Additional filter conditions, e.g., map["status"] = "active"
	AdditionalFilters map[string]interface{} `json:"additionalFilters,omitempty"`
}

// Validate performs basic validation on the SQLQueryRequest.
// Validate 对 SQLQueryRequest 执行基本验证。
func (req *SQLQueryRequest) Validate() error {
	if req.SQL == "" {
		return NewDomainError("SQL query string cannot be empty")
	}
	// Further validation for pagination, timeout, etc. can be added here.
	return nil
}

// Validate performs basic validation on the FullTextSearchRequest.
// Validate 对 FullTextSearchRequest 执行基本验证。
func (req *FullTextSearchRequest) Validate() error {
	if req.Keywords == "" {
		return NewDomainError("Keywords for full-text search cannot be empty")
	}
	// Further validation for pagination, etc. can be added here.
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
