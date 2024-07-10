package algorithm

import (
	"context"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

type TradingAlgorithm interface {
	Name() string
	Evaluate(ctx context.Context, data model.DataPoint) model.TradingSignal
}
