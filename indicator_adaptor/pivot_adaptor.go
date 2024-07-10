package indicator_adaptor

import (
	"context"
	"fmt"
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
	"go.uber.org/zap"
)

const (
	PIVOT_LEVEL_LABEL = "pivot_level_type"
)

type PivotPointMetrics struct {
	SignalCounter monitor.CounterMetric
	LevelGauge    monitor.GaugeMetric
	monitor       monitor.Monitoring
}

type PivotPointAdapter struct {
	PivotPoint             *indicator.PivotPoint
	MaxTotalHistoricalData int
	HistoricalValues       []model.DataPoint
	CurrentData            model.DataPoint
	Threshold              int
	logger                 logger.LoggerInterface
	metrics                *PivotPointMetrics
}

func NewPivotPointAdapter(ctx context.Context, maxTotalHistoricalData, threshold int, monitor monitor.Monitoring) *PivotPointAdapter {
	adapter := &PivotPointAdapter{
		PivotPoint:             indicator.NewPivotPoint(),
		HistoricalValues:       []model.DataPoint{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
		Threshold:              threshold,
		logger:                 logger.GetLogger(),
	}
	adapter.registerMetrics(ctx, monitor)
	return adapter
}

func (ppa *PivotPointAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = ppa.getUpdateContext(ctx)
	ppa.metrics = &PivotPointMetrics{
		monitor:       m,
		SignalCounter: m.RegisterCounter(ctx, "pivot_point_signals_generated", "Total number of signals generated", []string{constraint.SIGNAL_TYPE_LABEL, ADAPTOR_NAME_LABEL}),
		LevelGauge:    m.RegisterGauge(ctx, "pivot_point_levels", "Current levels of Pivot, Resistance, and Support", []string{PIVOT_LEVEL_LABEL, ADAPTOR_NAME_LABEL}),
	}
}

func (ppa *PivotPointAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewPivotPointAdapter(ctx, ppa.MaxTotalHistoricalData, ppa.Threshold, ppa.metrics.monitor)
}

func (ppa *PivotPointAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = ppa.getUpdateContext(ctx)
	ppa.logger.Debug(ctx, "Adding data point to PivotPointAdapter", zap.Int64("timestamp", data.Time))

	if err := ppa.PivotPoint.AddDataPoint(ctx, data); err != nil {
		return err
	}
	ppa.CurrentData = data
	ppa.updateHistoricalValues(ctx)
	return nil
}

func (ppa *PivotPointAdapter) updateHistoricalValues(ctx context.Context) {
	levels := ppa.PivotPoint.GetPivotLevels()
	ppa.HistoricalValues = append(ppa.HistoricalValues, ppa.CurrentData)
	if len(ppa.HistoricalValues) > ppa.MaxTotalHistoricalData {
		ppa.HistoricalValues = ppa.HistoricalValues[1:]
	}

	// Update Level metrics
	ppa.updateLevelMetrics(ctx, levels)
}

func (ppa *PivotPointAdapter) updateLevelMetrics(ctx context.Context, levels indicator.PivotLevels) {
	tags := monitor.NewTagsKV(ADAPTOR_NAME_LABEL, ppa.Name())
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Pivot, tags.With(PIVOT_LEVEL_LABEL, "Pivot"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Resistance1, tags.With(PIVOT_LEVEL_LABEL, "Resistance1"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Resistance2, tags.With(PIVOT_LEVEL_LABEL, "Resistance2"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Resistance3, tags.With(PIVOT_LEVEL_LABEL, "Resistance3"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Support1, tags.With(PIVOT_LEVEL_LABEL, "Support1"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Support2, tags.With(PIVOT_LEVEL_LABEL, "Support2"))
	ppa.metrics.LevelGauge.SetGauge(ctx, levels.Support3, tags.With(PIVOT_LEVEL_LABEL, "Support3"))
}

func (ppa *PivotPointAdapter) Name() string {
	return fmt.Sprintf("PivotPoint_%d_%d", ppa.MaxTotalHistoricalData, ppa.Threshold)
}

func (ppa *PivotPointAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = ppa.getUpdateContext(ctx)
	defer func() {
		tags := monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result))
		tags.Add(ADAPTOR_NAME_LABEL, ppa.Name())
		ppa.metrics.SignalCounter.IncrementCounter(ctx, tags)
	}()

	zaps := []zap.Field{zap.String("adapter", ppa.Name()), zap.Any("time", ppa.CurrentData.Time)}
	if len(ppa.HistoricalValues) == 0 {
		ppa.logger.Debug(ctx, "No historical data available", zaps...)
		return model.Wait
	}
	if !ppa.PivotPoint.Initialized {
		ppa.logger.Debug(ctx, "PivotPoint not initialized", zaps...)
		return model.Wait
	}

	lastData := ppa.CurrentData
	lastPivot := ppa.PivotPoint.Levels

	recentTests := ppa.calculateRecentTests()

	if ppa.evaluateBuySignal(lastData, lastPivot, recentTests) {
		ppa.logger.Info(ctx, "Buy signal detected", zaps...)
		return model.Buy
	}

	if ppa.evaluateSellSignal(lastData, lastPivot, recentTests) {
		ppa.logger.Info(ctx, "Sell signal detected", zaps...)
		return model.Sell
	}

	ppa.logger.Debug(ctx, "No trading signal detected", zaps...)
	return model.Wait
}

func (ppa *PivotPointAdapter) calculateRecentTests() map[string]int {
	recentTests := map[string]int{
		"Resistance1": 0,
		"Resistance2": 0,
		"Resistance3": 0,
		"Support1":    0,
		"Support2":    0,
		"Support3":    0,
	}

	pivot := ppa.PivotPoint.Levels
	for i := len(ppa.HistoricalValues) - 1; i >= 0 && i >= len(ppa.HistoricalValues)-5; i-- {
		dataPoint := ppa.HistoricalValues[i]
		if isWithinRange(dataPoint.Close, pivot.Pivot, pivot.Resistance1, pivot.Resistance2) {
			recentTests["Resistance1"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Resistance1, pivot.Resistance2, pivot.Resistance3) {
			recentTests["Resistance2"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Resistance2, pivot.Resistance3, 0) {
			recentTests["Resistance3"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Support2, pivot.Support1, pivot.Pivot) {
			recentTests["Support1"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Support3, pivot.Support2, pivot.Support1) {
			recentTests["Support2"]++
		}
		if isWithinRange(dataPoint.Close, 0, pivot.Support3, pivot.Support2) {
			recentTests["Support3"]++
		}
	}

	return recentTests
}

func (ppa *PivotPointAdapter) evaluateBuySignal(lastData model.DataPoint, lastPivot indicator.PivotLevels, recentTests map[string]int) bool {
	if recentTests["Resistance3"] >= ppa.Threshold && lastData.High > lastPivot.Resistance3 && lastData.Low < lastPivot.Resistance3 && lastData.Close > lastPivot.Resistance3 {
		return true
	}
	if recentTests["Resistance2"] >= ppa.Threshold && lastData.High > lastPivot.Resistance2 && lastData.Low < lastPivot.Resistance2 && lastData.Close > lastPivot.Resistance2 {
		return true
	}
	if recentTests["Resistance1"] >= ppa.Threshold && lastData.High > lastPivot.Resistance1 && lastData.Low < lastPivot.Resistance1 && lastData.Close > lastPivot.Resistance1 {
		return true
	}
	return false
}

func (ppa *PivotPointAdapter) evaluateSellSignal(lastData model.DataPoint, lastPivot indicator.PivotLevels, recentTests map[string]int) bool {
	if recentTests["Support3"] >= ppa.Threshold && lastData.Low < lastPivot.Support3 && lastData.High > lastPivot.Support3 && lastData.Close < lastPivot.Support3 {
		return true
	}
	if recentTests["Support2"] >= ppa.Threshold && lastData.Low < lastPivot.Support2 && lastData.High > lastPivot.Support2 && lastData.Close < lastPivot.Support2 {
		return true
	}
	if recentTests["Support1"] >= ppa.Threshold && lastData.Low < lastPivot.Support1 && lastData.High > lastPivot.Support1 && lastData.Close < lastPivot.Support1 {
		return true
	}
	return false
}

// Function to retrieve and update the slice from context
func (ppa *PivotPointAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, ppa.Name())
}

func isWithinRange(value, lowerReference, middleReference, upperReference float64) bool {
	diff := utils.Min(math.Abs(middleReference-lowerReference), math.Abs(upperReference-middleReference)) * 0.25
	lowerBound := middleReference - diff
	upperBound := middleReference + diff
	return value >= lowerBound && value <= upperBound
}
