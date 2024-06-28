package indicator_adaptor

import (
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

// FibonacciAdapter manages a Fibonacci indicator and maintains historical values.
type FibonacciAdapter struct {
	Fibonacci              *indicator.Fibonacci
	MaxTotalHistoricalData int
	HistoricalValues       []float64
	CurrentFibonacciLevels map[indicator.FibonacciLevel]float64
	CurrentData            model.DataPoint
}

// NewFibonacciAdapter initializes and returns a new FibonacciAdapter instance.
func NewFibonacciAdapter(size int) *FibonacciAdapter {
	return &FibonacciAdapter{
		Fibonacci:              indicator.NewFibonacci(size),
		HistoricalValues:       make([]float64, 0, size),
		MaxTotalHistoricalData: size,
	}
}

func (fa *FibonacciAdapter) Clone() IndicatorAdaptor {
	return NewFibonacciAdapter(fa.MaxTotalHistoricalData)
}

// AddDataPoint adds a new data point and updates the Fibonacci levels.
func (fa *FibonacciAdapter) AddDataPoint(data model.DataPoint) error {
	if err := fa.Fibonacci.AddDataPoint(data); err != nil {
		return err
	}
	fa.CurrentData = data
	fa.CurrentFibonacciLevels = fa.Fibonacci.GetFibonacciLevels()
	fa.HistoricalValues = append(fa.HistoricalValues, data.Close)
	if len(fa.HistoricalValues) > fa.MaxTotalHistoricalData {
		fa.HistoricalValues = fa.HistoricalValues[1:]
	}
	return nil
}

// Name returns the name of the Fibonacci adapter.
func (fa *FibonacciAdapter) Name() string {
	return "Fibonacci"
}

// GetSignal generates a trading signal based on the Fibonacci retracement levels.
func (fa *FibonacciAdapter) GetSignal() model.StockAction {
	if !fa.Fibonacci.IsInitialized {
		return model.Wait
	}

	currentData := fa.CurrentData
	currentLevels := fa.CurrentFibonacciLevels
	if fa.isOutsideLevels(currentData.Close, currentLevels[indicator.TwentyThree], currentLevels[indicator.SeventySix]) {
		return model.Wait
	}

	return fa.evaluateHistoricalPrices(currentLevels)
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
