package indicator

import (
	"errors"
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// SuperTrend represents the state of the Super Trend indicator.
type SuperTrend struct {
	Period         int
	Multiplier     float64
	ATR            float64
	SuperTrendLine float64
	History        []model.DataPoint
	IsUpTrend      bool
	Initialized    bool
}

// NewSuperTrend initializes a new Super Trend instance.
func NewSuperTrend(period int, multiplier float64) *SuperTrend {
	return &SuperTrend{
		Period:      period,
		Multiplier:  multiplier,
		Initialized: false,
	}
}

// AddDataPoint adds a new data point and updates the Super Trend calculation.
func (st *SuperTrend) AddDataPoint(data model.DataPoint) error {
	if len(st.History) > 0 && data.Time <= st.History[len(st.History)-1].Time {
		return errors.New("data point is not in chronological order")
	}

	// Check for negative or zero prices
	if data.Open <= 0 || data.High <= 0 || data.Low <= 0 || data.Close <= 0 {
		return errors.New("data point contains negative or zero prices")
	}

	st.History = append(st.History, data)
	if len(st.History) > st.Period {
		st.History = st.History[1:]
	} else if len(st.History) == st.Period {
		st.Initialized = true
	} else {
		return nil
	}

	st.calculateATR()
	st.calculateSuperTrend()
	return nil
}

func (st *SuperTrend) calculateATR() {
	if len(st.History) < st.Period {
		return
	}

	trSum := 0.0
	for i := 1; i < len(st.History); i++ {
		highLow := st.History[i].High - st.History[i].Low
		highClose := math.Abs(st.History[i].High - st.History[i-1].Close)
		lowClose := math.Abs(st.History[i].Low - st.History[i-1].Close)
		tr := math.Max(highLow, math.Max(highClose, lowClose))
		trSum += tr
	}
	st.ATR = trSum / float64(st.Period)
}

func (st *SuperTrend) calculateSuperTrend() {
	if len(st.History) < st.Period {
		return
	}

	midPoint := (st.History[len(st.History)-1].High + st.History[len(st.History)-1].Low) / 2
	if st.IsUpTrend {
		st.SuperTrendLine = midPoint - (st.Multiplier * st.ATR)
		if st.History[len(st.History)-1].Close < st.SuperTrendLine {
			st.IsUpTrend = false
		}
	} else {
		st.SuperTrendLine = midPoint + (st.Multiplier * st.ATR)
		if st.History[len(st.History)-1].Close > st.SuperTrendLine {
			st.IsUpTrend = true
		}
	}
}

// CalculateSuperTrend returns the current Super Trend value and the trend direction.
func (st *SuperTrend) CalculateSuperTrend() (float64, bool) {
	return st.SuperTrendLine, st.IsUpTrend
}
