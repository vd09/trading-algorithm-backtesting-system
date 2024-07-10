package indicator_adaptor

import (
	"context"
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
	"go.uber.org/zap"
)

const (
	SUPERTREND_LEVEL_LABEL = "supertrend_level_type"
)

type SuperTrendMetrics struct {
	SignalCounter monitor.CounterMetric
	TrendCounter  monitor.CounterMetric
	TrendLine     monitor.CounterMetric
	monitor       monitor.Monitoring
}

type InitializeStatus int

const (
	NOT_INITIALIZED InitializeStatus = iota
	INITIALIZED
	START_SIGNALING
)

// SuperTrendAdapter handles a single SuperTrend indicator and maintains historical values.
type SuperTrendAdapter struct {
	SuperTrend    *indicator.SuperTrend
	PreviousTrend bool
	CurrentTrend  bool
	initialized   InitializeStatus
	logger        logger.LoggerInterface
	metrics       *SuperTrendMetrics
}

// NewSuperTrendAdapter initializes a new SuperTrendAdapter instance.
func NewSuperTrendAdapter(ctx context.Context, period int, multiplier float64, monitor monitor.Monitoring) *SuperTrendAdapter {
	adapter := &SuperTrendAdapter{
		SuperTrend:  indicator.NewSuperTrend(period, multiplier),
		initialized: NOT_INITIALIZED,
		logger:      logger.GetLogger(),
	}
	adapter.registerMetrics(ctx, monitor)
	return adapter
}

func (sta *SuperTrendAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = sta.getUpdateContext(ctx)
	sta.metrics = &SuperTrendMetrics{
		monitor:       m,
		SignalCounter: m.RegisterCounter(ctx, "super_trend_signals_generated", "Total number of SuperTrend signals generated", monitor.Labels{constraint.SIGNAL_TYPE_LABEL}),
		TrendCounter:  m.RegisterCounter(ctx, "super_trend_direction", "trend direction in SuperTrend", nil),
		TrendLine:     m.RegisterCounter(ctx, "super_trend_line", "SuperTrend Line", nil),
	}
}

// Clone creates a new instance of SuperTrendAdapter with the same configuration.
func (sta *SuperTrendAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewSuperTrendAdapter(ctx, sta.SuperTrend.Period, sta.SuperTrend.Multiplier, sta.metrics.monitor)
}

// AddDataPoint adds a new data point and updates the SuperTrend.
func (sta *SuperTrendAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = sta.getUpdateContext(ctx)
	sta.logger.Debug(ctx, "Adding data point to SuperTrendAdapter", zap.Int64("timestamp", data.Time))

	if err := sta.SuperTrend.AddDataPoint(ctx, data); err != nil {
		sta.logger.Error(ctx, "Failed to add data point to SuperTrend", zap.Error(err))
		return err
	}

	if sta.SuperTrend.Initialized {
		SuperTrendLine, isUpTrend := sta.SuperTrend.CalculateSuperTrend()
		if sta.initialized == NOT_INITIALIZED {
			sta.initialized = INITIALIZED
		} else {
			sta.PreviousTrend = sta.CurrentTrend
			sta.initialized = START_SIGNALING
		}
		sta.CurrentTrend = isUpTrend
		sta.metrics.TrendCounter.SetValue(ctx, utils.B2F(sta.CurrentTrend), nil)
		sta.metrics.TrendLine.SetValue(ctx, SuperTrendLine, nil)
	}
	return nil
}

// Name returns the name of the SuperTrend adapter.
func (sta *SuperTrendAdapter) Name() string {
	return fmt.Sprintf("SuperTrend_%d_%.2f", sta.SuperTrend.Period, sta.SuperTrend.Multiplier)
}

// GetSignal returns a trading signal based on the SuperTrend logic.
// GetSignal returns a trading signal based on the SuperTrend logic.
func (sta *SuperTrendAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = sta.getUpdateContext(ctx)
	defer func() {
		sta.metrics.SignalCounter.IncrementCounter(ctx, monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result)))
	}()

	if sta.initialized != START_SIGNALING {
		sta.logger.Debug(ctx, "SuperTrend not ready for signaling", zap.String("status", "NOT_START_SIGNALING"))
		return model.Wait
	}
	if !sta.SuperTrend.Initialized {
		sta.logger.Debug(ctx, "SuperTrend not initialized")
		return model.Wait
	}

	if sta.PreviousTrend == false && sta.CurrentTrend == true {
		sta.logger.Info(ctx, "Buy signal detected", zap.Bool("previousTrend", sta.PreviousTrend), zap.Bool("currentTrend", sta.CurrentTrend))
		return model.Buy
	} else if sta.PreviousTrend == true && sta.CurrentTrend == false {
		sta.logger.Info(ctx, "Sell signal detected", zap.Bool("previousTrend", sta.PreviousTrend), zap.Bool("currentTrend", sta.CurrentTrend))
		return model.Sell
	}

	sta.logger.Debug(ctx, "No trading signal detected", zap.Bool("previousTrend", sta.PreviousTrend), zap.Bool("currentTrend", sta.CurrentTrend))
	return model.Wait
}

// Function to retrieve and update the context with common labels and the adapter name
func (sta *SuperTrendAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, sta.Name())
}
