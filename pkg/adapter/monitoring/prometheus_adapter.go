package monitoring

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/logger"
)

// prometheusCounter implements the Counter interface using Prometheus.
// prometheusCounter 使用Prometheus实现Counter接口。
type prometheusCounter struct {
	vec *prometheus.CounterVec // Underlying Prometheus CounterVec for labels
	// If no labels, it's a simple prometheus.Counter, but Vec is more general.
	// prometheus.Counter // For no labels case
}

func (pc *prometheusCounter) Inc() {
	if pc.vec != nil {
		// This Inc is for a Counter without specific labels filled in yet.
		// It's often an error to call Inc on a Vec directly without WithLabelValues.
		// However, if the vec was created with 0 labels, this is fine.
		// For labeled counters, use With().Inc().
		// For simplicity, if no labels expected, prometheus.NewCounter is used.
		// This adapter assumes labels might exist.
		logger.L().Warn("Inc called on prometheusCounter without specifying labels. This is a no-op if labels were defined. Use With(labelValues).Inc().")
		return
	}
	// If pc.counter is used for no-label case: pc.counter.Inc()
}

func (pc *prometheusCounter) Add(val float64) {
	if pc.vec != nil {
		logger.L().Warn("Add called on prometheusCounter without specifying labels. This is a no-op if labels were defined. Use With(labelValues).Add().")
		return
	}
	// If pc.counter is used: pc.counter.Add(val)
}

func (pc *prometheusCounter) With(labelValues ...string) Counter {
	if pc.vec == nil {
		logger.L().Error("Prometheus CounterVec is nil, cannot apply labels.")
		return &noopCounter{} // Return a no-op counter to prevent panic
	}
	return &prometheusCounterImpl{counter: pc.vec.WithLabelValues(labelValues...)}
}

// prometheusCounterImpl is the actual implementation after labels are applied.
type prometheusCounterImpl struct {
	counter prometheus.Counter
}

func (pci *prometheusCounterImpl) Inc()                   { pci.counter.Inc() }
func (pci *prometheusCounterImpl) Add(val float64)        { pci.counter.Add(val) }
func (pci *prometheusCounterImpl) With(...string) Counter { /* Already labeled */ return pci }

// prometheusGauge implements the Gauge interface using Prometheus.
// prometheusGauge 使用Prometheus实现Gauge接口。
type prometheusGauge struct {
	vec *prometheus.GaugeVec
}

func (pg *prometheusGauge) Set(val float64) {
	logger.L().Warn("Set called on prometheusGauge without labels.")
}
func (pg *prometheusGauge) Inc() { logger.L().Warn("Inc called on prometheusGauge without labels.") }
func (pg *prometheusGauge) Dec() { logger.L().Warn("Dec called on prometheusGauge without labels.") }
func (pg *prometheusGauge) Add(val float64) {
	logger.L().Warn("Add called on prometheusGauge without labels.")
}
func (pg *prometheusGauge) Sub(val float64) {
	logger.L().Warn("Sub called on prometheusGauge without labels.")
}
func (pg *prometheusGauge) With(labelValues ...string) Gauge {
	if pg.vec == nil {
		logger.L().Error("Prometheus GaugeVec is nil, cannot apply labels.")
		return &noopGauge{}
	}
	return &prometheusGaugeImpl{gauge: pg.vec.WithLabelValues(labelValues...)}
}

type prometheusGaugeImpl struct {
	gauge prometheus.Gauge
}

func (pgi *prometheusGaugeImpl) Set(val float64)      { pgi.gauge.Set(val) }
func (pgi *prometheusGaugeImpl) Inc()                 { pgi.gauge.Inc() }
func (pgi *prometheusGaugeImpl) Dec()                 { pgi.gauge.Dec() }
func (pgi *prometheusGaugeImpl) Add(val float64)      { pgi.gauge.Add(val) }
func (pgi *prometheusGaugeImpl) Sub(val float64)      { pgi.gauge.Sub(val) }
func (pgi *prometheusGaugeImpl) With(...string) Gauge { return pgi }

// prometheusHistogram implements the Histogram interface using Prometheus.
// prometheusHistogram 使用Prometheus实现Histogram接口。
type prometheusHistogram struct {
	vec *prometheus.HistogramVec
}

func (ph *prometheusHistogram) Observe(val float64) {
	logger.L().Warn("Observe called on prometheusHistogram without labels.")
}
func (ph *prometheusHistogram) With(labelValues ...string) Histogram {
	if ph.vec == nil {
		logger.L().Error("Prometheus HistogramVec is nil, cannot apply labels.")
		return &noopHistogram{}
	}
	return &prometheusHistogramImpl{histogram: ph.vec.WithLabelValues(labelValues...)}
}

type prometheusHistogramImpl struct {
	histogram prometheus.Histogram
}

func (phi *prometheusHistogramImpl) Observe(val float64)      { phi.histogram.Observe(val) }
func (phi *prometheusHistogramImpl) With(...string) Histogram { return phi }

// prometheusSummary implements the Summary interface using Prometheus.
type prometheusSummary struct {
	vec *prometheus.SummaryVec
}

func (ps *prometheusSummary) Observe(val float64) {
	logger.L().Warn("Observe called on prometheusSummary without labels.")
}
func (ps *prometheusSummary) With(labelValues ...string) Summary {
	if ps.vec == nil {
		logger.L().Error("Prometheus SummaryVec is nil, cannot apply labels.")
		return &noopSummary{}
	}
	return &prometheusSummaryImpl{summary: ps.vec.WithLabelValues(labelValues...)}
}

type prometheusSummaryImpl struct {
	summary prometheus.Summary
}

func (psi *prometheusSummaryImpl) Observe(val float64)    { psi.summary.Observe(val) }
func (psi *prometheusSummaryImpl) With(...string) Summary { return psi }

// prometheusExporter implements MetricsExporter using Prometheus.
// prometheusExporter 使用Prometheus实现MetricsExporter。
type prometheusExporter struct {
	registry *prometheus.Registry
	mu       sync.Mutex // To protect registration of metrics
}

// NewPrometheusExporter creates a new Prometheus metrics exporter.
// NewPrometheusExporter 创建一个新的Prometheus度量指标导出器。
func NewPrometheusExporter() (MetricsExporter, error) {
	reg := prometheus.NewRegistry()
	// Register default Go metrics and process metrics.
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return &prometheusExporter{
		registry: reg,
	}, nil
}

func (pe *prometheusExporter) RegisterCounter(name, help string, labels ...string) (Counter, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	if err := pe.registry.Register(counterVec); err != nil {
		// Handle already registered error specifically
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// Metric already registered, try to use the existing one.
			// This requires a way to fetch the existing metric, or assume it's compatible.
			// For simplicity, we'll return an error here or a wrapped existing one if possible.
			// This scenario needs careful handling in a robust system.
			logger.L().Warnw("Counter metric already registered", "name", name)
			// Potentially return a wrapper around the existing metric if fetchable,
			// or just return the new one and let Prometheus handle it (it might panic or error on re-registration).
			// The safe bet is to ensure names are unique or manage this state.
			// For now, we return the error.
		}
		return nil, errors.Wrapf(err, errors.InternalError, "failed to register Prometheus counter '%s'", name)
	}
	return &prometheusCounter{vec: counterVec}, nil
}

func (pe *prometheusExporter) RegisterGauge(name, help string, labels ...string) (Gauge, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	if err := pe.registry.Register(gaugeVec); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			logger.L().Warnw("Gauge metric already registered", "name", name)
		}
		return nil, errors.Wrapf(err, errors.InternalError, "failed to register Prometheus gauge '%s'", name)
	}
	return &prometheusGauge{vec: gaugeVec}, nil
}

func (pe *prometheusExporter) RegisterHistogram(name, help string, buckets []float64, labels ...string) (Histogram, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if buckets == nil {
		buckets = prometheus.DefBuckets // Default buckets
	}
	histoVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	if err := pe.registry.Register(histoVec); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			logger.L().Warnw("Histogram metric already registered", "name", name)
		}
		return nil, errors.Wrapf(err, errors.InternalError, "failed to register Prometheus histogram '%s'", name)
	}
	return &prometheusHistogram{vec: histoVec}, nil
}

func (pe *prometheusExporter) RegisterSummary(name, help string, objectives map[float64]float64, labels ...string) (Summary, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001} // Default objectives
	}
	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)
	if err := pe.registry.Register(summaryVec); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			logger.L().Warnw("Summary metric already registered", "name", name)
		}
		return nil, errors.Wrapf(err, errors.InternalError, "failed to register Prometheus summary '%s'", name)
	}
	return &prometheusSummary{vec: summaryVec}, nil
}

func (pe *prometheusExporter) ExposeHandler() http.Handler {
	return promhttp.HandlerFor(pe.registry, promhttp.HandlerOpts{})
}

// No-op implementations for error cases or when labels are not used correctly.
type noopCounter struct{}

func (c *noopCounter) Inc()                   {}
func (c *noopCounter) Add(float64)            {}
func (c *noopCounter) With(...string) Counter { return c }

type noopGauge struct{}

func (g *noopGauge) Set(float64)          {}
func (g *noopGauge) Inc()                 {}
func (g *noopGauge) Dec()                 {}
func (g *noopGauge) Add(float64)          {}
func (g *noopGauge) Sub(float64)          {}
func (g *noopGauge) With(...string) Gauge { return g }

type noopHistogram struct{}

func (h *noopHistogram) Observe(float64)          {}
func (h *noopHistogram) With(...string) Histogram { return h }

type noopSummary struct{}

func (s *noopSummary) Observe(float64)        {}
func (s *noopSummary) With(...string) Summary { return s }

func init() {
	// This is a common practice but can lead to "duplicate metrics collector registration attempted"
	// if multiple packages try to register default collectors to prometheus.DefaultRegisterer.
	// It's better to use a custom registry as done in NewPrometheusExporter.
	// prometheus.MustRegister(prometheus.NewGoCollector())
	// prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	fmt.Println("Prometheus adapter initialized. Use a custom registry for robust metric registration.")
}
