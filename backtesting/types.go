package backtesting

import "github.com/vd09/trading-algorithm-backtesting-system/model"

type PerformanceMetrics struct {
	Trades             int
	MaxProfit          float64
	MinProfit          float64
	ActivePositions    []OpenPosition
	CompletedPositions []OpenPosition
}

type OpenPosition struct {
	EntryPoint      model.DataPoint
	Signal          model.TradingSignal
	TotalPeEarnings int
	CurrentProfit   float64
	IterationCount  int
	IterationData   []IterationData
}

type IterationData struct {
	Time            int64
	Price           float64
	Profit          float64
	IterationNumber int
}

// IterationSummaryMetrics holds performance data for each iteration.
type IterationSummaryMetrics struct {
	IterationNumber         int
	Trades                  int
	Wins                    int
	TotalProfitPercentage   float64
	AverageProfitPercentage float64
	MaxProfitPercentage     float64
	WinRate                 float64
}
