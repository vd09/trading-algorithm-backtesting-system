package backtesting

import (
	"github.com/vd09/trading-algorithm-backtesting-system/algorithm"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func (be *BacktestEngine) AddAlgorithm(algo algorithm.TradingAlgorithm) {
	be.Algorithms = append(be.Algorithms, algo)
}

func (be *BacktestEngine) Run() {
	for _, dataPoint := range be.HistoricalData {
		for _, algo := range be.Algorithms {
			signal := algo.Evaluate(dataPoint)
			be.RecordPerformance(algo.Name(), signal, dataPoint)
		}
	}
}

func (be *BacktestEngine) RecordPerformance(algoName string, signal model.TradingSignal, data model.DataPoint) {
	if _, exists := be.Performance[algoName]; !exists {
		be.Performance[algoName] = PerformanceMetrics{}
	}
	metrics := be.Performance[algoName]
	metrics.TotalRequests++

	if signal.Action == model.Buy || signal.Action == model.Sell {
		if isSuccess(signal, data) {
			metrics.SuccessfulSignals++
			profit := calculateProfit(signal, data)
			metrics.TotalProfit += profit
		} else {
			metrics.FailedSignals++
		}
	}
	metrics.ProfitPercentage = (metrics.TotalProfit / float64(metrics.TotalRequests)) * 100
	be.Performance[algoName] = metrics
}

func isSuccess(signal model.TradingSignal, data model.DataPoint) bool {
	// Define the success criteria here
	return true
}

func calculateProfit(signal model.TradingSignal, data model.DataPoint) float64 {
	// Calculate profit based on signal and data
	return 0.0
}
