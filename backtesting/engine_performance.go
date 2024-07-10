package backtesting

import (
	"context"
	"fmt"

	"github.com/vd09/trading-algorithm-backtesting-system/algorithm"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

type BacktestEngine struct {
	Algorithms      []algorithm.TradingAlgorithm
	Performance     map[string]PerformanceMetrics
	HistoricalData  []model.DataPoint
	TrackIterations int
}

func (be *BacktestEngine) AddAlgorithm(algo algorithm.TradingAlgorithm) {
	be.Algorithms = append(be.Algorithms, algo)
}

func (be *BacktestEngine) AddAllAlgorithm(algos []algorithm.TradingAlgorithm) {
	be.Algorithms = append(be.Algorithms, algos...)
}

func (be *BacktestEngine) Run(ctx context.Context) {
	for _, dataPoint := range be.HistoricalData {
		for _, algo := range be.Algorithms {
			signal := algo.Evaluate(ctx, dataPoint)
			be.recordPerformance(algo.Name(), signal, dataPoint)
		}
	}
	be.printIterationPerformance()
}

func (be *BacktestEngine) recordPerformance(algoName string, signal model.TradingSignal, dataPoint model.DataPoint) {
	metrics := be.getPerformanceMetrics(algoName)
	be.handleNewPosition(metrics, signal, dataPoint)

	// Update existing open positions
	completedPositions := be.updateOpenPositions(metrics, dataPoint)

	// Remove completed positions from active positions
	be.filterActivePositions(metrics)
	// Move completed positions to the completed list
	metrics.CompletedPositions = append(metrics.CompletedPositions, completedPositions...)
}

func (be *BacktestEngine) getPerformanceMetrics(algoName string) *PerformanceMetrics {
	metrics, exists := be.Performance[algoName]
	if !exists {
		metrics = PerformanceMetrics{
			ActivePositions:    []OpenPosition{},
			CompletedPositions: []OpenPosition{},
		}
		be.Performance[algoName] = metrics
	}
	return &metrics
}

func (be *BacktestEngine) handleNewPosition(metrics *PerformanceMetrics, signal model.TradingSignal, dataPoint model.DataPoint) {
	if signal.Action == model.Buy || signal.Action == model.Sell {
		newPosition := OpenPosition{
			EntryPoint:      dataPoint,
			Signal:          signal,
			CurrentProfit:   0,
			IterationCount:  0,
			TotalPeEarnings: 0,
			IterationData:   []IterationData{},
		}
		metrics.ActivePositions = append(metrics.ActivePositions, newPosition)
	}
}

func (be *BacktestEngine) updateOpenPositions(metrics *PerformanceMetrics, dataPoint model.DataPoint) []OpenPosition {
	var completedPositions []OpenPosition

	for i := 0; i < len(metrics.ActivePositions); i++ {
		position := &metrics.ActivePositions[i]

		// Calculate profit/loss
		profit := be.calculateProfit(position, dataPoint)
		position.CurrentProfit = profit
		if profit > 0 {
			position.TotalPeEarnings++
		}

		// Add iteration data
		position.IterationCount++
		iterationData := IterationData{
			Time:            dataPoint.Time,
			Price:           dataPoint.Close,
			Profit:          profit,
			IterationNumber: position.IterationCount,
		}
		position.IterationData = append(position.IterationData, iterationData)

		// Update performance metrics
		be.updatePerformanceMetrics(metrics, profit)

		// Check if the position has reached the tracking limit
		if position.IterationCount >= be.TrackIterations {
			metrics.Trades++
			// Move the position to completed positions
			completedPositions = append(completedPositions, *position)
		}
	}
	return completedPositions
}

func (be *BacktestEngine) calculateProfit(position *OpenPosition, dataPoint model.DataPoint) float64 {
	if position.Signal.Action == model.Buy {
		return dataPoint.Close - position.EntryPoint.Close
	} else if position.Signal.Action == model.Sell {
		return position.EntryPoint.Close - dataPoint.Close
	}
	return 0
}

func (be *BacktestEngine) updatePerformanceMetrics(metrics *PerformanceMetrics, profit float64) {
	metrics.MaxProfit = utils.Max(metrics.MaxProfit, profit)
	metrics.MinProfit = utils.Min(metrics.MinProfit, profit)
}

func (be *BacktestEngine) filterActivePositions(metrics *PerformanceMetrics) {
	newActivePositions := []OpenPosition{}
	for _, position := range metrics.ActivePositions {
		if position.IterationCount < be.TrackIterations {
			newActivePositions = append(newActivePositions, position)
		}
	}
	metrics.ActivePositions = newActivePositions
}

func (be *BacktestEngine) printIterationPerformance() {
	for algoName, metrics := range be.Performance {
		iterationSummaryData := be.calculateAlgoIterationSummaryMetrics(metrics)
		be.printIterationMetrics(algoName, metrics, iterationSummaryData)
	}
}

// calculateIterationMetrics calculates the iteration-level metrics for completed positions.
func (be *BacktestEngine) calculateAlgoIterationSummaryMetrics(metrics PerformanceMetrics) map[int]*IterationSummaryMetrics {
	iterationData := make(map[int]*IterationSummaryMetrics)

	for _, position := range metrics.CompletedPositions {
		be.aggregateIterationData(position, iterationData)
	}
	for _, position := range metrics.ActivePositions {
		be.aggregateIterationData(position, iterationData)
	}

	for _, iterMetrics := range iterationData {
		iterMetrics.AverageProfitPercentage = iterMetrics.TotalProfitPercentage / float64(iterMetrics.Trades)
		iterMetrics.WinRate = (float64(iterMetrics.Wins) / float64(iterMetrics.Trades)) * 100
	}
	return iterationData
}

// aggregateIterationData aggregates the iteration data for a position.
func (be *BacktestEngine) aggregateIterationData(position OpenPosition, iterationData map[int]*IterationSummaryMetrics) {
	for _, iterData := range position.IterationData {
		iterMetrics, exists := iterationData[iterData.IterationNumber]
		if !exists {
			iterMetrics = &IterationSummaryMetrics{IterationNumber: iterData.IterationNumber}
			iterationData[iterData.IterationNumber] = iterMetrics
		}
		iterMetrics.Trades++
		profitPercentage := (iterData.Profit / position.EntryPoint.Close) * 100 // Profit in percentage
		iterMetrics.TotalProfitPercentage += profitPercentage
		if iterData.Profit > 0 {
			iterMetrics.Wins++
		}
		iterMetrics.MaxProfitPercentage = utils.Max(iterMetrics.MaxProfitPercentage, profitPercentage)
	}
}

func (be *BacktestEngine) printIterationMetrics(algoName string, metrics PerformanceMetrics, iterationSummaryData map[int]*IterationSummaryMetrics) {
	fmt.Printf("Algorithm: %s\n", algoName)
	fmt.Println("Performance by Iteration:")

	fmt.Printf("|%-13s | %-5s | ", "Position Time", "Signal")
	for i := 1; i <= len(iterationSummaryData); i++ {
		fmt.Printf("%9d | ", i)
	}
	fmt.Println()
	fmt.Println("|---------------------------------------------------------------------------------------------------")

	for _, position := range metrics.ActivePositions {
		fmt.Printf("|%-13d | %-5s | ", position.Signal.Time, position.Signal.Action)
		for _, iterData := range position.IterationData {
			profitPercentage := (iterData.Profit / position.EntryPoint.Close) * 100 // Profit in percentage
			fmt.Printf("%-9.2f | ", profitPercentage)
		}
		fmt.Println()
	}
	for _, position := range metrics.CompletedPositions {
		fmt.Printf("|%-13d | %-5s | ", position.Signal.Time, position.Signal.Action)
		for _, iterData := range position.IterationData {
			profitPercentage := (iterData.Profit / position.EntryPoint.Close) * 100 // Profit in percentage
			fmt.Printf("%-9.2f | ", profitPercentage)
		}
		fmt.Println()
	}

	fmt.Println("|---------------------------------------------------------------------------------------------------")
	printIterationData := func(label string, valueFunc func(*IterationSummaryMetrics) float64) {
		fmt.Printf("|%-21s | ", label)
		for i := 1; i <= len(iterationSummaryData); i++ {
			fmt.Printf(" %-9.2f | ", valueFunc(iterationSummaryData[i]))
		}
		fmt.Println()
	}

	printIterationData("Number of Trades", func(m *IterationSummaryMetrics) float64 {
		return float64(m.Trades)
	})
	printIterationData("Number of Wins", func(m *IterationSummaryMetrics) float64 {
		return float64(m.Wins)
	})
	printIterationData("Avg Profit Percent", func(m *IterationSummaryMetrics) float64 {
		return m.AverageProfitPercentage
	})
	printIterationData("Max Profit Percent", func(m *IterationSummaryMetrics) float64 {
		return m.MaxProfitPercentage
	})
	printIterationData("Win Rate", func(m *IterationSummaryMetrics) float64 {
		return m.WinRate
	})
	fmt.Println("|---------------------------------------------------------------------------------------------------")
	fmt.Println()
	fmt.Println()
}
