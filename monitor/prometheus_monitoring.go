package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"go.uber.org/zap"
)

// metricKey struct to uniquely identify metrics with a name and labels
type metricKey struct {
	name string
	//labels string
}

func (pm *PrometheusMonitoring) getMetricKey(name string, labels []string) metricKey {
	return metricKey{name: name}
	//return metricKey{name: name, labels: fmt.Sprintf("%v", labels)}
}

// PrometheusMonitoring struct containing mutex, maps for different types of metrics, and a logger
type PrometheusMonitoring struct {
	mu         sync.RWMutex
	counters   map[metricKey]*CounterVecMetric
	gauges     map[metricKey]*GaugeVecMetric
	histograms map[metricKey]*HistogramVecMetrics
	logger     logger.LoggerInterface
}

// NewPrometheusMonitoring initializes and returns a new PrometheusMonitoring instance
func NewPrometheusMonitoring() *PrometheusMonitoring {
	return &PrometheusMonitoring{
		counters:   make(map[metricKey]*CounterVecMetric),
		gauges:     make(map[metricKey]*GaugeVecMetric),
		histograms: make(map[metricKey]*HistogramVecMetrics),
		logger:     logger.GetLogger(),
	}
}

// registerCounter registers and returns a new counter metric if it doesn't already exist
func (pm *PrometheusMonitoring) registerCounter(ctx context.Context, name string, help string, labels []string) *CounterVecMetric {
	key := pm.getMetricKey(name, labels)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if counter, exists := pm.counters[key]; exists {
		return counter
	}
	newCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(newCounter)
	counterWithIncrement := &CounterVecMetric{
		CounterVec: newCounter,
		logger:     pm.logger,
	}
	pm.counters[key] = counterWithIncrement
	pm.logger.Info(ctx, "Registered new counter", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return counterWithIncrement
}

// registerGauge registers and returns a new gauge metric if it doesn't already exist
func (pm *PrometheusMonitoring) registerGauge(ctx context.Context, name string, help string, labels []string) GaugeMetric {
	key := pm.getMetricKey(name, labels)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if gauge, exists := pm.gauges[key]; exists {
		return gauge
	}
	newGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(newGauge)
	gaugeWithSet := &GaugeVecMetric{
		GaugeVec: newGauge,
		logger:   pm.logger,
	}
	pm.gauges[key] = gaugeWithSet
	pm.logger.Info(ctx, "Registered new gauge", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return gaugeWithSet
}

// registerHistogram registers and returns a new histogram metric if it doesn't already exist
func (pm *PrometheusMonitoring) registerHistogram(ctx context.Context, name string, help string, buckets []float64, labels []string) *HistogramVecMetrics {
	key := pm.getMetricKey(name, labels)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if histogram, exists := pm.histograms[key]; exists {
		return histogram
	}
	newHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)
	prometheus.MustRegister(newHistogram)
	histogramWithObserve := &HistogramVecMetrics{
		HistogramVec: newHistogram,
		logger:       pm.logger,
	}
	pm.histograms[key] = histogramWithObserve
	pm.logger.Info(ctx, "Registered new histogram", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return histogramWithObserve
}

// RegisterCounter is a public method to register a counter
func (pm *PrometheusMonitoring) RegisterCounter(ctx context.Context, name string, help string, labels Labels) CounterMetric {
	return pm.registerCounter(ctx, name, help, pm.addCommonLabels(ctx, labels))
}

// RegisterGauge is a public method to register a gauge
func (pm *PrometheusMonitoring) RegisterGauge(ctx context.Context, name string, help string, labels Labels) GaugeMetric {
	return pm.registerGauge(ctx, name, help, pm.addCommonLabels(ctx, labels))
}

// RegisterHistogram is a public method to register a histogram
func (pm *PrometheusMonitoring) RegisterHistogram(ctx context.Context, name string, help string, buckets []float64, labels Labels) HistogramMetrics {
	return pm.registerHistogram(ctx, name, help, buckets, pm.addCommonLabels(ctx, labels))
}

func (pm *PrometheusMonitoring) addCommonLabels(ctx context.Context, labels Labels) Labels {
	labels = append(Labels{}, labels...)
	if slice, ok := ctx.Value(constraint.COMMON_LABELS_CTX).(Labels); ok {
		labels = append(labels, slice...)
	}
	return labels
}

// ExposeMetrics sets up an HTTP server to expose the /metrics endpoint for Prometheus scraping
func (pm *PrometheusMonitoring) ExposeMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	pm.logger.Info(nil, "Exposing metrics", zap.String("address", addr))
	http.ListenAndServe(addr, nil)
}
