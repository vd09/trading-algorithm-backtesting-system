package algorithm

import "github.com/vd09/trading-algorithm-backtesting-system/model"

type TradingAlgorithm interface {
	Name() string
	Evaluate(data model.DataPoint) model.TradingSignal
}
