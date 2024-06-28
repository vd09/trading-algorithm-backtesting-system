package indicator_adaptor

import (
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

// MACDAdapter handles a single MACD indicator and maintains historical values.
type MACDAdapter struct {
	MACD                   *indicator.MACD
	MaxTotalHistoricalData int
	HistoricalValues       []indicator.MACDResult
}

// NewMACDAdapter initializes a new MACDAdapter instance.
func NewMACDAdapter(shortPeriod, longPeriod, signalPeriod, maxTotalHistoricalData int) *MACDAdapter {
	return &MACDAdapter{
		MACD:                   indicator.NewMACD(shortPeriod, longPeriod, signalPeriod),
		HistoricalValues:       []indicator.MACDResult{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
	}
}

func (ma *MACDAdapter) Clone() IndicatorAdaptor {
	return NewMACDAdapter(ma.MACD.ShortPeriod, ma.MACD.LongPeriod, ma.MACD.SignalPeriod, ma.MaxTotalHistoricalData)
}

// AddDataPoint adds a new data point and updates the MACD.
func (ma *MACDAdapter) AddDataPoint(data model.DataPoint) error {
	if err := ma.MACD.AddDataPoint(data); err != nil {
		return err
	}
	if ma.MACD.Initialized {
		macdResult := ma.MACD.CalculateMACD()
		ma.HistoricalValues = append(ma.HistoricalValues, macdResult)
		if len(ma.HistoricalValues) > ma.MaxTotalHistoricalData {
			ma.HistoricalValues = ma.HistoricalValues[1:]
		}
	}
	return nil
}

// Name returns the name of the MACD adapter.
func (ma *MACDAdapter) Name() string {
	return fmt.Sprintf("MACD_%d_%d_%d", ma.MACD.ShortPeriod, ma.MACD.LongPeriod, ma.MACD.SignalPeriod)
}

// GetSignal returns a trading signal based on the MACD logic.
func (ma *MACDAdapter) GetSignal() model.StockAction {
	if len(ma.HistoricalValues) < 2 {
		return model.Wait
	}
	if !ma.MACD.Initialized {
		return model.Wait
	}

	if ma.checkForBuySignal() {
		return model.Buy
	} else if ma.checkForSellSignal() {
		return model.Sell
	}
	return model.Wait
}

func (ma *MACDAdapter) checkForBuySignal() bool {
	MACDLines := make([]float64, len(ma.HistoricalValues))
	MACDSignals := make([]float64, len(ma.HistoricalValues))
	for i, hv := range ma.HistoricalValues {
		MACDLines[i] = hv.MACDLine
		MACDSignals[i] = hv.MACDSignal
	}

	if ma.HistoricalValues[0].MACDLine < ma.HistoricalValues[0].MACDSignal {
		if utils.IsLineIntersect(MACDSignals, MACDLines) {
			return true
		}
	}
	return false
}

func (ma *MACDAdapter) checkForSellSignal() bool {
	MACDLines := make([]float64, len(ma.HistoricalValues))
	MACDSignals := make([]float64, len(ma.HistoricalValues))
	for i, hv := range ma.HistoricalValues {
		MACDLines[i] = hv.MACDLine
		MACDSignals[i] = hv.MACDSignal
	}

	if ma.HistoricalValues[0].MACDLine > ma.HistoricalValues[0].MACDSignal {
		if utils.IsLineIntersect(MACDSignals, MACDLines) {
			return true
		}
	}
	return false
}
