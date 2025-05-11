package ingestion

import (
	"context"

	"github.com/turtacn/dataseap/pkg/domain/ingestion/model"
)

// Service defines the interface for the data ingestion service.
// Service 定义了数据采集服务的接口。
type Service interface {
	// IngestEvents ingests one or more raw events into the system.
	// IngestEvents 将一个或多个原始事件采集到系统中。
	// It returns the count of successfully ingested events, successfully parsed but failed to persist events,
	// and events that failed parsing/validation.
	// 它返回成功采集的事件数量，成功解析但持久化失败的事件数量，以及解析/验证失败的事件数量。
	IngestEvents(ctx context.Context, events []*model.RawEvent) (ingestedCount int, persistFailedCount int, validationFailedCount int, err error)

	// IngestEvent ingests a single raw event. Convenience method for IngestEvents.
	// IngestEvent 采集单个原始事件。是对 IngestEvents 的便捷方法。
	IngestEvent(ctx context.Context, event *model.RawEvent) error
}
