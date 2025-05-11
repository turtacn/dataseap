package query

import (
	"context"

	"github.com/turtacn/dataseap/pkg/domain/query/model"
)

// Service defines the interface for the data query service.
// Service 定义了数据查询服务的接口。
type Service interface {
	// ExecuteSQL executes a given SQL query and returns the results.
	// ExecuteSQL 执行给定的SQL查询并返回结果。
	ExecuteSQL(ctx context.Context, req *model.SQLQueryRequest) (*model.SQLQueryResult, error)

	// SearchFullText performs a full-text search based on the provided request.
	// SearchFullText 根据提供的请求执行全文检索。
	SearchFullText(ctx context.Context, req *model.FullTextSearchRequest) (*model.FullTextSearchResult, error)

	// TODO: Add other query capabilities as needed, e.g.,
	// GetAggregatedData(ctx context.Context, aggRequest *model.AggregationRequest) (*model.AggregationResult, error)
}
