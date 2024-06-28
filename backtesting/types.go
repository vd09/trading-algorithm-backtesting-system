package backtesting

//
//type BacktestEngine struct {
//	Algorithms     []algorithm.TradingAlgorithm
//	Performance    map[string]PerformanceMetrics
//	HistoricalData []model.DataPoint
//}

type PerformanceMetrics struct {
	TotalRequests     int
	SuccessfulSignals int
	FailedSignals     int
	TotalProfit       float64
	ProfitPercentage  float64
}
