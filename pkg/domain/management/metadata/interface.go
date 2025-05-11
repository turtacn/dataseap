package metadata

import (
	"context"

	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	"github.com/turtacn/dataseap/pkg/domain/management/metadata/model"
)

// Service defines the interface for managing metadata of data assets.
// Service 定义了管理数据资产元数据的接口。
type Service interface {
	// GetTableSchema retrieves the schema of a specific table.
	// GetTableSchema 检索特定表的模式。
	GetTableSchema(ctx context.Context, databaseName, tableName string) (*model.TableSchema, error)

	// ListTables lists tables within a given database, with optional pagination.
	// ListTables 列出给定数据库中的表，可选分页。
	ListTables(ctx context.Context, databaseName string, pagination *commontypes.PaginationRequest) (tableNames []string, total int64, err error)

	// CreateTable (可选的高级功能) 创建表。
	// CreateTable (Optional advanced feature) Creates a table.
	// CreateTable(ctx context.Context, tableSchema *model.TableSchema) error

	// AlterTable (可选的高级功能) 修改表结构。
	// AlterTable (Optional advanced feature) Alters a table structure.
	// AlterTable(ctx context.Context, databaseName, tableName string, alterations []model.TableAlteration) error

	// CreateIndex creates an index on a table.
	// CreateIndex 在表上创建索引。
	CreateIndex(ctx context.Context, indexDef *model.IndexDefinition) error

	// DropIndex drops an index from a table.
	// DropIndex 从表中删除索引。
	DropIndex(ctx context.Context, databaseName, tableName, indexName string) error

	// ListIndexes retrieves all index definitions for a specific table.
	// ListIndexes 检索特定表的所有索引定义。
	ListIndexes(ctx context.Context, databaseName, tableName string) ([]*model.IndexDefinition, error)

	// CreateMaterializedView creates a new materialized view.
	// CreateMaterializedView 创建一个新的物化视图。
	CreateMaterializedView(ctx context.Context, mvDef *model.MaterializedViewDefinition) error

	// DropMaterializedView drops an existing materialized view.
	// DropMaterializedView 删除一个已存在的物化视图。
	DropMaterializedView(ctx context.Context, databaseName, viewName string) error

	// ListMaterializedViews lists materialized views in a database.
	// ListMaterializedViews 列出数据库中的物化视图。
	ListMaterializedViews(ctx context.Context, databaseName string, pagination *commontypes.PaginationRequest) (views []*model.MaterializedViewDefinition, total int64, err error)

	// GetMaterializedViewDefinition retrieves the definition of a specific materialized view.
	// GetMaterializedViewDefinition 检索特定物化视图的定义。
	GetMaterializedViewDefinition(ctx context.Context, databaseName, viewName string) (*model.MaterializedViewDefinition, error)
}
