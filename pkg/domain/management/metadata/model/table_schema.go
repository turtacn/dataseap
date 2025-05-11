package model

import (
	"github.com/turtacn/dataseap/pkg/common/types/enum"
)

// FieldSchema represents the schema of a single field (column) in a table.
// FieldSchema 代表表中单个字段（列）的模式。
type FieldSchema struct {
	// Name 字段名称。
	// Name Field name.
	Name string `json:"name"`

	// DataType 字段的数据类型。
	// DataType Data type of the field.
	DataType enum.DataType `json:"dataType"`

	// TypeString 原始的类型字符串，例如 "VARCHAR(255)", "DECIMAL(10,2)"。
	// TypeString Original type string, e.g., "VARCHAR(255)", "DECIMAL(10,2)".
	TypeString string `json:"typeString"`

	// IsNullable 字段是否允许为空。
	// IsNullable Whether the field can be null.
	IsNullable bool `json:"isNullable"`

	// IsPrimaryKey 字段是否为主键的一部分。
	// IsPrimaryKey Whether the field is part of the primary key.
	IsPrimaryKey bool `json:"isPrimaryKey,omitempty"`

	// DefaultValue (可选) 字段的默认值 (以字符串表示)。
	// DefaultValue (Optional) Default value of the field (as a string).
	DefaultValue string `json:"defaultValue,omitempty"`

	// Comment (可选) 字段的注释。
	// Comment (Optional) Comment for the field.
	Comment string `json:"comment,omitempty"`

	// AggregationType (可选) StarRocks特定：字段的聚合类型 (例如 "SUM", "REPLACE", "NONE")。
	// AggregationType (Optional) StarRocks specific: Aggregation type for the field (e.g., "SUM", "REPLACE", "NONE").
	AggregationType string `json:"aggregationType,omitempty"`

	// Collate (可选) 字符集和排序规则。
	// Collate (Optional) Character set and collation.
	Collate string `json:"collate,omitempty"`
}

// TableSchema represents the schema of a table.
// TableSchema 代表表的模式。
type TableSchema struct {
	// DatabaseName 表所属的数据库名称。
	// DatabaseName Name of the database the table belongs to.
	DatabaseName string `json:"databaseName"`

	// TableName 表的名称。
	// TableName Name of the table.
	TableName string `json:"tableName"`

	// Fields 表中所有字段的模式列表。
	// Fields List of schemas for all fields in the table.
	Fields []*FieldSchema `json:"fields"`

	// TableType (可选) StarRocks特定：表类型 (例如 "OLAP", "MYSQL", "HIVE", "ICEBERG")。
	// TableType (Optional) StarRocks specific: Table type (e.g., "OLAP", "MYSQL", "HIVE", "ICEBERG").
	TableType string `json:"tableType,omitempty"`

	// KeysType (可选) StarRocks特定：OLAP表的键类型 (例如 "DUPLICATE KEY", "AGGREGATE KEY", "UNIQUE KEY", "PRIMARY KEY")。
	// KeysType (Optional) StarRocks specific: Key type for OLAP tables (e.g., "DUPLICATE KEY", "AGGREGATE KEY", "UNIQUE KEY", "PRIMARY KEY").
	KeysType string `json:"keysType,omitempty"`

	// PartitionInfo (可选) StarRocks特定：表的分区信息描述 (可能是结构化对象或字符串)。
	// PartitionInfo (Optional) StarRocks specific: Description of the table's partitioning information (could be a structured object or string).
	PartitionInfo string `json:"partitionInfo,omitempty"` // Could be a more structured type

	// DistributionInfo (可选) StarRocks特定：表的分桶（分布）信息描述 (可能是结构化对象或字符串)。
	// DistributionInfo (Optional) StarRocks specific: Description of the table's bucketing (distribution) information (could be a structured object or string).
	DistributionInfo string `json:"distributionInfo,omitempty"` // Could be a more structured type

	// Properties (可选) StarRocks特定：表的其他属性 (例如 "replication_num"="3")。
	// Properties (Optional) StarRocks specific: Other properties of the table (e.g., "replication_num"="3").
	Properties map[string]string `json:"properties,omitempty"`

	// Comment (可选) 表的注释。
	// Comment (Optional) Comment for the table.
	Comment string `json:"comment,omitempty"`

	// CreateTableDDL (可选) 生成该表的DDL语句。
	// CreateTableDDL (Optional) The DDL statement that creates this table.
	CreateTableDDL string `json:"createTableDdl,omitempty"`
}

// IndexDefinition represents the definition of an index on a table.
// IndexDefinition 代表表上索引的定义。
type IndexDefinition struct {
	// IndexName 索引的名称。
	// IndexName Name of the index.
	IndexName string `json:"indexName"`

	// TableName 索引所属的表名。
	// TableName Name of the table the index belongs to.
	TableName string `json:"tableName"`

	// DatabaseName 索引所属的数据库名。
	// DatabaseName Name of the database the index belongs to.
	DatabaseName string `json:"databaseName"`

	// IndexType 索引类型 (例如 "BITMAP", "INVERTED")。
	// IndexType Type of the index (e.g., "BITMAP", "INVERTED").
	// Consider using an enum if types are fixed, e.g., enum.IndexType.
	IndexType string `json:"indexType"`

	// Fields 索引包含的字段列表。
	// Fields List of fields included in the index.
	Fields []string `json:"fields"`

	// Properties (可选) 索引的特定属性 (例如 倒排索引的分词器 'parser'='chinese')。
	// Properties (Optional) Specific properties of the index (e.g., tokenizer 'parser'='chinese' for inverted index).
	Properties map[string]string `json:"properties,omitempty"`

	// Comment (可选) 索引的注释。
	// Comment (Optional) Comment for the index.
	Comment string `json:"comment,omitempty"`
}

// MaterializedViewDefinition represents the definition of a materialized view.
// MaterializedViewDefinition 代表物化视图的定义。
type MaterializedViewDefinition struct {
	// DatabaseName 物化视图所属的数据库名称。
	// DatabaseName Name of the database the materialized view belongs to.
	DatabaseName string `json:"databaseName"`

	// ViewName 物化视图的名称。
	// ViewName Name of the materialized view.
	ViewName string `json:"viewName"`

	// Query 定义物化视图的SQL查询语句。
	// Query The SQL query statement defining the materialized view.
	Query string `json:"query"`

	// RefreshType (可选) StarRocks特定：刷新类型 (例如 "ASYNC", "MANUAL")。
	// RefreshType (Optional) StarRocks specific: Refresh type (e.g., "ASYNC", "MANUAL").
	RefreshType string `json:"refreshType,omitempty"`

	// RefreshSchedule (可选) StarRocks特定：如果是异步刷新，则为刷新调度表达式 (例如 "EVERY 1 HOUR")。
	// RefreshSchedule (Optional) StarRocks specific: If asynchronous refresh, the refresh schedule expression (e.g., "EVERY 1 HOUR").
	RefreshSchedule string `json:"refreshSchedule,omitempty"`

	// Properties (可选) StarRocks特定：物化视图的其他属性。
	// Properties (Optional) StarRocks specific: Other properties of the materialized view.
	Properties map[string]string `json:"properties,omitempty"`

	// Comment (可选) 物化视图的注释。
	// Comment (Optional) Comment for the materialized view.
	Comment string `json:"comment,omitempty"`
}

// Validate performs basic validation.
func (fs *FieldSchema) Validate() error {
	if fs.Name == "" {
		return NewDomainError("FieldSchema name cannot be empty")
	}
	if fs.TypeString == "" && fs.DataType == enum.DataTypeUnknown {
		return NewDomainError("FieldSchema TypeString or DataType must be specified")
	}
	return nil
}

// Validate performs basic validation.
func (ts *TableSchema) Validate() error {
	if ts.DatabaseName == "" {
		return NewDomainError("TableSchema DatabaseName cannot be empty")
	}
	if ts.TableName == "" {
		return NewDomainError("TableSchema TableName cannot be empty")
	}
	if len(ts.Fields) == 0 {
		return NewDomainError("TableSchema must have at least one field")
	}
	for _, f := range ts.Fields {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate performs basic validation.
func (id *IndexDefinition) Validate() error {
	if id.DatabaseName == "" {
		return NewDomainError("IndexDefinition DatabaseName cannot be empty")
	}
	if id.TableName == "" {
		return NewDomainError("IndexDefinition TableName cannot be empty")
	}
	if id.IndexName == "" {
		return NewDomainError("IndexDefinition IndexName cannot be empty")
	}
	if id.IndexType == "" {
		return NewDomainError("IndexDefinition IndexType cannot be empty")
	}
	if len(id.Fields) == 0 {
		return NewDomainError("IndexDefinition Fields cannot be empty")
	}
	return nil
}

// Validate performs basic validation.
func (mvd *MaterializedViewDefinition) Validate() error {
	if mvd.DatabaseName == "" {
		return NewDomainError("MaterializedViewDefinition DatabaseName cannot be empty")
	}
	if mvd.ViewName == "" {
		return NewDomainError("MaterializedViewDefinition ViewName cannot be empty")
	}
	if mvd.Query == "" {
		return NewDomainError("MaterializedViewDefinition Query cannot be empty")
	}
	return nil
}

// DomainError for metadata models
type DomainError struct{ message string }

func NewDomainError(message string) *DomainError { return &DomainError{message: message} }
func (e *DomainError) Error() string             { return e.message }
