package indicator_adaptor

import (
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// RSIAdapter handles a single RSI indicator and maintains historical values.
type RSIAdapter struct {
	RSI                    *indicator.RSI
	MaxTotalHistoricalData int
	HistoricalValues       []float64
	OverboughtThreshold    float64
	OversoldThreshold      float64
}

// NewRSIAdapter initializes a new RSIAdapter instance.
func NewRSIAdapter(period, maxTotalHistoricalData int, overboughtThreshold, oversoldThreshold float64) *RSIAdapter {
	return &RSIAdapter{
		RSI:                    indicator.NewRSI(period),
		HistoricalValues:       []float64{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
		OverboughtThreshold:    overboughtThreshold,
		OversoldThreshold:      oversoldThreshold,
	}
}

func (ra *RSIAdapter) Clone() IndicatorAdaptor {
	return NewRSIAdapter(ra.RSI.Period, ra.MaxTotalHistoricalData, ra.OverboughtThreshold, ra.OversoldThreshold)
}

// AddDataPoint adds a new data point and updates the RSI.
func (ra *RSIAdapter) AddDataPoint(data model.DataPoint) error {
	if err := ra.RSI.AddDataPoint(data); err != nil {
		return err
	}
	if ra.RSI.Initialized {
		ra.HistoricalValues = append(ra.HistoricalValues, ra.RSI.CalculateRSI())
		if len(ra.HistoricalValues) > ra.MaxTotalHistoricalData {
			ra.HistoricalValues = ra.HistoricalValues[1:]
		}
	}
	return nil
}

// Name returns the name of the RSI adapter.
func (ra *RSIAdapter) Name() string {
	return fmt.Sprintf("RSI_%d", ra.RSI.Period)
}

// GetSignal returns a trading signal based on the RSI logic.
func (ra *RSIAdapter) GetSignal() model.StockAction {
	if len(ra.HistoricalValues) < 2 {
		return model.Wait
	}
	if !ra.RSI.Initialized {
		return model.Wait
	}

	currentRSI := ra.HistoricalValues[len(ra.HistoricalValues)-1]
	// Find the last time RSI was out of the range (not equal to thresholds)
	for i := len(ra.HistoricalValues) - 2; i >= 0; i-- {
		if ra.HistoricalValues[i] > ra.OverboughtThreshold {
			if currentRSI <= ra.OverboughtThreshold {
				return model.Sell
			}
			break
		} else if ra.HistoricalValues[i] < ra.OversoldThreshold {
			if currentRSI >= ra.OversoldThreshold {
				return model.Buy
			}
			break
		} else if ra.HistoricalValues[i] > ra.OversoldThreshold && ra.HistoricalValues[i] < ra.OverboughtThreshold {
			break
		}
	}

	return model.Wait
}
