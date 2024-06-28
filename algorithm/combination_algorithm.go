package algorithm

import (
	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// CombinationTradingAlgorithm implements the TradingAlgorithm interface
type CombinationTradingAlgorithm struct {
	adaptors []indicator_adaptor.IndicatorAdaptor
}

// NewCombinationTradingAlgorithm creates a new instance of CombinationTradingAlgorithm
func NewCombinationTradingAlgorithm(adaptors []indicator_adaptor.IndicatorAdaptor) *CombinationTradingAlgorithm {
	return &CombinationTradingAlgorithm{
		adaptors: adaptors,
	}
}

// Name returns the name of the trading algorithm
func (ta *CombinationTradingAlgorithm) Name() string {
	name := ""
	for _, adaptor := range ta.adaptors {
		name += adaptor.Name() + "_"
	}
	return name[:len(name)-1] // Remove the trailing underscore
}

// Evaluate evaluates the signals from all adaptors and determines the action
func (ta *CombinationTradingAlgorithm) Evaluate(data model.DataPoint) model.TradingSignal {
	buyCount, sellCount := 0, 0
	for _, adaptor := range ta.adaptors {
		adaptor.AddDataPoint(data)
		signal := adaptor.GetSignal()
		switch signal {
		case model.Buy:
			buyCount++
		case model.Sell:
			sellCount++
		case model.Wait:
			return model.TradingSignal{Time: data.Time, Action: model.Wait}
		}
	}

	if buyCount == len(ta.adaptors) {
		return model.TradingSignal{Time: data.Time, Action: model.Buy}
	} else if sellCount == len(ta.adaptors) {
		return model.TradingSignal{Time: data.Time, Action: model.Sell}
	} else {
		return model.TradingSignal{Time: data.Time, Action: model.Wait}
	}
}

func CreateCombinationTradingAlgorithms(adaptors []indicator_adaptor.IndicatorAdaptor) []TradingAlgorithm {
	var tradingAlgorithms []TradingAlgorithm
	var generate func([]indicator_adaptor.IndicatorAdaptor, int)

	generate = func(currentCombination []indicator_adaptor.IndicatorAdaptor, start int) {
		if len(currentCombination) > 0 {
			combo := append([]indicator_adaptor.IndicatorAdaptor{}, currentCombination...)
			clonedCombo := make([]indicator_adaptor.IndicatorAdaptor, len(combo))
			for i, adaptor := range combo {
				clonedCombo[i] = adaptor.Clone()
			}
			tradingAlgorithms = append(tradingAlgorithms, NewCombinationTradingAlgorithm(clonedCombo))
		}
		for i := start; i < len(adaptors); i++ {
			generate(append(currentCombination, adaptors[i]), i+1)
		}
	}

	generate([]indicator_adaptor.IndicatorAdaptor{}, 0)
	return tradingAlgorithms
}
