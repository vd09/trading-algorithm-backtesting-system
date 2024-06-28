package indicator

import (
	"errors"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// PivotPoint represents the state of the Pivot Point indicator.
type PivotPoint struct {
	PreviousData model.DataPoint
	Initialized  bool
	Levels       PivotLevels
}

// PivotLevels represents the pivot point and its support/resistance levels.
type PivotLevels struct {
	Pivot       float64
	Support1    float64
	Support2    float64
	Support3    float64
	Resistance1 float64
	Resistance2 float64
	Resistance3 float64
}

// NewPivotPoint initializes a new Pivot Point instance.
func NewPivotPoint() *PivotPoint {
	return &PivotPoint{Initialized: false}
}

// AddDataPoint adds a new data point and updates the Pivot Point calculation.
func (pp *PivotPoint) AddDataPoint(data model.DataPoint) error {
	if pp.PreviousData.Time > 0 && data.Time <= pp.PreviousData.Time {
		return errors.New("data point is not in chronological order")
	}

	if pp.Initialized {
		pp.calculatePivotLevels()
	} else {
		pp.Initialized = true
	}
	pp.PreviousData = data
	return nil
}

func (pp *PivotPoint) calculatePivotLevels() {
	high := pp.PreviousData.High
	low := pp.PreviousData.Low
	close := pp.PreviousData.Close

	pp.Levels = PivotLevels{}
	pp.Levels.Pivot = (high + low + close) / 3
	pp.Levels.Support1 = 2*pp.Levels.Pivot - high
	pp.Levels.Support2 = pp.Levels.Pivot - (high - low)
	pp.Levels.Support3 = low - 2*(high-pp.Levels.Pivot)
	pp.Levels.Resistance1 = 2*pp.Levels.Pivot - low
	pp.Levels.Resistance2 = pp.Levels.Pivot + (high - low)
	pp.Levels.Resistance3 = high + 2*(pp.Levels.Pivot-low)
}

// GetPivotLevels returns the current Pivot Point levels.
func (pp *PivotPoint) GetPivotLevels() PivotLevels {
	return pp.Levels
}
