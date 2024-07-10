package monitor

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"go.uber.org/zap"
)

// CounterVecMetric embeds prometheus.CounterVec and adds IncrementCounter method
type CounterVecMetric struct {
	*prometheus.CounterVec
	logger logger.LoggerInterface
}

// IncrementCounter increments the counter with the provided tags and logs the operation
func (c *CounterVecMetric) IncrementCounter(ctx context.Context, tags Tags) {
	tags.AddTagsFromCtx(ctx)
	c.With(tags.Get()).Inc()
	c.logger.Debug(ctx, "Incremented counter", zap.Any("tags", tags.Get()))
}

func (c *CounterVecMetric) SetValue(ctx context.Context, value float64, tags Tags) {
	tags.AddTagsFromCtx(ctx)
	c.With(tags.Get()).Add(value)
	c.logger.Debug(ctx, "Set counter", zap.Any("tags", tags.Get()))
}

// GaugeMetric embeds prometheus.GaugeVec and adds SetGauge method
type GaugeVecMetric struct {
	*prometheus.GaugeVec
	logger logger.LoggerInterface
}

// SetGauge sets the gauge value and logs the operation
func (g *GaugeVecMetric) SetGauge(ctx context.Context, value float64, tags Tags) {
	tags.AddTagsFromCtx(ctx)
	g.With(tags.Get()).Set(value)
	g.logger.Debug(ctx, "Set gauge", zap.Float64("value", value), zap.Any("tags", tags.Get()))
}

// HistogramVecMetrics embeds prometheus.HistogramVec and adds ObserveHistogram method
type HistogramVecMetrics struct {
	*prometheus.HistogramVec
	logger logger.LoggerInterface
}

// ObserveHistogram records a value in the histogram and logs the operation
func (h *HistogramVecMetrics) ObserveHistogram(ctx context.Context, value float64, tags Tags) {
	tags.AddTagsFromCtx(ctx)
	h.With(tags.Get()).Observe(value)
	h.logger.Debug(ctx, "Observed histogram", zap.Float64("value", value), zap.Any("tags", tags.Get()))
}
