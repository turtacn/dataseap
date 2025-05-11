package ingestion

import (
	"context"
	"fmt"

	"github.com/turtacn/dataseap/pkg/adapter/starrocks" // StarRocks adapter for data persistence
	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/domain/ingestion/model"
	"github.com/turtacn/dataseap/pkg/logger"
	// "github.com/turtacn/dataseap/pkg/adapter/pulsar" // Optional: Pulsar adapter for message queue
)

type serviceImpl struct {
	starrocksClient starrocks.Client
	// pulsarProducer pulsar.Producer // Optional: if data is first sent to a message queue
	// validator       some_validation_package.Validator // Optional: for complex validation logic
}

// NewService creates a new instance of the ingestion service.
// NewService 创建一个新的采集服务实例。
func NewService(srClient starrocks.Client /*, pulsarProd pulsar.Producer */) Service {
	return &serviceImpl{
		starrocksClient: srClient,
		// pulsarProducer: pulsarProd,
	}
}

// IngestEvents ingests one or more raw events into the system.
// IngestEvents 将一个或多个原始事件采集到系统中。
func (s *serviceImpl) IngestEvents(ctx context.Context, events []*model.RawEvent) (ingestedCount int, persistFailedCount int, validationFailedCount int, err error) {
	l := logger.L().Ctx(ctx).With("method", "IngestEvents", "event_count", len(events))
	l.Info("Attempting to ingest events")

	if len(events) == 0 {
		l.Info("No events to ingest")
		return 0, 0, 0, nil
	}

	// Placeholder implementation:
	// 1. Validate events
	// 2. Transform events if necessary
	// 3. Batch events
	// 4. Send to StarRocks (or Pulsar first)

	for _, event := range events {
		if err := event.Validate(); err != nil {
			l.Warnw("Event validation failed", "event_id", event.ID, "data_source_id", event.DataSourceID, "error", err)
			validationFailedCount++
			continue // Skip this event or collect errors
		}

		// TODO: Implement actual data ingestion logic using s.starrocksClient.StreamLoad or similar
		// For now, simulate ingestion
		l.Debugw("Simulating ingestion for event", "event_id", event.ID, "data_type", event.DataType)
		// Example:
		// jsonData, _ := json.Marshal(event.Data)
		// streamLoadOpts := &starrocks.StreamLoadOptions{Format: "json"}
		// _, err := s.starrocksClient.StreamLoad(ctx, "your_database", event.DataType /*table name*/, bytes.NewReader(jsonData), streamLoadOpts)
		// if err != nil {
		//     l.Errorw("Failed to persist event to StarRocks", "event_id", event.ID, "error", err)
		//     persistFailedCount++
		//     continue
		// }

		// Simulate success for skeleton
		ingestedCount++
	}

	if validationFailedCount > 0 || persistFailedCount > 0 {
		l.Warnw("Some events failed during ingestion process",
			"total", len(events),
			"succeeded", ingestedCount,
			"validation_failed", validationFailedCount,
			"persist_failed", persistFailedCount,
		)
		// Decide on error return strategy. If some succeed, is it still an overall error?
		// For now, return a generic error if any failures occurred.
		if persistFailedCount > 0 {
			return ingestedCount, persistFailedCount, validationFailedCount, errors.New(errors.DatabaseError, fmt.Sprintf("%d events failed to persist", persistFailedCount))
		}
		if validationFailedCount > 0 {
			return ingestedCount, persistFailedCount, validationFailedCount, errors.New(errors.InvalidArgument, fmt.Sprintf("%d events failed validation", validationFailedCount))
		}
	}

	l.Infow("Ingestion process completed", "succeeded", ingestedCount, "validation_failed", validationFailedCount, "persist_failed", persistFailedCount)
	return ingestedCount, persistFailedCount, validationFailedCount, nil
}

// IngestEvent ingests a single raw event.
// IngestEvent 采集单个原始事件。
func (s *serviceImpl) IngestEvent(ctx context.Context, event *model.RawEvent) error {
	l := logger.L().Ctx(ctx).With("method", "IngestEvent", "event_id", event.ID)
	l.Info("Attempting to ingest a single event")

	_, _, validationFailed, err := s.IngestEvents(ctx, []*model.RawEvent{event})
	if err != nil {
		return err
	}
	if validationFailed > 0 {
		return errors.New(errors.InvalidArgument, "event validation failed")
	}
	// If IngestEvents returns an error when persistFailed > 0, that error is already propagated.
	return nil
}
