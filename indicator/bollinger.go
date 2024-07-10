package indicator

import (
	"context"
	"errors"
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// BollingerBandsValues represents the calculated values of Bollinger Bands.
type BollingerBandsValues struct {
	MovingAverage float64
	UpperBand     float64
	LowerBand     float64
}

// BollingerBands represents the state of the Bollinger Bands indicator.
type BollingerBands struct {
	Period  int
	History []model.DataPoint
	Values  BollingerBandsValues
}

// NewBollingerBands initializes a new Bollinger Bands instance.
func NewBollingerBands(period int) *BollingerBands {
	return &BollingerBands{
		Period: period,
	}
}

// AddDataPoint adds a new data point and updates the Bollinger Bands calculation.
func (bb *BollingerBands) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	if len(bb.History) > 0 && data.Time <= bb.History[len(bb.History)-1].Time {
		return errors.New("data point is not in chronological order")
	}

	bb.History = append(bb.History, data)
	if len(bb.History) > bb.Period {
		bb.History = bb.History[1:]
	}

	bb.calculateBollingerBands()
	return nil
}

func (bb *BollingerBands) calculateBollingerBands() {
	if len(bb.History) < bb.Period {
		return
	}

	// Calculate the moving average
	sum := 0.0
	for i := len(bb.History) - bb.Period; i < len(bb.History); i++ {
		sum += bb.History[i].Close
	}
	bb.Values.MovingAverage = sum / float64(bb.Period)

	// Calculate the standard deviation
	sumOfSquares := 0.0
	for i := len(bb.History) - bb.Period; i < len(bb.History); i++ {
		sumOfSquares += math.Pow(bb.History[i].Close-bb.Values.MovingAverage, 2)
	}
	stdDev := math.Sqrt(sumOfSquares / float64(bb.Period))

	// Calculate the upper and lower bands
	bb.Values.UpperBand = bb.Values.MovingAverage + (2 * stdDev)
	bb.Values.LowerBand = bb.Values.MovingAverage - (2 * stdDev)
}

// GetBollingerBands returns the current Bollinger Bands levels.
func (bb *BollingerBands) GetBollingerBands() BollingerBandsValues {
	return bb.Values
}
