package tracing

import (
	"context"
	"time"

	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Use a recent stable version
	"go.opentelemetry.io/otel/trace"
)

// Config holds the configuration for the tracer provider.
// Config 保存追踪提供程序的配置。
type Config struct {
	Enabled        bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	ExporterType   string  `mapstructure:"exporterType" json:"exporterType" yaml:"exporterType"` // "otlp", "stdout", "jaeger", "zipkin" (otlp is preferred)
	OTLPEndpoint   string  `mapstructure:"otlpEndpoint" json:"otlpEndpoint" yaml:"otlpEndpoint"` // e.g., "localhost:4317" for gRPC
	OTLPInsecure   bool    `mapstructure:"otlpInsecure" json:"otlpInsecure" yaml:"otlpInsecure"` // For OTLP gRPC, use insecure connection
	Sampler        string  `mapstructure:"sampler" json:"sampler" yaml:"sampler"`                // "always", "never", "parentbased_always", "traceidratio"
	SampleRatio    float64 `mapstructure:"sampleRatio" json:"sampleRatio" yaml:"sampleRatio"`    // For traceidratio sampler
	ServiceName    string  `mapstructure:"serviceName" json:"serviceName" yaml:"serviceName"`
	ServiceVersion string  `mapstructure:"serviceVersion" json:"serviceVersion" yaml:"serviceVersion"`
}

var tracerProvider *sdktrace.TracerProvider

// DefaultTracerConfig returns a default configuration for tracing.
// DefaultTracerConfig 返回追踪的默认配置。
func DefaultTracerConfig() Config {
	return Config{
		Enabled:        false, // Disabled by default
		ExporterType:   "stdout",
		OTLPEndpoint:   "localhost:4317",
		OTLPInsecure:   true,
		Sampler:        "parentbased_always",
		SampleRatio:    1.0,
		ServiceName:    constants.ServiceName,
		ServiceVersion: constants.ServiceVersion,
	}
}

// InitTracerProvider initializes the OpenTelemetry tracer provider.
// InitTracerProvider 初始化OpenTelemetry追踪提供程序。
// It should be called once at application startup.
// 它应该在应用程序启动时调用一次。
func InitTracerProvider(cfg Config) (shutdown func(context.Context) error, err error) {
	l := logger.L().With("component", "TracerProvider")
	if !cfg.Enabled {
		l.Info("Distributed tracing is disabled.")
		// Set a NoOpTracerProvider if tracing is disabled
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	var exporter sdktrace.SpanExporter
	switch cfg.ExporterType {
	case "otlp":
		ctx := context.Background()
		clientOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint)}
		if cfg.OTLPInsecure {
			clientOpts = append(clientOpts, otlptracegrpc.WithInsecure())
		} else {
			// TODO: Add TLS credentials if OTLPInsecure is false
			// clientOpts = append(clientOpts, otlptracegrpc.WithTLSCredentials(...))
		}
		exporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient(clientOpts...))
		if err != nil {
			l.Errorw("Failed to create OTLP trace exporter", "endpoint", cfg.OTLPEndpoint, "error", err)
			return nil, err
		}
		l.Infow("Using OTLP trace exporter", "endpoint", cfg.OTLPEndpoint)
	case "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			l.Errorw("Failed to create stdout trace exporter", "error", err)
			return nil, err
		}
		l.Info("Using stdout trace exporter")
	// TODO: Add Jaeger, Zipkin exporters if needed
	default:
		l.Warnw("Unsupported tracer exporter type, defaulting to stdout", "type", cfg.ExporterType)
		exporter, _ = stdouttrace.New(stdouttrace.WithPrettyPrint()) // Default to stdout
	}

	res, err := resource.New(context.Background(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			// Add other common resource attributes here
		),
	)
	if err != nil {
		l.Errorw("Failed to create OpenTelemetry resource", "error", err)
		return nil, err
	}

	var sampler sdktrace.Sampler
	switch cfg.Sampler {
	case "always":
		sampler = sdktrace.AlwaysSample()
	case "never":
		sampler = sdktrace.NeverSample()
	case "traceidratio":
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRatio)
	case "parentbased_always":
		fallthrough
	default:
		sampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
	l.Infow("Using sampler", "type", cfg.Sampler, "ratio", cfg.SampleRatio)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tracerProvider)
	// Set up W3C TraceContext and Baggage propagators
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	l.Info("OpenTelemetry TracerProvider initialized successfully.")

	shutdown = func(ctx context.Context) error {
		l.Info("Shutting down TracerProvider...")
		// Give it a bit of time to flush remaining spans
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tracerProvider.Shutdown(shutdownCtx)
	}
	return shutdown, nil
}

// GetTracer returns a named tracer from the global TracerProvider.
// GetTracer 从全局TracerProvider返回一个命名的追踪器。
// If the provider is not initialized, it returns a NoOpTracer.
// 如果提供程序未初始化，则返回NoOpTracer。
func GetTracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if tracerProvider == nil {
		// This case should ideally not happen if InitTracerProvider is called at startup.
		// Fallback to global otel provider which might be NoOp if not initialized.
		logger.L().Warn("TracerProvider not initialized, using default OpenTelemetry provider (might be NoOp).")
		return otel.Tracer(name, opts...)
	}
	return tracerProvider.Tracer(name, opts...)
}

// StartSpan starts a new span from the given context.
// StartSpan 从给定的上下文中开始一个新的span。
// It's a utility function for convenience.
// 这是一个方便使用的工具函数。
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Assuming the tracer name is the service name or a common instrumentation name
	tracer := GetTracer(constants.ServiceName + "/instrumentation")
	return tracer.Start(ctx, spanName, opts...)
}
