package indicator

import (
	"context"
	"errors"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// RSI represents the state of the RSI indicator.
type RSI struct {
	Period      int
	Gains       []float64
	Losses      []float64
	AvgGain     float64
	AvgLoss     float64
	History     []model.DataPoint
	Initialized bool
}

// NewRSI initializes a new RSI instance.
func NewRSI(period int) *RSI {
	return &RSI{
		Period:      period,
		Initialized: false,
		History:     []model.DataPoint{},
		Gains:       []float64{},
		Losses:      []float64{},
	}
}

// AddDataPoint adds a new data point and updates the RSI calculation.
func (r *RSI) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	if len(r.History) > 0 && data.Time <= r.History[len(r.History)-1].Time {
		return errors.New("data point is not in chronological order")
	}

	r.History = append(r.History, data)
	if len(r.History) > r.Period {
		r.History = r.History[1:]
	} else if len(r.History) == r.Period {
		r.Initialized = true
	} else {
		return nil
	}

	r.calculateRSI()
	return nil
}

func (r *RSI) calculateRSI() {
	if len(r.Gains) == 0 {
		for i := 1; i < r.Period; i++ {
			change := r.History[i].Close - r.History[i-1].Close
			if change > 0 {
				r.Gains = append(r.Gains, change)
				r.Losses = append(r.Losses, 0)
			} else {
				r.Gains = append(r.Gains, 0)
				r.Losses = append(r.Losses, -change)
			}
		}
		r.AvgGain = sum(r.Gains) / float64(r.Period)
		r.AvgLoss = sum(r.Losses) / float64(r.Period)
	} else {
		change := r.History[len(r.History)-1].Close - r.History[len(r.History)-2].Close
		gain := 0.0
		loss := 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		r.AvgGain = (r.AvgGain*float64(r.Period-1) + gain) / float64(r.Period)
		r.AvgLoss = (r.AvgLoss*float64(r.Period-1) + loss) / float64(r.Period)
	}
}

func sum(arr []float64) float64 {
	total := 0.0
	for _, v := range arr {
		total += v
	}
	return total
}

// CalculateRSI returns the current RSI value.
func (r *RSI) CalculateRSI() float64 {
	if r.AvgLoss == 0 {
		return 100
	}
	rs := r.AvgGain / r.AvgLoss
	return 100 - (100 / (1 + rs))
}
