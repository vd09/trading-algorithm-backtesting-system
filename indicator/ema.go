package indicator

import (
	"context"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// EMA represents the state of the EMA indicator for a single period.
type EMA struct {
	Period           int
	Multiplier       float64
	Value            float64
	Initialized      bool
	InitPricesForSMA []float64
}

// NewEMA initializes a new EMA instance for a single period.
func NewEMA(period int) *EMA {
	return &EMA{
		Period:           period,
		Multiplier:       2.0 / (float64(period) + 1),
		InitPricesForSMA: []float64{},
		Initialized:      false,
	}
}

// AddDataPoint updates the EMA with a new data point and recalculates the EMA.
func (e *EMA) AddDataPoint(ctx context.Context, data model.DataPoint) {
	if !e.Initialized {
		e.InitPricesForSMA = append(e.InitPricesForSMA, data.Close)
		if len(e.InitPricesForSMA) == e.Period {
			e.Value = simpleMovingAverage(e.InitPricesForSMA)
			e.Initialized = true
		}
	} else {
		e.Value = (data.Close-e.Value)*e.Multiplier + e.Value
	}
}

// Helper function to calculate the simple moving average for initial EMA value.
func simpleMovingAverage(data []float64) float64 {
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}
