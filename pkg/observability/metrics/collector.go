package metrics

import (
	"net/http"
	"sync"

	"github.com/turtacn/dataseap/pkg/adapter/monitoring" // Our monitoring adapter interface
	"github.com/turtacn/dataseap/pkg/logger"
)

// AppMetrics holds all application-specific metrics.
// AppMetrics 保存所有应用程序特定的度量指标。
type AppMetrics struct {
	// HTTP Metrics
	HTTPRequestTotal    monitoring.Counter   // http_requests_total (method, path, status_code)
	HTTPRequestDuration monitoring.Histogram // http_request_duration_seconds (method, path)
	HTTPActiveRequests  monitoring.Gauge     // http_active_requests

	// gRPC Metrics
	GRPCRequestTotal    monitoring.Counter   // grpc_requests_total (service, method, status_code)
	GRPCRequestDuration monitoring.Histogram // grpc_request_duration_seconds (service, method)

	// Ingestion Metrics
	EventsIngestedTotal monitoring.Counter   // events_ingested_total (data_type, status) status: success, validation_failed, persist_failed
	EventIngestionLag   monitoring.Histogram // event_ingestion_lag_seconds (data_type) (event_timestamp - processing_timestamp)

	// Database/Adapter Metrics
	StarRocksQueryDuration monitoring.Histogram // starrocks_query_duration_seconds (query_type, table)
	StarRocksErrorsTotal   monitoring.Counter   // starrocks_errors_total (operation_type, error_code)

	// Add other application-specific metrics here
	// ...

	// Exporter used to register these metrics
	exporter monitoring.MetricsExporter
}

var (
	globalAppMetrics *AppMetrics
	metricsOnce      sync.Once
)

// NewAppMetrics initializes and registers all application-specific metrics.
// NewAppMetrics 初始化并注册所有应用程序特定的度量指标。
// It should be called once at application startup.
// 它应该在应用程序启动时调用一次。
func NewAppMetrics(exporter monitoring.MetricsExporter) (*AppMetrics, error) {
	var err error
	metricsOnce.Do(func() {
		l := logger.L().With("component", "AppMetrics")
		if exporter == nil {
			err = logger.NewDomainError("MetricsExporter cannot be nil") // Using domain error for example
			l.Error(err.Error())
			return
		}

		m := &AppMetrics{exporter: exporter}

		// Register HTTP Metrics
		m.HTTPRequestTotal, err = exporter.RegisterCounter(
			"dataseap_http_requests_total",
			"Total number of HTTP requests.",
			"method", "path", "status_code",
		)
		if err != nil {
			l.Errorw("Failed to register http_requests_total", "error", err)
			return
		}

		m.HTTPRequestDuration, err = exporter.RegisterHistogram(
			"dataseap_http_request_duration_seconds",
			"Histogram of HTTP request latencies.",
			nil, // Default buckets
			"method", "path",
		)
		if err != nil {
			l.Errorw("Failed to register http_request_duration_seconds", "error", err)
			return
		}

		m.HTTPActiveRequests, err = exporter.RegisterGauge(
			"dataseap_http_active_requests",
			"Number of active HTTP requests.",
		)
		if err != nil {
			l.Errorw("Failed to register http_active_requests", "error", err)
			return
		}

		// Register gRPC Metrics
		m.GRPCRequestTotal, err = exporter.RegisterCounter(
			"dataseap_grpc_requests_total",
			"Total number of gRPC requests.",
			"service", "method", "status_code",
		)
		if err != nil {
			l.Errorw("Failed to register grpc_requests_total", "error", err)
			return
		}

		m.GRPCRequestDuration, err = exporter.RegisterHistogram(
			"dataseap_grpc_request_duration_seconds",
			"Histogram of gRPC request latencies.",
			nil, // Default buckets
			"service", "method",
		)
		if err != nil {
			l.Errorw("Failed to register grpc_request_duration_seconds", "error", err)
			return
		}

		// Register Ingestion Metrics
		m.EventsIngestedTotal, err = exporter.RegisterCounter(
			"dataseap_events_ingested_total",
			"Total number of ingested events by data type and status.",
			"data_type", "status", // status: success, validation_failed, persist_failed
		)
		if err != nil {
			l.Errorw("Failed to register events_ingested_total", "error", err)
			return
		}

		m.EventIngestionLag, err = exporter.RegisterHistogram(
			"dataseap_event_ingestion_lag_seconds",
			"Lag between event timestamp and processing timestamp.",
			prometheus.ExponentialBuckets(0.1, 2, 15), // Example: 0.1s to ~9 hours
			"data_type",
		)
		if err != nil {
			l.Errorw("Failed to register event_ingestion_lag_seconds", "error", err)
			return
		}

		// Register StarRocks/Database Metrics
		m.StarRocksQueryDuration, err = exporter.RegisterHistogram(
			"dataseap_starrocks_query_duration_seconds",
			"Histogram of StarRocks query latencies.",
			nil,                   // Default buckets
			"query_type", "table", // query_type: "select", "stream_load", "ddl"
		)
		if err != nil {
			l.Errorw("Failed to register starrocks_query_duration_seconds", "error", err)
			return
		}

		m.StarRocksErrorsTotal, err = exporter.RegisterCounter(
			"dataseap_starrocks_errors_total",
			"Total number of StarRocks operation errors.",
			"operation_type", "error_code_group", // error_code_group: "connection", "timeout", "query_fail"
		)
		if err != nil {
			l.Errorw("Failed to register starrocks_errors_total", "error", err)
			return
		}

		// ... Register other metrics ...

		if err != nil {
			// If any registration failed, set globalAppMetrics to nil
			globalAppMetrics = nil
			l.Errorw("One or more application metrics failed to register", "final_error", err)
			return
		}

		globalAppMetrics = m
		l.Info("Application metrics registered successfully.")
	})
	if err != nil {
		return nil, err
	}
	return globalAppMetrics, nil
}

// Get returns the global AppMetrics instance.
// Get 返回全局的AppMetrics实例。
// It panics if NewAppMetrics has not been called or failed.
// 如果NewAppMetrics未被调用或失败，则会panic。
func Get() *AppMetrics {
	if globalAppMetrics == nil {
		// This indicates a programming error (NewAppMetrics not called or failed during startup).
		// For robustness, one might return a no-op AppMetrics implementation here instead of panicking.
		logger.L().Panic("AppMetrics not initialized. Call NewAppMetrics first.")
	}
	return globalAppMetrics
}

// ExposeHandler returns an http.Handler that can be used to expose metrics
// (e.g., for Prometheus scraping at /metrics).
// ExposeHandler 返回一个 http.Handler，可用于暴露度量指标。
func ExposeHandler() http.Handler {
	if globalAppMetrics == nil || globalAppMetrics.exporter == nil {
		logger.L().Error("AppMetrics or its exporter is not initialized, cannot expose metrics handler.")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Metrics system not initialized", http.StatusInternalServerError)
		})
	}
	return globalAppMetrics.exporter.ExposeHandler()
}

// ---- Example of how to use these metrics in code: ----
//
// func someHTTPHandler(w http.ResponseWriter, r *http.Request) {
//     startTime := time.Now()
//     metrics.Get().HTTPActiveRequests.Inc()
//     defer metrics.Get().HTTPActiveRequests.Dec()
//
//     // ... handle request ...
//     statusCode := http.StatusOK // or other status
//
//     metrics.Get().HTTPRequestTotal.With(r.Method, r.URL.Path, strconv.Itoa(statusCode)).Inc()
//     metrics.Get().HTTPRequestDuration.With(r.Method, r.URL.Path).Observe(time.Since(startTime).Seconds())
// }
//
// func ingestData(event model.RawEvent) {
//     lag := time.Since(event.Timestamp).Seconds()
//     metrics.Get().EventIngestionLag.With(event.DataType).Observe(lag)
//
//     err := persistToStarRocks(event)
//     status := "success"
//     if err != nil { status = "persist_failed" }
//     metrics.Get().EventsIngestedTotal.With(event.DataType, status).Inc()
// }
//
