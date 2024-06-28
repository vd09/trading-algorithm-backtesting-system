package algorithm_test

import (
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/algorithm"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

// MockIndicatorAdaptor is a mock implementation of the IndicatorAdaptor interface
type MockIndicatorAdaptor struct {
	name   string
	signal model.StockAction
}

func (m *MockIndicatorAdaptor) Name() string {
	return m.name
}

func (m *MockIndicatorAdaptor) Clone() indicator_adaptor.IndicatorAdaptor {
	return m
}

func (m *MockIndicatorAdaptor) AddDataPoint(data model.DataPoint) error {
	return nil
}

func (m *MockIndicatorAdaptor) GetSignal() model.StockAction {
	return m.signal
}

func TestCreateCombinationTradingAlgorithms(t *testing.T) {
	adaptor1 := &MockIndicatorAdaptor{name: "Adaptor1", signal: model.Buy}
	adaptor2 := &MockIndicatorAdaptor{name: "Adaptor2", signal: model.Sell}
	adaptor3 := &MockIndicatorAdaptor{name: "Adaptor3", signal: model.Wait}
	adaptor4 := &MockIndicatorAdaptor{name: "Adaptor4", signal: model.Wait}
	adaptor5 := &MockIndicatorAdaptor{name: "Adaptor5", signal: model.Wait}

	adaptors := []indicator_adaptor.IndicatorAdaptor{adaptor1, adaptor2, adaptor3, adaptor4, adaptor5}
	tradingAlgorithms := algorithm.CreateCombinationTradingAlgorithms(adaptors)

	expectedNames := []string{
		"Adaptor1",
		"Adaptor2",
		"Adaptor3",
		"Adaptor4",
		"Adaptor5",
		"Adaptor1_Adaptor2",
		"Adaptor1_Adaptor3",
		"Adaptor1_Adaptor4",
		"Adaptor2_Adaptor3",
		"Adaptor2_Adaptor4",
		"Adaptor3_Adaptor4",
		"Adaptor1_Adaptor2_Adaptor3",
		"Adaptor1_Adaptor2_Adaptor4",
		"Adaptor1_Adaptor3_Adaptor4",
		"Adaptor2_Adaptor3_Adaptor4",
		"Adaptor1_Adaptor2_Adaptor3_Adaptor4",
		"Adaptor1_Adaptor5",
		"Adaptor2_Adaptor5",
		"Adaptor3_Adaptor5",
		"Adaptor4_Adaptor5",
		"Adaptor1_Adaptor2_Adaptor5",
		"Adaptor1_Adaptor3_Adaptor5",
		"Adaptor1_Adaptor4_Adaptor5",
		"Adaptor2_Adaptor3_Adaptor5",
		"Adaptor2_Adaptor4_Adaptor5",
		"Adaptor3_Adaptor4_Adaptor5",
		"Adaptor1_Adaptor2_Adaptor3_Adaptor5",
		"Adaptor1_Adaptor2_Adaptor4_Adaptor5",
		"Adaptor1_Adaptor3_Adaptor4_Adaptor5",
		"Adaptor2_Adaptor3_Adaptor4_Adaptor5",
		"Adaptor1_Adaptor2_Adaptor3_Adaptor4_Adaptor5",
	}

	// Create a map of expected names for easy lookup
	expectedNamesMap := make(map[string]bool)
	for _, name := range expectedNames {
		expectedNamesMap[name] = true
	}

	// Assert that the number of generated algorithms matches the expected number
	utils.AssertEqual(t, len(expectedNames), len(tradingAlgorithms), "Number of generated algorithms does not match")

	// Assert that each generated algorithm name is in the expected names map
	for _, algo := range tradingAlgorithms {
		utils.AssertTrue(t, expectedNamesMap[algo.Name()], "Algorithm name does not match any expected name "+algo.Name())
	}
}

func TestCombinationTradingAlgorithm_Evaluate(t *testing.T) {
	adaptor1 := &MockIndicatorAdaptor{name: "Adaptor1", signal: model.Buy}
	adaptor2 := &MockIndicatorAdaptor{name: "Adaptor2", signal: model.Buy}
	adaptor3 := &MockIndicatorAdaptor{name: "Adaptor3", signal: model.Buy}

	adaptors := []indicator_adaptor.IndicatorAdaptor{adaptor1, adaptor2, adaptor3}
	algorithm := algorithm.NewCombinationTradingAlgorithm(adaptors)

	data := model.DataPoint{Time: 1625097600}
	signal := algorithm.Evaluate(data)

	// Assert that the signal is a buy signal
	utils.AssertEqual(t, model.Buy, signal.Action, "Expected Buy signal")
}

func TestCombinationTradingAlgorithm_EvaluateMixed(t *testing.T) {
	adaptor1 := &MockIndicatorAdaptor{name: "Adaptor1", signal: model.Buy}
	adaptor2 := &MockIndicatorAdaptor{name: "Adaptor2", signal: model.Sell}
	adaptor3 := &MockIndicatorAdaptor{name: "Adaptor3", signal: model.Wait}

	adaptors := []indicator_adaptor.IndicatorAdaptor{adaptor1, adaptor2, adaptor3}
	algorithm := algorithm.NewCombinationTradingAlgorithm(adaptors)

	data := model.DataPoint{Time: 1625097600}
	signal := algorithm.Evaluate(data)

	// Assert that the signal is a wait signal
	utils.AssertEqual(t, model.Wait, signal.Action, "Expected Wait signal")
}
