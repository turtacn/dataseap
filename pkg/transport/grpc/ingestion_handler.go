package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/turtacn/dataseap/api/v1"
	"github.com/turtacn/dataseap/pkg/domain/ingestion"
	ingestionmodel "github.com/turtacn/dataseap/pkg/domain/ingestion/model"
	"github.com/turtacn/dataseap/pkg/logger"
)

// ingestionHandler implements the apiv1.IngestionServiceServer interface.
// ingestionHandler 实现 apiv1.IngestionServiceServer 接口。
type ingestionHandler struct {
	apiv1.UnimplementedIngestionServiceServer // For forward compatibility
	domainService                             ingestion.Service
}

// NewIngestionHandler creates a new gRPC handler for the ingestion service.
// NewIngestionHandler 为采集服务创建一个新的gRPC处理器。
func NewIngestionHandler(service ingestion.Service) apiv1.IngestionServiceServer {
	return &ingestionHandler{
		domainService: service,
	}
}

// IngestData handles incoming data ingestion requests.
// IngestData 处理传入的数据采集请求。
func (h *ingestionHandler) IngestData(ctx context.Context, req *apiv1.IngestDataRequest) (*apiv1.IngestDataResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "IngestData", "request_id", req.GetRequestId(), "num_records", len(req.GetRecords()))
	l.Info("Received IngestData request")

	if len(req.GetRecords()) == 0 {
		l.Info("No records to ingest")
		return &apiv1.IngestDataResponse{Success: true, Message: "No records provided"}, nil
	}

	domainEvents := make([]*ingestionmodel.RawEvent, len(req.GetRecords()))
	for i, r := range req.GetRecords() {
		var eventData map[string]interface{}
		if r.GetData() != nil {
			eventData = r.GetData().AsMap()
		}

		var ts time.Time
		if r.GetTimestamp() != nil && r.GetTimestamp().IsValid() {
			ts = r.GetTimestamp().AsTime()
		} else {
			ts = time.Now().UTC() // Default to now if not provided or invalid
		}

		domainEvents[i] = &ingestionmodel.RawEvent{
			// ID: (if provided in proto, map here)
			DataSourceID: r.GetDataSourceId(),
			DataType:     r.GetDataType(),
			Timestamp:    ts,
			Data:         eventData,
			// RawPayload: (if provided in proto)
			Tags: r.GetTags(),
		}
	}

	ingested, persistFailed, validationFailed, err := h.domainService.IngestEvents(ctx, domainEvents)
	if err != nil {
		l.Errorw("Ingestion service returned an error", "error", err)
		// Map domain error to gRPC status
		// TODO: More sophisticated error mapping based on err type
		st := status.New(codes.Internal, "Ingestion failed")
		// if errors.Is(err, custom_errors.ValidationError) {
		// 	st = status.New(codes.InvalidArgument, err.Error())
		// } else if errors.Is(err, custom_errors.DatabaseError) {
		//  st = status.New(codes.Unavailable, "Failed to persist data")
		// }
		// For now, simple mapping:
		detailedError, _ := st.WithDetails(&apiv1.ErrorDetail{Code: "INGESTION_ERROR", Message: err.Error()})
		if detailedError != nil {
			st = detailedError
		}
		return &apiv1.IngestDataResponse{
			Success:       false,
			IngestedCount: int64(ingested),
			FailedCount:   int64(persistFailed + validationFailed),
			ErrorMessage:  err.Error(),
			ErrorCode:     "INGESTION_ERROR", // Or more specific code
		}, st.Err()
	}

	l.Infow("Ingestion successful", "ingested", ingested, "persist_failed", persistFailed, "validation_failed", validationFailed)
	return &apiv1.IngestDataResponse{
		Success:       true,
		IngestedCount: int64(ingested),
		FailedCount:   int64(persistFailed + validationFailed), // Sum of failures
		Message:       "Data ingestion processed.",
	}, nil
}

// Helper to map proto ErrorDetail to gRPC status detail
func toProtoErrorDetail(code, message string) *apiv1.ErrorDetail {
	return &apiv1.ErrorDetail{
		Code:    code,
		Message: message,
	}
}
