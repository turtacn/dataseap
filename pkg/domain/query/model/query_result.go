package model

import (
	"time"

	commontypes "github.com/turtacn/dataseap/pkg/common/types"
)

// QueryStats holds statistics about a query execution.
// QueryStats 保存查询执行的统计信息。
type QueryStats struct {
	ScanRows   int64         `json:"scanRows,omitempty"`   // 扫描的行数 Number of rows scanned
	ScanBytes  int64         `json:"scanBytes,omitempty"`  // 扫描的字节数 Number of bytes scanned
	Duration   time.Duration `json:"duration,omitempty"`   // 查询耗时 Query duration
	PeakMemory int64         `json:"peakMemory,omitempty"` // 峰值内存使用量 (字节) Peak memory usage in bytes
	CPUTime    time.Duration `json:"cpuTime,omitempty"`    // CPU耗时 CPU time spent
	Message    string        `json:"message,omitempty"`    // 其他信息或备注 Any other message or notes
}

// SQLQueryResult represents the result of an SQL query.
// SQLQueryResult 代表SQL查询的结果。
type SQLQueryResult struct {
	// Columns (可选) 列名列表。
	// Columns (Optional) List of column names.
	Columns []string `json:"columns,omitempty"`

	// Rows 查询结果的数据行。每行是一个map，键是列名，值是列值。
	// Rows Data rows of the query result. Each row is a map where key is column name and value is column value.
	Rows []map[string]interface{} `json:"rows"`

	// Pagination (可选) 分页信息，如果请求中包含了分页。
	// Pagination (Optional) Pagination information if pagination was included in the request.
	Pagination *commontypes.PaginationResponse `json:"pagination,omitempty"`

	// AffectedRows 对于DML语句，表示影响的行数。对于SELECT，通常为0。
	// AffectedRows For DML statements, represents the number of affected rows. Usually 0 for SELECT.
	AffectedRows int64 `json:"affectedRows,omitempty"`

	// Stats (可选) 查询执行的统计信息。
	// Stats (Optional) Statistics about the query execution.
	Stats *QueryStats `json:"stats,omitempty"`

	// ExecutionTime (可选) 查询在服务端的总执行时间。
	// ExecutionTime (Optional) Total execution time of the query on the server side.
	ExecutionTime time.Duration `json:"executionTime,omitempty"`
}

// SearchHit represents a single item found in a full-text search.
// SearchHit 代表全文检索中找到的单个条目。
type SearchHit struct {
	// SourceTable 命中的数据所在的表名或来源标识。
	// SourceTable Name of the table or source identifier where the hit occurred.
	SourceTable string `json:"sourceTable"`

	// ID (可选) 命中条目的唯一标识符。
	// ID (Optional) Unique identifier of the hit item.
	ID string `json:"id,omitempty"`

	// Score (可选) 结果的相关性得分。
	// Score (Optional) Relevance score of the hit.
	Score float32 `json:"score,omitempty"`

	// Document 完整的文档/行数据，通常是map[string]interface{}。
	// Document The full document/row data, typically map[string]interface{}.
	Document map[string]interface{} `json:"document"`

	// HitFields (可选) 命中关键字的具体字段和片段 (例如: map["message"] = "snippet with *keyword*...").
	// HitFields (Optional) Specific fields and snippets where keywords were hit (e.g., map["message"] = "snippet with *keyword*...").
	HitFields map[string]string `json:"hitFields,omitempty"`

	// Timestamp (可选) 文档的时间戳。
	// Timestamp (Optional) Timestamp of the document.
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// FullTextSearchResult represents the result of a full-text search operation.
// FullTextSearchResult 代表全文检索操作的结果。
type FullTextSearchResult struct {
	// Hits 检索到的命中结果列表。
	// Hits List of search hits found.
	Hits []*SearchHit `json:"hits"`

	// Pagination (可选) 分页信息。
	// Pagination (Optional) Pagination information.
	Pagination *commontypes.PaginationResponse `json:"pagination,omitempty"`

	// TotalHits (可选) 匹配的总命中数，可能与len(Hits)不同（如果分页）。
	// TotalHits (Optional) Total number of matching hits, may differ from len(Hits) if paginated.
	TotalHits int64 `json:"totalHits,omitempty"`

	// ExecutionTime (可选) 查询在服务端的总执行时间。
	// ExecutionTime (Optional) Total execution time of the query on the server side.
	ExecutionTime time.Duration `json:"executionTime,omitempty"`
}
