package indicator_adaptor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

// EMAAdapter handles multiple EMAs for different periods and maintains historical values.
type EMAAdapter struct {
	EMAs                   map[int]*indicator.EMA
	MaxTotalHistoricalData int
	HistoricalValues       map[int][]float64
	periods                []int
}

// NewEMAAdapter initializes a new EMAAdapter instance.
func NewEMAAdapter(periods []int, maxTotalHistoricalData int) *EMAAdapter {
	emas := make(map[int]*indicator.EMA)
	historicalValues := make(map[int][]float64)
	for _, period := range periods {
		emas[period] = indicator.NewEMA(period)
		historicalValues[period] = []float64{}
	}
	sort.Ints(periods)
	return &EMAAdapter{
		EMAs:                   emas,
		HistoricalValues:       historicalValues,
		MaxTotalHistoricalData: maxTotalHistoricalData,
		periods:                periods,
	}
}

func (ea *EMAAdapter) Clone() IndicatorAdaptor {
	return NewEMAAdapter(ea.periods, ea.MaxTotalHistoricalData)
}

// AddDataPoint adds a new data point and updates all EMAs.
func (ea *EMAAdapter) AddDataPoint(data model.DataPoint) error {
	for period, ema := range ea.EMAs {
		ema.AddDataPoint(data)
		if ema.Initialized {
			ea.HistoricalValues[period] = append(ea.HistoricalValues[period], ema.Value)
		}
		if len(ea.HistoricalValues[period]) > ea.MaxTotalHistoricalData {
			ea.HistoricalValues[period] = ea.HistoricalValues[period][1:]
		}
	}
	return nil
}

// Name returns the name of the EMA adapter.
func (ea *EMAAdapter) Name() string {
	return fmt.Sprintf("EMA_%s", ea.getPeriodsString())
}

// GetSignal returns a trading signal based on the EMA crossings logic.
func (ea *EMAAdapter) GetSignal() model.StockAction {
	for _, ema := range ea.EMAs {
		if !ema.Initialized {
			return model.Wait
		}
	}

	HighestPeriod := ea.periods[len(ea.periods)-1]
	if len(ea.HistoricalValues[HighestPeriod]) < 2 {
		return model.Wait
	}

	if ea.checkForBuySignal() {
		return model.Buy
	} else if ea.checkForSellSignal() {
		return model.Sell
	}
	return model.Wait
}

// Helper function to get the periods as a string.
func (ea *EMAAdapter) getPeriodsString() string {
	var periods []string
	for _, period := range ea.periods {
		periods = append(periods, fmt.Sprintf("%d", period))
	}
	return strings.Join(periods, "_")
}

func (ea *EMAAdapter) checkForBuySignal() bool {
	smallestPeriod := ea.periods[0]
	for _, period := range ea.periods[1:] {
		if ea.HistoricalValues[smallestPeriod][0] >= ea.HistoricalValues[period][0] {
			return false
		}
		if !utils.IsLineIntersect(ea.HistoricalValues[smallestPeriod], ea.HistoricalValues[period]) {
			return false
		}
	}
	return true
}

func (ea *EMAAdapter) checkForSellSignal() bool {
	smallestPeriod := ea.periods[0]
	for _, period := range ea.periods[1:] {
		if ea.HistoricalValues[smallestPeriod][0] <= ea.HistoricalValues[period][0] {
			return false
		}
		if !utils.IsLineIntersect(ea.HistoricalValues[smallestPeriod], ea.HistoricalValues[period]) {
			return false
		}
	}
	return true
}
