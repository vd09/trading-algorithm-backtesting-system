package indicator_adaptor

import (
	"math"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

// PivotPointAdapter handles a single PivotPoint indicator and maintains historical values.
type PivotPointAdapter struct {
	PivotPoint             *indicator.PivotPoint
	MaxTotalHistoricalData int
	HistoricalValues       []model.DataPoint
	CurrentData            model.DataPoint
	Threshold              int
}

// NewPivotPointAdapter initializes a new PivotPointAdapter instance.
func NewPivotPointAdapter(maxTotalHistoricalData, threshold int) *PivotPointAdapter {
	return &PivotPointAdapter{
		PivotPoint:             indicator.NewPivotPoint(),
		HistoricalValues:       []model.DataPoint{},
		MaxTotalHistoricalData: maxTotalHistoricalData,
		Threshold:              threshold,
	}
}

func (ppa *PivotPointAdapter) Clone() IndicatorAdaptor {
	return NewPivotPointAdapter(ppa.MaxTotalHistoricalData, ppa.Threshold)
}

// AddDataPoint adds a new data point and updates the PivotPoint.
func (ppa *PivotPointAdapter) AddDataPoint(data model.DataPoint) error {
	if err := ppa.PivotPoint.AddDataPoint(data); err != nil {
		return err
	}
	ppa.CurrentData = data
	ppa.updateHistoricalValues()
	return nil
}

func (ppa *PivotPointAdapter) updateHistoricalValues() {
	ppa.PivotPoint.GetPivotLevels()
	ppa.HistoricalValues = append(ppa.HistoricalValues, ppa.CurrentData)
	if len(ppa.HistoricalValues) > ppa.MaxTotalHistoricalData {
		ppa.HistoricalValues = ppa.HistoricalValues[1:]
	}
}

// Name returns the name of the PivotPoint adapter.
func (ppa *PivotPointAdapter) Name() string {
	return "PivotPoint"
}

// GetSignal returns a trading signal based on the PivotPoint logic.
func (ppa *PivotPointAdapter) GetSignal() model.StockAction {
	if len(ppa.HistoricalValues) == 0 {
		return model.Wait
	}
	if !ppa.PivotPoint.Initialized {
		return model.Wait
	}

	lastData := ppa.CurrentData
	lastPivot := ppa.PivotPoint.Levels

	recentTests := ppa.calculateRecentTests()

	// Evaluate buy signals based on resistance levels
	if ppa.evaluateBuySignal(lastData, lastPivot, recentTests) {
		return model.Buy
	}

	// Evaluate sell signals based on support levels
	if ppa.evaluateSellSignal(lastData, lastPivot, recentTests) {
		return model.Sell
	}

	return model.Wait
}

func (ppa *PivotPointAdapter) calculateRecentTests() map[string]int {
	recentTests := map[string]int{
		"Resistance1": 0,
		"Resistance2": 0,
		"Resistance3": 0,
		"Support1":    0,
		"Support2":    0,
		"Support3":    0,
	}

	pivot := ppa.PivotPoint.Levels
	for i := len(ppa.HistoricalValues) - 1; i >= 0 && i >= len(ppa.HistoricalValues)-5; i-- {
		dataPoint := ppa.HistoricalValues[i]
		if isWithinRange(dataPoint.Close, pivot.Pivot, pivot.Resistance1, pivot.Resistance2) {
			recentTests["Resistance1"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Resistance1, pivot.Resistance2, pivot.Resistance3) {
			recentTests["Resistance2"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Resistance2, pivot.Resistance3, 0) {
			recentTests["Resistance3"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Support2, pivot.Support1, pivot.Pivot) {
			recentTests["Support1"]++
		}
		if isWithinRange(dataPoint.Close, pivot.Support3, pivot.Support2, pivot.Support1) {
			recentTests["Support2"]++
		}
		if isWithinRange(dataPoint.Close, 0, pivot.Support3, pivot.Support2) {
			recentTests["Support3"]++
		}
	}

	return recentTests
}

func (ppa *PivotPointAdapter) evaluateBuySignal(lastData model.DataPoint, lastPivot indicator.PivotLevels, recentTests map[string]int) bool {
	if recentTests["Resistance3"] >= ppa.Threshold && lastData.High > lastPivot.Resistance3 && lastData.Low < lastPivot.Resistance3 && lastData.Close > lastPivot.Resistance3 {
		return true
	}
	if recentTests["Resistance2"] >= ppa.Threshold && lastData.High > lastPivot.Resistance2 && lastData.Low < lastPivot.Resistance2 && lastData.Close > lastPivot.Resistance2 {
		return true
	}
	if recentTests["Resistance1"] >= ppa.Threshold && lastData.High > lastPivot.Resistance1 && lastData.Low < lastPivot.Resistance1 && lastData.Close > lastPivot.Resistance1 {
		return true
	}
	return false
}

func (ppa *PivotPointAdapter) evaluateSellSignal(lastData model.DataPoint, lastPivot indicator.PivotLevels, recentTests map[string]int) bool {
	if recentTests["Support3"] >= ppa.Threshold && lastData.Low < lastPivot.Support3 && lastData.High > lastPivot.Support3 && lastData.Close < lastPivot.Support3 {
		return true
	}
	if recentTests["Support2"] >= ppa.Threshold && lastData.Low < lastPivot.Support2 && lastData.High > lastPivot.Support2 && lastData.Close < lastPivot.Support2 {
		return true
	}
	if recentTests["Support1"] >= ppa.Threshold && lastData.Low < lastPivot.Support1 && lastData.High > lastPivot.Support1 && lastData.Close < lastPivot.Support1 {
		return true
	}
	return false
}

func isWithinRange(value, lowerReference, middleReference, upperReference float64) bool {
	diff := utils.Min(math.Abs(middleReference-lowerReference), math.Abs(upperReference-middleReference)) * 0.25
	lowerBound := middleReference - diff
	upperBound := middleReference + diff
	return value >= lowerBound && value <= upperBound
}
