package indicator_adaptor

import (
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

type InitializeStatus int

const (
	NOT_INITIALIZED InitializeStatus = iota
	INITIALIZED
	START_SIGNALING
)

// SuperTrendAdapter handles a single SuperTrend indicator and maintains historical values.
type SuperTrendAdapter struct {
	SuperTrend    *indicator.SuperTrend
	PreviousTrend bool
	CurrentTrend  bool
	initialized   InitializeStatus
}

// NewSuperTrendAdapter initializes a new SuperTrendAdapter instance.
func NewSuperTrendAdapter(period int, multiplier float64) *SuperTrendAdapter {
	return &SuperTrendAdapter{
		SuperTrend:  indicator.NewSuperTrend(period, multiplier),
		initialized: NOT_INITIALIZED,
	}
}

func (sta *SuperTrendAdapter) Clone() IndicatorAdaptor {
	return NewSuperTrendAdapter(sta.SuperTrend.Period, sta.SuperTrend.Multiplier)
}

// AddDataPoint adds a new data point and updates the SuperTrend.
func (sta *SuperTrendAdapter) AddDataPoint(data model.DataPoint) error {
	if err := sta.SuperTrend.AddDataPoint(data); err != nil {
		return err
	}
	if sta.SuperTrend.Initialized {
		_, isUpTrend := sta.SuperTrend.CalculateSuperTrend()
		if sta.initialized == NOT_INITIALIZED {
			sta.initialized = INITIALIZED
		} else {
			sta.PreviousTrend = sta.CurrentTrend
			sta.initialized = START_SIGNALING
		}
		sta.CurrentTrend = isUpTrend
	}
	return nil
}

// Name returns the name of the SuperTrend adapter.
func (sta *SuperTrendAdapter) Name() string {
	return fmt.Sprintf("SuperTrend_%d_%.2f", sta.SuperTrend.Period, sta.SuperTrend.Multiplier)
}

// GetSignal returns a trading signal based on the SuperTrend logic.
func (sta *SuperTrendAdapter) GetSignal() model.StockAction {
	if sta.initialized != START_SIGNALING {
		return model.Wait
	}
	if !sta.SuperTrend.Initialized {
		return model.Wait
	}

	if sta.PreviousTrend == false && sta.CurrentTrend == true {
		return model.Buy
	} else if sta.PreviousTrend == true && sta.CurrentTrend == false {
		return model.Sell
	}
	return model.Wait
}
