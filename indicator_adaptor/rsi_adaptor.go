package indicator_adaptor

import (
	"context"
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
	"go.uber.org/zap"
)

const (
	RSI_LEVEL_LABEL = "rsi_level_type"
)

type RSIMetrics struct {
	SignalCounter   monitor.CounterMetric
	RsiValueCounter monitor.CounterMetric
	LevelGauge      monitor.GaugeMetric
	monitor         monitor.Monitoring
}

type RSIAdapter struct {
	RSI                    *indicator.RSI
	MaxTotalHistoricalData int
	HistoricalValues       []float64
	OverboughtThreshold    float64
	OversoldThreshold      float64
	logger                 logger.LoggerInterface
	metrics                *RSIMetrics
}

func NewRSIAdapter(ctx context.Context, period, maxTotalHistoricalData int, overboughtThreshold, oversoldThreshold float64, monitor monitor.Monitoring) *RSIAdapter {
	adapter := &RSIAdapter{
		RSI:                    indicator.NewRSI(period),
		HistoricalValues:       []float64{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
		OverboughtThreshold:    overboughtThreshold,
		OversoldThreshold:      oversoldThreshold,
		logger:                 logger.GetLogger(),
	}
	adapter.registerMetrics(ctx, monitor)
	return adapter
}

func (ra *RSIAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = ra.getUpdateContext(ctx)
	ra.metrics = &RSIMetrics{
		monitor:         m,
		SignalCounter:   m.RegisterCounter(ctx, "rsi_signals_generated", "Total number of RSI signals generated", monitor.Labels{constraint.SIGNAL_TYPE_LABEL}),
		RsiValueCounter: m.RegisterCounter(ctx, "rsi_line_value", "Current value of RSI", nil),
		LevelGauge:      m.RegisterGauge(ctx, "rsi_levels", "Current levels of RSI", monitor.Labels{RSI_LEVEL_LABEL}),
	}
}

func (ra *RSIAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewRSIAdapter(ctx, ra.RSI.Period, ra.MaxTotalHistoricalData, ra.OverboughtThreshold, ra.OversoldThreshold, ra.metrics.monitor)
}

func (ra *RSIAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = ra.getUpdateContext(ctx)
	ra.metrics.LevelGauge.SetGauge(ctx, ra.OversoldThreshold, monitor.NewTagsKV(RSI_LEVEL_LABEL, "over_sold_threshold"))
	ra.metrics.LevelGauge.SetGauge(ctx, ra.OverboughtThreshold, monitor.NewTagsKV(RSI_LEVEL_LABEL, "over_bought_threshold"))

	ra.logger.Debug(ctx, "Adding data point to RSIAdapter", zap.Int64("timestamp", data.Time))
	if err := ra.RSI.AddDataPoint(ctx, data); err != nil {
		ra.logger.Error(ctx, "Failed to add data point to RSI", zap.Error(err))
		return err
	}

	if ra.RSI.Initialized {
		newRSI := ra.RSI.CalculateRSI()
		ra.HistoricalValues = append(ra.HistoricalValues, newRSI)
		if len(ra.HistoricalValues) > ra.MaxTotalHistoricalData {
			ra.HistoricalValues = ra.HistoricalValues[1:]
		}
		ra.metrics.RsiValueCounter.SetValue(ctx, newRSI, nil)
	}

	return nil
}

func (ra *RSIAdapter) Name() string {
	return fmt.Sprintf("RSI_P(%d)_OBT(%f)_OST(%f)_L(%d)", ra.RSI.Period, ra.OverboughtThreshold, ra.OversoldThreshold, ra.MaxTotalHistoricalData)
}

func (ra *RSIAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = ra.getUpdateContext(ctx)
	defer func() {
		ra.metrics.SignalCounter.IncrementCounter(ctx, monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result)))
	}()

	if len(ra.HistoricalValues) < 2 {
		ra.logger.Debug(ctx, "Not enough historical values for RSI signal generation", zap.Int("historicalValuesLength", len(ra.HistoricalValues)))
		return model.Wait
	}
	if !ra.RSI.Initialized {
		ra.logger.Debug(ctx, "RSI not initialized")
		return model.Wait
	}

	currentRSI := ra.HistoricalValues[len(ra.HistoricalValues)-1]
	ra.logger.Debug(ctx, "Current RSI value", zap.Float64("currentRSI", currentRSI))

	for i := len(ra.HistoricalValues) - 2; i >= 0; i-- {
		if ra.HistoricalValues[i] > ra.OverboughtThreshold {
			if currentRSI <= ra.OverboughtThreshold {
				ra.logger.Info(ctx, "Sell signal detected", zap.Float64("currentRSI", currentRSI), zap.Float64("overboughtThreshold", ra.OverboughtThreshold))
				return model.Sell
			}
			break
		} else if ra.HistoricalValues[i] < ra.OversoldThreshold {
			if currentRSI >= ra.OversoldThreshold {
				ra.logger.Info(ctx, "Buy signal detected", zap.Float64("currentRSI", currentRSI), zap.Float64("oversoldThreshold", ra.OversoldThreshold))
				return model.Buy
			}
			break
		} else if ra.HistoricalValues[i] > ra.OversoldThreshold && ra.HistoricalValues[i] < ra.OverboughtThreshold {
			break
		}
	}

	ra.logger.Debug(ctx, "No trading signal detected", zap.Float64("currentRSI", currentRSI))
	return model.Wait
}

// Function to retrieve and update the slice from context
func (ra *RSIAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, ra.Name())
}
