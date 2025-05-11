package query

import (
	"context"

	"github.com/turtacn/dataseap/pkg/adapter/starrocks"
	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/domain/query/model"
	"github.com/turtacn/dataseap/pkg/logger"
	// metadataService "github.com/turtacn/dataseap/pkg/domain/management/metadata" // For schema info, etc.
)

// FullTextSearchSubService defines the interface for the full-text search specific logic.
// This helps in separating concerns within the query domain.
// FullTextSearchSubService 定义了全文检索特定逻辑的接口。
// 这有助于在查询领域内分离关注点。
type FullTextSearchSubService interface {
	Search(ctx context.Context, req *model.FullTextSearchRequest) (*model.FullTextSearchResult, error)
}

type serviceImpl struct {
	starrocksClient  starrocks.Client
	fullTextSearcher FullTextSearchSubService
	// metadataSvc      metadataService.Service // Optional: for query planning or validation
}

// NewService creates a new instance of the query service.
// NewService 创建一个新的查询服务实例。
func NewService(srClient starrocks.Client, ftSearcher FullTextSearchSubService /*, metaSvc metadataService.Service*/) Service {
	return &serviceImpl{
		starrocksClient:  srClient,
		fullTextSearcher: ftSearcher,
		// metadataSvc:      metaSvc,
	}
}

// ExecuteSQL executes a given SQL query and returns the results.
// ExecuteSQL 执行给定的SQL查询并返回结果。
func (s *serviceImpl) ExecuteSQL(ctx context.Context, req *model.SQLQueryRequest) (*model.SQLQueryResult, error) {
	l := logger.L().Ctx(ctx).With("method", "ExecuteSQL", "sql_query_length", len(req.SQL))
	l.Info("Attempting to execute SQL query")

	if err := req.Validate(); err != nil {
		l.Warnw("SQLQueryRequest validation failed", "error", err)
		return nil, errors.Wrap(err, errors.InvalidArgument, "invalid SQL query request")
	}

	// TODO: Add query sanitization or validation if necessary, though StarRocks client should handle parameters safely if supported.
	// For basic HTTP API, parameters are not directly supported, so SQL must be pre-formatted.

	// TODO: If req.Database is provided, ensure it's used. The current starrocksClient.Execute
	// might use a default database from its config or require specific handling.
	// Potentially use a temporary session property: SET DATABASE = req.Database;

	srResult, err := s.starrocksClient.Execute(ctx, req.SQL /*, req.Params - if supported */)
	if err != nil {
		l.Errorw("Failed to execute SQL query via StarRocks client", "error", err)
		return nil, errors.Wrap(err, errors.DatabaseError, "failed to execute SQL query")
	}

	// Transform starrocks.QueryResult to model.SQLQueryResult
	domainResult := &model.SQLQueryResult{
		Columns: srResult.Columns,
		Rows:    make([]map[string]interface{}, len(srResult.Rows)),
		// AffectedRows: srResult.AffectedRows, // Assuming starrocks.QueryResult has this
		// Stats: &model.QueryStats{...} // Map from srResult.Stats
		// Pagination: req.Pagination ... // This should be part of the result from adapter if it handles pagination
	}

	for i, srRow := range srResult.Rows {
		rowData := make(map[string]interface{})
		for j, colName := range srResult.Columns {
			if j < len(srRow) {
				rowData[colName] = srRow[j]
			}
		}
		domainResult.Rows[i] = rowData
	}

	if srResult.Stats != nil {
		domainResult.Stats = &model.QueryStats{
			ScanRows:   srResult.Stats.ScanRows,
			ScanBytes:  srResult.Stats.ScanBytes,
			Duration:   srResult.Stats.Duration,
			PeakMemory: srResult.Stats.PeakMemory,
			CPUTime:    srResult.Stats.CPUTime,
			Message:    srResult.Stats.Message,
		}
	}

	l.Info("SQL query executed successfully")
	return domainResult, nil
}

// SearchFullText performs a full-text search based on the provided request.
// SearchFullText 根据提供的请求执行全文检索。
func (s *serviceImpl) SearchFullText(ctx context.Context, req *model.FullTextSearchRequest) (*model.FullTextSearchResult, error) {
	l := logger.L().Ctx(ctx).With("method", "SearchFullText", "keywords", req.Keywords)
	l.Info("Attempting to perform full-text search")

	if err := req.Validate(); err != nil {
		l.Warnw("FullTextSearchRequest validation failed", "error", err)
		return nil, errors.Wrap(err, errors.InvalidArgument, "invalid full-text search request")
	}

	if s.fullTextSearcher == nil {
		l.Error("FullTextSearchSubService is not initialized")
		return nil, errors.New(errors.InternalError, "FullTextSearchSubService is not available")
	}

	result, err := s.fullTextSearcher.Search(ctx, req)
	if err != nil {
		l.Errorw("Full-text search failed", "error", err)
		// Error already wrapped by sub-service or adapter
		return nil, err
	}

	l.Info("Full-text search completed successfully")
	return result, nil
}
