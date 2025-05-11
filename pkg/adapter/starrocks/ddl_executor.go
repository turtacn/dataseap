package starrocks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/config" // Assuming config is available
	"github.com/turtacn/dataseap/pkg/logger"
)

// starrocksDDLExecutor implements the DDLExecutor interface for StarRocks.
// starrocksDDLExecutor 实现StarRocks的DDLExecutor接口。
type starrocksDDLExecutor struct {
	client Client                 // Uses the general client for executing SQL DDLs
	cfg    config.StarRocksConfig // StarRocks configuration
}

// NewDDLExecutor creates a new StarRocks DDL executor.
// NewDDLExecutor 创建一个新的StarRocks DDL执行器。
func NewDDLExecutor(client Client, cfg config.StarRocksConfig) (DDLExecutor, error) {
	if client == nil {
		return nil, errors.New(errors.InvalidArgument, "StarRocks client cannot be nil for DDLExecutor")
	}
	return &starrocksDDLExecutor{
		client: client,
		cfg:    cfg,
	}, nil
}

func (e *starrocksDDLExecutor) executeDDL(ctx context.Context, ddl string) error {
	l := logger.L().With("method", "executeDDL", "ddl", ddl)
	l.Info("Executing DDL statement")

	result, err := e.client.Execute(ctx, ddl)
	if err != nil {
		l.Errorw("Failed to execute DDL", "error", err)
		return errors.Wrapf(err, errors.DatabaseError, "failed to execute DDL: %s", ddl)
	}
	if result.Error != nil {
		l.Errorw("DDL execution resulted in error", "starrocks_error", result.Error)
		return errors.Wrapf(result.Error, errors.DatabaseError, "DDL execution resulted in StarRocks error: %s", ddl)
	}
	// For DDL, result.Rows is usually empty or has status info.
	// We might want to check specific success messages if available in result.Stats or a conventional row.
	l.Info("DDL statement executed successfully")
	return nil
}

// CreateTable creates a new table based on the provided schema.
// CreateTable 根据提供的schema创建新表。
func (e *starrocksDDLExecutor) CreateTable(ctx context.Context, schema *TableSchemaDef) error {
	// TODO: Implement DDL generation from TableSchemaDef
	// This would involve constructing a "CREATE TABLE ..." string.
	// Example:
	// var sb strings.Builder
	// sb.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (", schema.DatabaseName, schema.TableName))
	// for i, field := range schema.Fields {
	//     sb.WriteString(fmt.Sprintf("%s %s", field.Name, field.Type))
	//     if !field.IsNullable {
	//         sb.WriteString(" NOT NULL")
	//     }
	//     if field.Comment != "" {
	//         sb.WriteString(fmt.Sprintf(" COMMENT '%s'", field.Comment))
	//     }
	//     if i < len(schema.Fields)-1 {
	//         sb.WriteString(", ")
	//     }
	// }
	// sb.WriteString(") ")
	// // Add ENGINE, KEYS, PARTITION, DISTRIBUTION, PROPERTIES clauses
	// ddl := sb.String()
	// return e.executeDDL(ctx, ddl)
	logger.L().Warnw("CreateTable DDL generation not fully implemented", "schema", schema)
	return errors.New(errors.UnknownError, "CreateTable DDL generation not fully implemented")
}

// AlterTable modifies an existing table.
// AlterTable 修改现有表。
func (e *starrocksDDLExecutor) AlterTable(ctx context.Context, database, table, alterDDLStatement string) error {
	// Ensure database and table are part of the DDL or correctly prefixed if needed
	// For now, assume alterDDLStatement is a complete and valid DDL.
	if !strings.HasPrefix(strings.ToUpper(alterDDLStatement), "ALTER TABLE") {
		return errors.Newf(errors.InvalidArgument, "statement is not an ALTER TABLE DDL: %s", alterDDLStatement)
	}
	// Add database context if not in statement
	// e.g., by temporarily setting session database or ensuring table name is fully qualified.
	// For simplicity, assume the DDL statement is self-contained or client handles database context.
	return e.executeDDL(ctx, alterDDLStatement)
}

// DropTable drops an existing table.
// DropTable 删除现有表。
func (e *starrocksDDLExecutor) DropTable(ctx context.Context, database, table string) error {
	ddl := fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", database, table)
	return e.executeDDL(ctx, ddl)
}

// CreateIndex creates an index on a table.
// CreateIndex 在表上创建索引。
func (e *starrocksDDLExecutor) CreateIndex(ctx context.Context, database, table string, index *IndexDefinitionDef) error {
	// Example DDL: ALTER TABLE db.table ADD INDEX index_name (col1, col2) USING BITMAP COMMENT 'comment';
	// Example DDL for Inverted: ALTER TABLE db.table ADD INDEX index_name (col_text) USING INVERTED PROPERTIES("parser" = "chinese") COMMENT 'comment';
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s.%s ADD INDEX %s (", database, table, index.IndexName))
	sb.WriteString(strings.Join(index.Fields, ", "))
	sb.WriteString(fmt.Sprintf(") USING %s", strings.ToUpper(index.IndexType)))

	if len(index.Properties) > 0 {
		sb.WriteString(" PROPERTIES(")
		props := []string{}
		for k, v := range index.Properties {
			props = append(props, fmt.Sprintf("\"%s\" = \"%s\"", k, v))
		}
		sb.WriteString(strings.Join(props, ", "))
		sb.WriteString(")")
	}

	if index.Comment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT '%s'", index.Comment))
	}
	sb.WriteString(";")
	return e.executeDDL(ctx, sb.String())
}

// DropIndex drops an index from a table.
// DropIndex 从表中删除索引。
func (e *starrocksDDLExecutor) DropIndex(ctx context.Context, database, table, indexName string) error {
	ddl := fmt.Sprintf("ALTER TABLE %s.%s DROP INDEX %s", database, table, indexName)
	return e.executeDDL(ctx, ddl)
}

// GetTableSchema retrieves the schema of a specific table.
// GetTableSchema 检索特定表的schema。
func (e *starrocksDDLExecutor) GetTableSchema(ctx context.Context, database, table string) (*TableSchemaDef, error) {
	// DDL: DESCRIBE database.table; or SHOW CREATE TABLE database.table;
	// Parsing DESCRIBE output is generally easier.
	ddl := fmt.Sprintf("DESCRIBE %s.%s", database, table)
	result, err := e.client.Execute(ctx, ddl)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	schema := &TableSchemaDef{
		DatabaseName: database,
		TableName:    table,
		Fields:       []FieldSchemaDef{},
	}

	// Expected columns from DESCRIBE: Field, Type, Null, Key, Default, Extra
	if len(result.Columns) < 6 {
		return nil, errors.Newf(errors.InternalError, "unexpected DESCRIBE result format: expected at least 6 columns, got %d", len(result.Columns))
	}

	for _, row := range result.Rows {
		if len(row) < 6 {
			logger.L().Warnw("Skipping malformed row in DESCRIBE result", "row", row)
			continue
		}
		field := FieldSchemaDef{}
		if fName, ok := row[0].(string); ok {
			field.Name = fName
		}
		if fType, ok := row[1].(string); ok {
			field.Type = fType
		}
		if fNull, ok := row[2].(string); ok {
			field.IsNullable = (strings.ToUpper(fNull) == "YES")
		}
		// Key, Default, Extra can also be parsed
		// For now, simple parsing.
		schema.Fields = append(schema.Fields, field)
	}
	return schema, nil
}

// CreateWorkloadGroup creates a new workload group.
// CreateWorkloadGroup 创建一个新的工作负载组。
func (e *starrocksDDLExecutor) CreateWorkloadGroup(ctx context.Context, group *WorkloadGroupDef) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE WORKLOAD GROUP %s PROPERTIES (", group.Name))
	props := []string{fmt.Sprintf("'cpu_share'='%d'", group.CPUShare)}
	if group.MemoryLimit != "" {
		props = append(props, fmt.Sprintf("'memory_limit'='%s'", group.MemoryLimit))
	}
	if group.ConcurrencyLimit > 0 {
		props = append(props, fmt.Sprintf("'concurrency_limit'='%d'", group.ConcurrencyLimit))
	}
	// Add other properties if any
	sb.WriteString(strings.Join(props, ", "))
	sb.WriteString(");")
	return e.executeDDL(ctx, sb.String())
}

// AlterWorkloadGroup modifies an existing workload group.
// AlterWorkloadGroup 修改现有的工作负载组。
func (e *starrocksDDLExecutor) AlterWorkloadGroup(ctx context.Context, group *WorkloadGroupDef) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER WORKLOAD GROUP %s PROPERTIES (", group.Name))
	props := []string{fmt.Sprintf("'cpu_share'='%d'", group.CPUShare)}
	if group.MemoryLimit != "" {
		props = append(props, fmt.Sprintf("'memory_limit'='%s'", group.MemoryLimit))
	}
	if group.ConcurrencyLimit > 0 {
		props = append(props, fmt.Sprintf("'concurrency_limit'='%d'", group.ConcurrencyLimit))
	}
	sb.WriteString(strings.Join(props, ", "))
	sb.WriteString(");")
	return e.executeDDL(ctx, sb.String())
}

// DropWorkloadGroup drops an existing workload group.
// DropWorkloadGroup 删除现有的工作负载组。
func (e *starrocksDDLExecutor) DropWorkloadGroup(ctx context.Context, groupName string) error {
	ddl := fmt.Sprintf("DROP WORKLOAD GROUP %s;", groupName)
	return e.executeDDL(ctx, ddl)
}

// GetWorkloadGroup retrieves information about a specific workload group.
// GetWorkloadGroup 检索关于特定工作负载组的信息。
func (e *starrocksDDLExecutor) GetWorkloadGroup(ctx context.Context, groupName string) (*WorkloadGroupDef, error) {
	// DDL: SHOW WORKLOAD GROUPS WHERE Name = 'groupName';
	// Or SHOW WORKLOAD GROUPS; and filter client-side.
	ddl := fmt.Sprintf("SHOW WORKLOAD GROUPS WHERE Name = '%s'", groupName)
	result, err := e.client.Execute(ctx, ddl)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	if len(result.Rows) == 0 {
		return nil, errors.Newf(errors.NotFoundError, "workload group '%s' not found", groupName)
	}

	// Assuming columns: Id, Name, CpuShare, MemLimit, ConcurrencyLimit, MaxQueueSize, MemTrackerType, SpillThresholdLow, SpillThresholdHigh, EnableMemoryOvercommit
	row := result.Rows[0]
	group := &WorkloadGroupDef{Name: groupName}

	colMap := make(map[string]int)
	for i, colName := range result.Columns {
		colMap[colName] = i
	}

	if idx, ok := colMap["CpuShare"]; ok && len(row) > idx {
		if val, valOk := row[idx].(float64); valOk { // JSON numbers are float64
			group.CPUShare = int32(val)
		} else if valStr, valStrOk := row[idx].(string); valStrOk { // Sometimes it might be string
			cpuShare, _ := strconv.Atoi(valStr)
			group.CPUShare = int32(cpuShare)
		}
	}
	if idx, ok := colMap["MemLimit"]; ok && len(row) > idx {
		if val, valOk := row[idx].(string); valOk {
			group.MemoryLimit = val
		}
	}
	if idx, ok := colMap["ConcurrencyLimit"]; ok && len(row) > idx {
		if val, valOk := row[idx].(float64); valOk {
			group.ConcurrencyLimit = int32(val)
		} else if valStr, valStrOk := row[idx].(string); valStrOk {
			concurrency, _ := strconv.Atoi(valStr)
			group.ConcurrencyLimit = int32(concurrency)
		}
	}
	// Parse other properties as needed

	return group, nil
}
