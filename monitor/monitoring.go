package monitor

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_$GOPACKAGE/$GOFILE -package=mock_$GOPACKAGE

import (
	"context"
)

type Labels []string

type GaugeMetric interface {
	SetGauge(ctx context.Context, value float64, tags Tags)
}

type HistogramMetrics interface {
	ObserveHistogram(ctx context.Context, value float64, tags Tags)
}

type CounterMetric interface {
	IncrementCounter(ctx context.Context, tags Tags)
	SetValue(ctx context.Context, value float64, tags Tags)
}

// Monitoring is the interface for monitoring operations
type Monitoring interface {
	RegisterCounter(ctx context.Context, name string, help string, labels Labels) CounterMetric
	RegisterGauge(ctx context.Context, name string, help string, labels Labels) GaugeMetric
	RegisterHistogram(ctx context.Context, name string, help string, buckets []float64, labels Labels) HistogramMetrics
	ExposeMetrics(addr string)
}
