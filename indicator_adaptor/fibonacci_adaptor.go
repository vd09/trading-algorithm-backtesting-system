package indicator_adaptor

import (
	"context"
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
	FIBONACCI_LEVEL_LABEL = "fibonacci_level_name"
)

type FibonacciMetrics struct {
	SignalCounter monitor.CounterMetric
	LevelCounter  monitor.CounterMetric
	monitor       monitor.Monitoring
}

// FibonacciAdapter manages a Fibonacci indicator and maintains historical values.
type FibonacciAdapter struct {
	Fibonacci              *indicator.Fibonacci
	MaxTotalHistoricalData int
	HistoricalValues       []float64
	CurrentFibonacciLevels map[indicator.FibonacciLevel]float64
	CurrentData            model.DataPoint
	logger                 logger.LoggerInterface
	metrics                *FibonacciMetrics
}

// NewFibonacciAdapter initializes and returns a new FibonacciAdapter instance.
func NewFibonacciAdapter(ctx context.Context, size int, monitor monitor.Monitoring) *FibonacciAdapter {
	adapter := &FibonacciAdapter{
		Fibonacci:              indicator.NewFibonacci(size),
		HistoricalValues:       make([]float64, 0, size),
		MaxTotalHistoricalData: size,
		logger:                 logger.GetLogger(),
	}
	adapter.registerMetrics(ctx, monitor)
	return adapter
}

func (fa *FibonacciAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = fa.getUpdateContext(ctx)
	fa.metrics = &FibonacciMetrics{
		monitor:       m,
		SignalCounter: m.RegisterCounter(ctx, "fibonacci_signals_generated", "Total number of Fibonacci signals generated", monitor.Labels{constraint.SIGNAL_TYPE_LABEL}),
		LevelCounter:  m.RegisterCounter(ctx, "fibonacci_level_set", "Fibonacci level data set", monitor.Labels{FIBONACCI_LEVEL_LABEL}),
	}
}

func (fa *FibonacciAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewFibonacciAdapter(ctx, fa.MaxTotalHistoricalData, fa.metrics.monitor)
}

// AddDataPoint adds a new data point and updates the Fibonacci levels.
func (fa *FibonacciAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = fa.getUpdateContext(ctx)
	fa.logger.Debug(ctx, "Adding data point to FibonacciAdapter", zap.Int64("timestamp", data.Time))

	if err := fa.Fibonacci.AddDataPoint(ctx, data); err != nil {
		fa.logger.Error(ctx, "Failed to add data point to Fibonacci", zap.Error(err))
		return err
	}

	fa.CurrentData = data
	fa.CurrentFibonacciLevels = fa.Fibonacci.GetFibonacciLevels()
	fa.HistoricalValues = append(fa.HistoricalValues, data.Close)
	if len(fa.HistoricalValues) > fa.MaxTotalHistoricalData {
		fa.HistoricalValues = fa.HistoricalValues[1:]
	}
	fa.updateLevelCounters(ctx)
	return nil
}

// updateLevelCounters updates the counter for each Fibonacci level.
func (fa *FibonacciAdapter) updateLevelCounters(ctx context.Context) {
	for level, value := range fa.CurrentFibonacciLevels {
		fa.metrics.LevelCounter.SetValue(ctx, value, monitor.NewTagsKV(FIBONACCI_LEVEL_LABEL, level))
	}
}

// Name returns the name of the Fibonacci adapter.
func (fa *FibonacciAdapter) Name() string {
	return "Fibonacci"
}

// GetSignal generates a trading signal based on the Fibonacci retracement levels.
func (fa *FibonacciAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = fa.getUpdateContext(ctx)
	defer func() {
		fa.metrics.SignalCounter.IncrementCounter(ctx, monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result)))
	}()

	if !fa.Fibonacci.IsInitialized {
		fa.logger.Debug(ctx, "Fibonacci not initialized")
		return model.Wait
	}

	currentData := fa.CurrentData
	currentLevels := fa.CurrentFibonacciLevels
	if fa.isOutsideLevels(currentData.Close, currentLevels[indicator.TwentyThree], currentLevels[indicator.SeventySix]) {
		fa.logger.Debug(ctx, "Price outside key Fibonacci levels", zap.Float64("currentPrice", currentData.Close))
		return model.Wait
	}

	result = fa.evaluateHistoricalPrices(currentLevels)
	fa.logger.Info(ctx, "Generated signal", zap.String("signal", string(result)), zap.Float64("currentPrice", currentData.Close))
	return result
}

// isOutsideLevels checks if the current price is outside the key Fibonacci levels.
func (fa *FibonacciAdapter) isOutsideLevels(currentPrice, lowerLevel, upperLevel float64) bool {
	return currentPrice <= lowerLevel || currentPrice >= upperLevel
}

// evaluateHistoricalPrices checks historical prices to determine the appropriate trading signal.
func (fa *FibonacciAdapter) evaluateHistoricalPrices(currentLevels map[indicator.FibonacciLevel]float64) model.StockAction {
	for i := len(fa.HistoricalValues) - 2; i >= 0; i-- {
		prevPrice := fa.HistoricalValues[i]
		if fa.isWithinRange(prevPrice, currentLevels[indicator.TwentyThree], currentLevels[indicator.Zero], currentLevels[indicator.ThirtyEight]-currentLevels[indicator.TwentyThree]) {
			return model.Sell
		}
		if fa.isWithinRange(prevPrice, currentLevels[indicator.Hundred], currentLevels[indicator.SeventySix], currentLevels[indicator.Hundred]-currentLevels[indicator.SeventySix]) {
			return model.Buy
		}
		if prevPrice > currentLevels[indicator.TwentyThree] && prevPrice < currentLevels[indicator.SeventySix] {
			break
		}
	}
	return model.Wait
}

// isWithinRange determines if a value is within a specified range with a buffer.
func (fa *FibonacciAdapter) isWithinRange(value, lowerReference, upperReference, rangeReference float64) bool {
	buffer := utils.Min(math.Abs(upperReference-lowerReference), math.Abs(rangeReference)) * 0.25
	return value >= (lowerReference-buffer) && value <= (upperReference+buffer)
}

// getUpdateContext retrieves and updates the context with common labels and the adapter name
func (fa *FibonacciAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, fa.Name())
}
