package monitoring

import "net/http"

// Counter is a metric that can only be incremented.
// Counter 是一个只能递增的度量指标。
type Counter interface {
	Inc()
	Add(float64)
	// With returns a Counter with the given label values.
	// With 返回带有给定标签值的Counter。
	// The number of label values must match the number of labels defined during registration.
	// 标签值的数量必须与注册时定义的标签数量相匹配。
	With(labelValues ...string) Counter
}

// Gauge is a metric that can be set to any value.
// Gauge 是一个可以设置为任意值的度量指标。
type Gauge interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
	// With returns a Gauge with the given label values.
	// With 返回带有给定标签值的Gauge。
	With(labelValues ...string) Gauge
}

// Histogram is a metric that samples observations (usually things like request durations or response sizes)
// and counts them in configurable buckets. It also provides a sum of all observed values.
// Histogram 是一个对观察值（通常是请求持续时间或响应大小等）进行采样并在可配置的桶中计数的度量指标。
// 它还提供所有观察值的总和。
type Histogram interface {
	Observe(float64)
	// With returns a Histogram with the given label values.
	// With 返回带有给定标签值的Histogram。
	With(labelValues ...string) Histogram
}

// Summary is similar to a Histogram but calculates configurable φ-quantiles (e.g., 0.5, 0.9, 0.99)
// over a sliding time window.
// Summary 类似于Histogram，但在滑动时间窗口内计算可配置的φ-分位数（例如0.5、0.9、0.99）。
type Summary interface {
	Observe(float64)
	// With returns a Summary with the given label values.
	// With 返回带有给定标签值的Summary。
	With(labelValues ...string) Summary
}

// MetricsExporter defines the interface for a metrics system adapter (e.g., Prometheus).
// MetricsExporter 定义了度量系统适配器（例如Prometheus）的接口。
type MetricsExporter interface {
	// RegisterCounter creates and registers a new Counter.
	// RegisterCounter 创建并注册一个新的Counter。
	// Labels are optional and allow for dimensionality.
	// 标签是可选的，允许维度化。
	RegisterCounter(name, help string, labels ...string) (Counter, error)

	// RegisterGauge creates and registers a new Gauge.
	// RegisterGauge 创建并注册一个新的Gauge。
	RegisterGauge(name, help string, labels ...string) (Gauge, error)

	// RegisterHistogram creates and registers a new Histogram.
	// RegisterHistogram 创建并注册一个新的Histogram。
	// Buckets define the observation buckets for the histogram.
	// Buckets 定义直方图的观察桶。
	RegisterHistogram(name, help string, buckets []float64, labels ...string) (Histogram, error)

	// RegisterSummary creates and registers a new Summary.
	// RegisterSummary 创建并注册一个新的Summary。
	// Objectives map quantiles to allowed absolute errors (e.g., {0.5: 0.05, 0.9: 0.01}).
	// Objectives 将分位数映射到允许的绝对误差 (例如, {0.5: 0.05, 0.9: 0.01})。
	RegisterSummary(name, help string, objectives map[float64]float64, labels ...string) (Summary, error)

	// ExposeHandler returns an http.Handler that can be used to expose metrics
	// (e.g., for Prometheus scraping at /metrics).
	// ExposeHandler 返回一个 http.Handler，可用于暴露度量指标
	// (例如，供Prometheus在/metrics处抓取)。
	ExposeHandler() http.Handler
}
