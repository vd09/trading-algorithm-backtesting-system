package indicator

import (
	"errors"
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

type FibonacciLevel int

const (
	Zero FibonacciLevel = iota
	TwentyThree
	ThirtyEight
	Fifty
	SixtyOne
	SeventySix
	Hundred
)

// Fibonacci represents the state of the Fibonacci indicator.
type Fibonacci struct {
	History       []model.DataPoint
	High          float64
	Low           float64
	Levels        map[FibonacciLevel]float64
	Size          int
	IsInitialized bool
}

// NewFibonacci initializes a new Fibonacci instance with a given history size.
func NewFibonacci(size int) *Fibonacci {
	return &Fibonacci{
		History:       make([]model.DataPoint, 0, size),
		Levels:        make(map[FibonacciLevel]float64),
		Size:          size,
		IsInitialized: false,
	}
}

// AddDataPoint adds a new data point and updates the Fibonacci calculation.
func (f *Fibonacci) AddDataPoint(data model.DataPoint) error {
	if len(f.History) > 0 && data.Time <= f.History[len(f.History)-1].Time {
		return errors.New("data point is not in chronological order")
	}

	f.History = append(f.History, data)
	if len(f.History) > f.Size {
		f.History = f.History[1:]
	} else if len(f.History) == f.Size {
		f.IsInitialized = true
	} else {
		return nil
	}
	f.calculateHighLow()
	f.calculateFibonacciLevels()
	return nil
}

func (f *Fibonacci) calculateHighLow() {
	f.High = math.Inf(-1)
	f.Low = math.Inf(1)
	for _, dp := range f.History {
		f.Low = utils.Min(f.Low, dp.Low)
		f.High = utils.Max(f.High, dp.High)
	}
}

func (f *Fibonacci) calculateFibonacciLevels() {
	rangeDiff := f.High - f.Low
	f.Levels[Zero] = f.High
	f.Levels[TwentyThree] = f.High - (0.236 * rangeDiff)
	f.Levels[ThirtyEight] = f.High - (0.382 * rangeDiff)
	f.Levels[Fifty] = f.High - (0.5 * rangeDiff)
	f.Levels[SixtyOne] = f.High - (0.618 * rangeDiff)
	f.Levels[SeventySix] = f.High - (0.764 * rangeDiff)
	f.Levels[Hundred] = f.Low
}

// GetFibonacciLevels returns the current Fibonacci retracement levels.
func (f *Fibonacci) GetFibonacciLevels() map[FibonacciLevel]float64 {
	return f.Levels
}
