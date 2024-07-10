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

type MACDMetrics struct {
	DataPointsCounter monitor.CounterMetric
	SignalCounter     monitor.CounterMetric
	MACDLineCounter   monitor.CounterMetric
	SignalLineCounter monitor.CounterMetric
	monitor           monitor.Monitoring
}

// MACDAdapter handles a single MACD indicator and maintains historical values.
type MACDAdapter struct {
	MACD                   *indicator.MACD
	CurrentData            model.DataPoint
	MaxTotalHistoricalData int
	HistoricalValues       []indicator.MACDResult
	logger                 logger.LoggerInterface
	metrics                *MACDMetrics
}

// NewMACDAdapter initializes a new MACDAdapter instance.
func NewMACDAdapter(ctx context.Context, shortPeriod, longPeriod, signalPeriod, maxTotalHistoricalData int, monitor monitor.Monitoring) *MACDAdapter {
	macdAdapter := &MACDAdapter{
		MACD:                   indicator.NewMACD(shortPeriod, longPeriod, signalPeriod),
		HistoricalValues:       []indicator.MACDResult{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
		logger:                 logger.GetLogger(),
	}
	macdAdapter.registerMetrics(ctx, monitor)
	return macdAdapter
}

func (ma *MACDAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = ma.getUpdateContext(ctx)
	ma.metrics = &MACDMetrics{
		monitor:           m,
		DataPointsCounter: m.RegisterCounter(ctx, "macd_data_points_added_total", "Total number of data points added", nil),
		SignalCounter:     m.RegisterCounter(ctx, "macd_signals_generated", "Total number of signals generated", monitor.Labels{constraint.SIGNAL_TYPE_LABEL}),
		MACDLineCounter:   m.RegisterCounter(ctx, "macd_line", "MACD line values", nil),
		SignalLineCounter: m.RegisterCounter(ctx, "signal_line", "Signal line values", nil),
	}
}

func (ma *MACDAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewMACDAdapter(ctx, ma.MACD.ShortPeriod, ma.MACD.LongPeriod, ma.MACD.SignalPeriod, ma.MaxTotalHistoricalData, ma.metrics.monitor)
}

// AddDataPoint adds a new data point and updates the MACD.
func (ma *MACDAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = ma.getUpdateContext(ctx)
	ma.logger.Debug(ctx, "Adding data point to MACDAdapter", zap.Int64("timestamp", data.Time))

	ma.CurrentData = data
	if err := ma.MACD.AddDataPoint(ctx, data); err != nil {
		ma.logger.Error(ctx, "Failed to add data point to MACD", zap.Error(err))
		return err
	}
	if ma.MACD.Initialized {
		macdResult := ma.MACD.CalculateMACD()
		ma.HistoricalValues = append(ma.HistoricalValues, macdResult)
		if len(ma.HistoricalValues) > ma.MaxTotalHistoricalData {
			ma.HistoricalValues = ma.HistoricalValues[1:]
		}

		ma.metrics.MACDLineCounter.SetValue(ctx, macdResult.MACDLine, monitor.NewTagsKV(ADAPTOR_NAME_LABEL, ma.Name()))
		ma.metrics.SignalLineCounter.SetValue(ctx, macdResult.MACDSignal, monitor.NewTagsKV(ADAPTOR_NAME_LABEL, ma.Name()))
	}
	ma.metrics.DataPointsCounter.IncrementCounter(ctx, nil)
	return nil
}

// Name returns the name of the MACD adapter.
func (ma *MACDAdapter) Name() string {
	return fmt.Sprintf("MACD_%d_%d_%d", ma.MACD.ShortPeriod, ma.MACD.LongPeriod, ma.MACD.SignalPeriod)
}

// GetSignal returns a trading signal based on the MACD logic.
func (ma *MACDAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = ma.getUpdateContext(ctx)
	defer func() {
		tags := monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result))
		tags.Add(ADAPTOR_NAME_LABEL, ma.Name())
		ma.metrics.SignalCounter.IncrementCounter(ctx, tags)
	}()
	zaps := []zap.Field{zap.String("adapter", ma.Name()), zap.Any("time", ma.CurrentData.Time)}

	if len(ma.HistoricalValues) < 2 {
		ma.logger.Debug(ctx, "Not enough historical data for signal generation", zaps...)
		return model.Wait
	}
	if !ma.MACD.Initialized {
		ma.logger.Debug(ctx, "MACD not initialized", zaps...)
		return model.Wait
	}

	if ma.checkForBuySignal() {
		ma.logger.Info(ctx, "Buy signal detected", zaps...)
		return model.Buy
	} else if ma.checkForSellSignal() {
		ma.logger.Info(ctx, "Sell signal detected", zaps...)
		return model.Sell
	}

	ma.logger.Debug(ctx, "No trading signal detected", zaps...)
	return model.Wait
}

func (ma *MACDAdapter) checkForBuySignal() bool {
	MACDLines := make([]float64, len(ma.HistoricalValues))
	MACDSignals := make([]float64, len(ma.HistoricalValues))
	for i, hv := range ma.HistoricalValues {
		MACDLines[i] = hv.MACDLine
		MACDSignals[i] = hv.MACDSignal
	}

	if ma.HistoricalValues[0].MACDLine < ma.HistoricalValues[0].MACDSignal {
		if utils.IsLineIntersect(MACDSignals, MACDLines) {
			return true
		}
	}
	return false
}

func (ma *MACDAdapter) checkForSellSignal() bool {
	MACDLines := make([]float64, len(ma.HistoricalValues))
	MACDSignals := make([]float64, len(ma.HistoricalValues))
	for i, hv := range ma.HistoricalValues {
		MACDLines[i] = hv.MACDLine
		MACDSignals[i] = hv.MACDSignal
	}

	if ma.HistoricalValues[0].MACDLine > ma.HistoricalValues[0].MACDSignal {
		if utils.IsLineIntersect(MACDSignals, MACDLines) {
			return true
		}
	}
	return false
}

// Function to retrieve and update the slice from context
func (ma *MACDAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, ma.Name())
}
