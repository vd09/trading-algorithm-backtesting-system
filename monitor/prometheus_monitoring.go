package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"go.uber.org/zap"
)

type metricKey struct {
	name   string
	labels string
}

type PrometheusMonitoring struct {
	mu         sync.Mutex
	counters   map[metricKey]*prometheus.CounterVec
	gauges     map[metricKey]*prometheus.GaugeVec
	histograms map[metricKey]*prometheus.HistogramVec
	logger     logger.LoggerInterface
}

func NewPrometheusMonitoring() *PrometheusMonitoring {
	return &PrometheusMonitoring{
		counters:   make(map[metricKey]*prometheus.CounterVec),
		gauges:     make(map[metricKey]*prometheus.GaugeVec),
		histograms: make(map[metricKey]*prometheus.HistogramVec),
		logger:     logger.GetLogger(),
	}
}

func (pm *PrometheusMonitoring) getMetricKey(name string, labels []string) metricKey {
	return metricKey{name: name, labels: fmt.Sprintf("%v", labels)}
}

func (pm *PrometheusMonitoring) registerCounter(ctx context.Context, name string, help string, labels []string) *prometheus.CounterVec {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	key := pm.getMetricKey(name, labels)
	if counter, exists := pm.counters[key]; exists {
		return counter
	}
	newCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(newCounter)
	pm.counters[key] = newCounter
	pm.logger.Info(ctx, "Registered new counter", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return newCounter
}

func (pm *PrometheusMonitoring) registerGauge(ctx context.Context, name string, help string, labels []string) *prometheus.GaugeVec {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	key := pm.getMetricKey(name, labels)
	if gauge, exists := pm.gauges[key]; exists {
		return gauge
	}
	newGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(newGauge)
	pm.gauges[key] = newGauge
	pm.logger.Info(ctx, "Registered new gauge", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return newGauge
}

func (pm *PrometheusMonitoring) registerHistogram(ctx context.Context, name string, help string, buckets []float64, labels []string) *prometheus.HistogramVec {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	key := pm.getMetricKey(name, labels)
	if histogram, exists := pm.histograms[key]; exists {
		return histogram
	}
	newHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)
	prometheus.MustRegister(newHistogram)
	pm.histograms[key] = newHistogram
	pm.logger.Info(ctx, "Registered new histogram", zap.String("name", name), zap.String("labels", fmt.Sprintf("%v", labels)))
	return newHistogram
}

func (pm *PrometheusMonitoring) RegisterCounter(ctx context.Context, name string, help string, labels []string) *prometheus.CounterVec {
	return pm.registerCounter(ctx, name, help, labels)
}

func (pm *PrometheusMonitoring) RegisterGauge(ctx context.Context, name string, help string, labels []string) *prometheus.GaugeVec {
	return pm.registerGauge(ctx, name, help, labels)
}

func (pm *PrometheusMonitoring) RegisterHistogram(ctx context.Context, name string, help string, buckets []float64, labels []string) *prometheus.HistogramVec {
	return pm.registerHistogram(ctx, name, help, buckets, labels)
}

func (pm *PrometheusMonitoring) IncrementCounter(ctx context.Context, counter *prometheus.CounterVec, tags Tags) {
	counter.With(tags.Get()).Inc()
	pm.logger.Info(ctx, "Incremented counter", zap.Any("tags", tags.Get()))
}

func (pm *PrometheusMonitoring) SetGauge(ctx context.Context, gauge *prometheus.GaugeVec, value float64, tags Tags) {
	gauge.With(tags.Get()).Set(value)
	pm.logger.Info(ctx, "Set gauge", zap.Float64("value", value), zap.Any("tags", tags.Get()))
}

func (pm *PrometheusMonitoring) ObserveHistogram(ctx context.Context, histogram *prometheus.HistogramVec, value float64, tags Tags) {
	histogram.With(tags.Get()).Observe(value)
	pm.logger.Info(ctx, "Observed histogram", zap.Float64("value", value), zap.Any("tags", tags.Get()))
}

func (pm *PrometheusMonitoring) ExposeMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	pm.logger.Info(nil, "Exposing metrics", zap.String("address", addr))
	http.ListenAndServe(addr, nil)
}
