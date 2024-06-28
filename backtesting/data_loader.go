package backtesting

import (
	"github.com/vd09/trading-algorithm-backtesting-system/algorithm"
	"github.com/vd09/trading-algorithm-backtesting-system/datafetcher"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

type BacktestEngine struct {
	Algorithms     []algorithm.TradingAlgorithm
	Performance    map[string]PerformanceMetrics
	HistoricalData []model.DataPoint
}

func (be *BacktestEngine) LoadData(request *model.HistoricalDataRequest) error {
	data, err := be.loadHistoricalData(request)
	if err != nil {
		return err
	}
	be.HistoricalData = data.Results
	return nil
}

func (be *BacktestEngine) loadHistoricalData(request *model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	return datafetcher.GetHistoricalData(request)
}
