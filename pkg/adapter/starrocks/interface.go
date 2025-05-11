package starrocks

import (
	"context"
	"io"
	"time"
	// Assuming proto definitions are available or will be mapped to domain models
	// For now, we'll define placeholder structs or use basic types.
	// We can later refine these to use generated gRPC types or domain models.
	// "github.com/turtacn/dataseap/api/v1;apiv1"
	// "github.com/turtacn/dataseap/pkg/domain/management/metadata" // Example for domain model
)

// QueryResult holds the result of a StarRocks query.
// QueryResult 保存 StarRocks 查询的结果。
type QueryResult struct {
	Columns []string        // 列名列表 Column names
	Rows    [][]interface{} // 数据行, 每行是值的切片 Rows of data, each row is a slice of values
	Error   error           // 查询期间发生的错误 Error during query execution
	Stats   *QueryStats     // 查询统计信息 Query statistics
}

// QueryStats holds statistics about a query execution.
// QueryStats 保存查询执行的统计信息。
type QueryStats struct {
	ScanRows   int64         // 扫描的行数 Number of rows scanned
	ScanBytes  int64         // 扫描的字节数 Number of bytes scanned
	Duration   time.Duration // 查询耗时 Query duration
	PeakMemory int64         // 峰值内存使用量 (字节) Peak memory usage in bytes
	CPUTime    time.Duration // CPU耗时 CPU time spent
	Message    string        // 其他信息或备注 Any other message or notes
}

// StreamLoadOptions holds options for a StarRocks stream load operation.
// StreamLoadOptions 保存 StarRocks Stream Load 操作的选项。
type StreamLoadOptions struct {
	Format          string            // 数据格式 (e.g., "json", "csv") Data format
	ColumnSeparator string            // CSV列分隔符 (e.g., ",") CSV column separator
	RowDelimiter    string            // CSV行分隔符 (e.g., "\n") CSV row delimiter
	StripOuterArray bool              // JSON格式：是否去除外层数组 JSON format: whether to strip the outer array
	TimeoutSeconds  int               // 加载超时时间 (秒) Load timeout in seconds
	MaxFilterRatio  float64           // 最大容错率 Maximum filter ratio
	Headers         map[string]string // 自定义HTTP头部 Custom HTTP headers (e.g., for specific columns from CSV header)
	MergeCondition  string            // (可选) 用于部分更新的合并条件 (Optional) Merge condition for partial updates
	TwoPhaseCommit  bool              // (可选) 是否开启两阶段提交 (Optional) Enable two-phase commit
	TransactionID   string            // (可选) 用于两阶段提交的事务ID (Optional) Transaction ID for two-phase commit
}

// StreamLoadResponse holds the response from a StarRocks stream load operation.
// StreamLoadResponse 保存 StarRocks Stream Load 操作的响应。
type StreamLoadResponse struct {
	TxnID                  int64  // 事务ID Transaction ID
	Label                  string // 加载任务的标签 Label of the load job
	Status                 string // 加载状态 (e.g., "Success", "Fail", "Publish Timeout") Load status
	ExistingTxnID          int64  // (可选) 如果Label重复，则为已存在的事务ID (Optional) Existing transaction ID if label is duplicated
	Message                string // 详细信息 Detailed message
	NumberTotalRows        int64  // 总处理行数 Total rows processed
	NumberLoadedRows       int64  // 成功导入行数 Successfully loaded rows
	NumberFilteredRows     int64  // 过滤行数 Filtered rows
	NumberUnselectedRows   int64  // 未选中行数 Unselected rows (due to where clause)
	LoadBytes              int64  // 加载的字节数 Bytes loaded
	LoadTimeMs             int64  // 加载耗时 (毫秒) Load time in milliseconds
	BeginTxnTimeMs         int64  // 开始事务耗时 (毫秒) Begin transaction time in milliseconds
	StreamLoadPutTimeMs    int64  // Stream Load Put 操作耗时 (毫秒) Stream Load Put operation time in milliseconds
	ReadDataTimeMs         int64  // 读取数据耗时 (毫秒) Read data time in milliseconds
	WriteDataTimeMs        int64  // 写入数据耗时 (毫秒) Write data time in milliseconds
	CommitAndPublishTimeMs int64  // 提交和发布耗时 (毫秒) Commit and publish time in milliseconds
	ErrorURL               string // 如果有错误，相关的错误日志URL If errors occurred, the URL for error logs
}

// Client defines the interface for interacting with StarRocks for queries and data loading.
// Client 定义了与StarRocks进行查询和数据加载交互的接口。
type Client interface {
	// Execute performs a DQL (SELECT) or DML (INSERT, UPDATE, DELETE - less common via this for OLAP) query.
	// Execute 执行 DQL (SELECT) 或 DML (INSERT, UPDATE, DELETE - OLAP场景下不常用) 查询。
	Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)

	// StreamLoad ingests data into a StarRocks table using the Stream Load method.
	// StreamLoad 使用 Stream Load 方法将数据导入到 StarRocks 表中。
	// 'data' is an io.Reader providing the data to be loaded.
	// 'data' 是一个提供待加载数据的 io.Reader。
	StreamLoad(ctx context.Context, database, table string, data io.Reader, opts *StreamLoadOptions) (*StreamLoadResponse, error)

	// BeginTransaction (可选) 开始一个两阶段提交事务 (用于Stream Load)
	// BeginTransaction (Optional) begins a two-phase commit transaction (for Stream Load).
	BeginTransaction(ctx context.Context, database, table, label string, timeoutSeconds int) (int64, error) // Returns TxnID

	// CommitTransaction (可选) 提交一个两阶段提交事务
	// CommitTransaction (Optional) commits a two-phase commit transaction.
	CommitTransaction(ctx context.Context, database string, txnID int64) error

	// AbortTransaction (可选) 中止一个两阶段提交事务
	// AbortTransaction (Optional) aborts a two-phase commit transaction.
	AbortTransaction(ctx context.Context, database string, txnID int64) error

	// Close terminates any open connections to StarRocks.
	// Close 关闭所有到 StarRocks 的打开连接。
	Close() error
}

// Placeholder types that would typically come from domain or proto definitions.
// We define them here temporarily for the interface.
// 这些占位符类型通常来自领域模型或 proto 定义。
// 我们在此处临时定义它们以供接口使用。

// TableSchemaDef 定义表的结构信息
// TableSchemaDef defines the schema information of a table.
type TableSchemaDef struct {
	DatabaseName string
	TableName    string
	Fields       []FieldSchemaDef
	// ... 其他如分区、分桶、属性等信息
	// ... Other info like partitioning, bucketing, properties etc.
}

// FieldSchemaDef 定义表字段的结构信息
// FieldSchemaDef defines the schema information of a table field.
type FieldSchemaDef struct {
	Name       string
	Type       string // e.g., "INT", "VARCHAR(255)"
	IsNullable bool
	Comment    string
	// ... 其他属性
	// ... Other properties
}

// IndexDefinitionDef 定义索引信息
// IndexDefinitionDef defines index information.
type IndexDefinitionDef struct {
	IndexName  string
	IndexType  string // "BITMAP", "INVERTED"
	Fields     []string
	Properties map[string]string // e.g., for inverted index parser
	Comment    string
}

// WorkloadGroupDef 定义工作负载组信息
// WorkloadGroupDef defines workload group information.
type WorkloadGroupDef struct {
	Name             string
	CPUShare         int32
	MemoryLimit      string
	ConcurrencyLimit int32
	// ... 其他属性
	// ... Other properties
}

// DDLExecutor defines the interface for executing Data Definition Language (DDL) commands.
// DDLExecutor 定义了执行数据定义语言 (DDL) 命令的接口。
type DDLExecutor interface {
	// CreateTable creates a new table based on the provided schema.
	// CreateTable 根据提供的schema创建新表。
	CreateTable(ctx context.Context, schema *TableSchemaDef) error

	// AlterTable modifies an existing table (e.g., add column, drop column).
	// AlterTable 修改现有表 (例如，添加列，删除列)。
	// For simplicity, a raw DDL string might be passed, or more structured requests.
	// 为简单起见，可以传递原始DDL字符串，或更结构化的请求。
	AlterTable(ctx context.Context, database, table, alterDDLStatement string) error

	// DropTable drops an existing table.
	// DropTable 删除现有表。
	DropTable(ctx context.Context, database, table string) error

	// CreateIndex creates an index on a table.
	// CreateIndex 在表上创建索引。
	CreateIndex(ctx context.Context, database, table string, index *IndexDefinitionDef) error

	// DropIndex drops an index from a table.
	// DropIndex 从表中删除索引。
	DropIndex(ctx context.Context, database, table, indexName string) error

	// GetTableSchema retrieves the schema of a specific table.
	// GetTableSchema 检索特定表的schema。
	GetTableSchema(ctx context.Context, database, table string) (*TableSchemaDef, error)

	// CreateWorkloadGroup creates a new workload group.
	// CreateWorkloadGroup 创建一个新的工作负载组。
	CreateWorkloadGroup(ctx context.Context, group *WorkloadGroupDef) error

	// AlterWorkloadGroup modifies an existing workload group.
	// AlterWorkloadGroup 修改现有的工作负载组。
	AlterWorkloadGroup(ctx context.Context, group *WorkloadGroupDef) error

	// DropWorkloadGroup drops an existing workload group.
	// DropWorkloadGroup 删除现有的工作负载组。
	DropWorkloadGroup(ctx context.Context, groupName string) error

	// GetWorkloadGroup retrieves information about a specific workload group.
	// GetWorkloadGroup 检索关于特定工作负载组的信息。
	GetWorkloadGroup(ctx context.Context, groupName string) (*WorkloadGroupDef, error)
}
