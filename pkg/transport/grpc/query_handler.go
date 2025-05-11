package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/turtacn/dataseap/api/v1"
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	"github.com/turtacn/dataseap/pkg/domain/query"
	querymodel "github.com/turtacn/dataseap/pkg/domain/query/model"
	"github.com/turtacn/dataseap/pkg/logger"
)

// queryHandler implements the apiv1.QueryServiceServer interface.
// queryHandler 实现 apiv1.QueryServiceServer 接口。
type queryHandler struct {
	apiv1.UnimplementedQueryServiceServer // For forward compatibility
	domainService                         query.Service
}

// NewQueryHandler creates a new gRPC handler for the query service.
// NewQueryHandler 为查询服务创建一个新的gRPC处理器。
func NewQueryHandler(service query.Service) apiv1.QueryServiceServer {
	return &queryHandler{
		domainService: service,
	}
}

// ExecuteSQLQuery handles incoming SQL query requests.
// ExecuteSQLQuery 处理传入的SQL查询请求。
func (h *queryHandler) ExecuteSQLQuery(ctx context.Context, req *apiv1.ExecuteSQLQueryRequest) (*apiv1.ExecuteSQLQueryResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "ExecuteSQLQuery", "request_id", req.GetRequestId())
	l.Info("Received ExecuteSQLQuery request")

	domainReq := &querymodel.SQLQueryRequest{
		SQL:              req.GetSqlQuery(),
		Params:           make(map[string]interface{}), // TODO: Convert req.GetParameters() if it's map<string, Value>
		WorkloadGroup:    req.GetWorkloadGroup(),
		QueryTimeoutSecs: int(req.GetQueryTimeoutSeconds()),
		// Database: req.GetDatabase(), // If database is added to proto
	}
	if req.GetParameters() != nil {
		for k, v := range req.GetParameters() {
			domainReq.Params[k] = v.AsInterface()
		}
	}

	if req.GetPagination() != nil {
		domainReq.Pagination = &commontypes.PaginationRequest{
			Page:     int(req.GetPagination().GetPage()),
			PageSize: int(req.GetPagination().GetPageSize()),
		}
	}

	result, err := h.domainService.ExecuteSQL(ctx, domainReq)
	if err != nil {
		l.Errorw("Query service ExecuteSQL returned an error", "error", err)
		// TODO: Map domain error to gRPC status code
		return &apiv1.ExecuteSQLQueryResponse{
			Success: false,
			Message: err.Error(),
			Error:   toProtoErrorDetail("SQL_EXECUTION_ERROR", err.Error()),
		}, status.Error(codes.Internal, err.Error()) // Or more specific code
	}

	// Map domain result to gRPC response
	rows := make([]*apiv1.DataRow, len(result.Rows))
	for i, domainRowMap := range result.Rows {
		pbStruct, err := structpb.NewStruct(domainRowMap)
		if err != nil {
			l.Errorw("Failed to convert domain row to protobuf struct", "error", err)
			return &apiv1.ExecuteSQLQueryResponse{
				Success: false,
				Message: "Failed to process query results",
				Error:   toProtoErrorDetail("RESULT_PROCESSING_ERROR", err.Error()),
			}, status.Error(codes.Internal, "failed to process query results")
		}
		rows[i] = &apiv1.DataRow{Fields: pbStruct}
	}

	resp := &apiv1.ExecuteSQLQueryResponse{
		Success:      true,
		Message:      "Query executed successfully",
		ColumnNames:  result.Columns,
		Rows:         rows,
		AffectedRows: result.AffectedRows,
	}
	if result.Pagination != nil {
		resp.Pagination = &apiv1.PaginationResponse{
			Page:       int32(result.Pagination.Page),
			PageSize:   int32(result.Pagination.PageSize),
			TotalItems: result.Pagination.Total, // Assuming PaginationResponse has TotalItems
		}
	}
	// TODO: Map result.Stats to a proto message if defined

	l.Info("ExecuteSQLQuery request processed successfully")
	return resp, nil
}

// FullTextSearch handles incoming full-text search requests.
// FullTextSearch 处理传入的全文检索请求。
func (h *queryHandler) FullTextSearch(ctx context.Context, req *apiv1.FullTextSearchRequest) (*apiv1.FullTextSearchResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "FullTextSearch", "request_id", req.GetRequestId(), "keywords", req.GetKeywords())
	l.Info("Received FullTextSearch request")

	domainReq := &querymodel.FullTextSearchRequest{
		Keywords:       req.GetKeywords(),
		TargetTables:   req.GetTargetTables(),
		TargetFields:   req.GetTargetFields(),
		Tokenizer:      req.GetTokenizer(),
		RecallPriority: req.GetRecallPriority(),
		// AdditionalFilters: (map if present in proto)
	}
	if req.GetPagination() != nil {
		domainReq.Pagination = &commontypes.PaginationRequest{
			Page:     int(req.GetPagination().GetPage()),
			PageSize: int(req.GetPagination().GetPageSize()),
		}
	}
	if req.GetTimeRangeFilter() != nil {
		domainReq.TimeRangeFilter = &commontypes.TimeRange{
			StartTime: req.GetTimeRangeFilter().GetStartTime().AsTime(),
			EndTime:   req.GetTimeRangeFilter().GetEndTime().AsTime(),
		}
	}
	// Map SortBy if added to proto

	result, err := h.domainService.SearchFullText(ctx, domainReq)
	if err != nil {
		l.Errorw("Query service SearchFullText returned an error", "error", err)
		return &apiv1.FullTextSearchResponse{
			Success: false,
			Message: err.Error(),
			Error:   toProtoErrorDetail("FULLTEXT_SEARCH_ERROR", err.Error()),
		}, status.Error(codes.Internal, err.Error()) // Or more specific
	}

	// Map domain result to gRPC response
	hits := make([]*apiv1.SearchHit, len(result.Hits))
	for i, domainHit := range result.Hits {
		docPbStruct, err := structpb.NewStruct(domainHit.Document)
		if err != nil {
			l.Errorw("Failed to convert domain hit document to protobuf struct", "error", err)
			// continue or return error
		}
		hits[i] = &apiv1.SearchHit{
			SourceTable: domainHit.SourceTable,
			HitFields:   domainHit.HitFields,
			Score:       domainHit.Score,
			Document:    &apiv1.DataRow{Fields: docPbStruct},
			Id:          domainHit.ID,
		}
	}

	resp := &apiv1.FullTextSearchResponse{
		Success: true,
		Message: "Full-text search completed successfully",
		Hits:    hits,
	}
	if result.Pagination != nil { // Assuming domain result includes pagination response
		resp.Pagination = &apiv1.PaginationResponse{
			Page:       int32(result.Pagination.Page),
			PageSize:   int32(result.Pagination.PageSize),
			TotalItems: result.TotalHits, // Using TotalHits from domain model
		}
	}

	l.Info("FullTextSearch request processed successfully")
	return resp, nil
}

// Helper function to map domain TimeRange to proto TimeRange
func toProtoTimeRange(tr *commontypes.TimeRange) *apiv1.TimeRange {
	if tr == nil {
		return nil
	}
	return &apiv1.TimeRange{
		StartTime: timestamppb.New(tr.StartTime),
		EndTime:   timestamppb.New(tr.EndTime),
	}
}
