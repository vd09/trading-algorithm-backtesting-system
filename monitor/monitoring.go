package monitor

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Monitoring is the interface for monitoring operations
type Monitoring interface {
	RegisterCounter(ctx context.Context, name string, help string, labels []string) *prometheus.CounterVec
	RegisterGauge(ctx context.Context, name string, help string, labels []string) *prometheus.GaugeVec
	RegisterHistogram(ctx context.Context, name string, help string, buckets []float64, labels []string) *prometheus.HistogramVec
	IncrementCounter(ctx context.Context, counter *prometheus.CounterVec, tags Tags)
	SetGauge(ctx context.Context, gauge *prometheus.GaugeVec, value float64, tags Tags)
	ObserveHistogram(ctx context.Context, histogram *prometheus.HistogramVec, value float64, tags Tags)
	ExposeMetrics(addr string)
}
