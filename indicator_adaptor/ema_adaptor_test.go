package indicator_adaptor_test

import (
	"context"
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils/test_utils"
)

// TestNewEMAAdapter tests the initialization of a new EMAAdapter instance.
func TestNewEMAAdapter(t *testing.T) {
	periods := []int{5, 10, 20}
	maxData := 100
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewEMAAdapter(context.Background(), periods, maxData, mock)

	if len(adapter.EMAs) != len(periods) {
		t.Errorf("expected %d EMAs, got %d", len(periods), len(adapter.EMAs))
	}
	if len(adapter.HistoricalValues) != len(periods) {
		t.Errorf("expected %d HistoricalValues, got %d", len(periods), len(adapter.HistoricalValues))
	}
	if adapter.MaxTotalHistoricalData != maxData {
		t.Errorf("expected MaxTotalHistoricalData %d, got %d", maxData, adapter.MaxTotalHistoricalData)
	}
}

// TestAddDataPoint tests adding a new data point to the EMAAdapter.
func TestAddDataPoint(t *testing.T) {
	periods := []int{3, 5}
	maxData := 10
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewEMAAdapter(context.Background(), periods, maxData, mock)

	dataPoints := []model.DataPoint{
		{Close: 10.0},
		{Close: 20.0},
		{Close: 30.0},
		{Close: 40.0},
		{Close: 50.0},
		{Close: 60.0},
	}
	ctx := context.Background()

	for _, data := range dataPoints {
		err := adapter.AddDataPoint(ctx, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	for _, period := range periods {
		if len(adapter.HistoricalValues[period]) == 0 {
			t.Errorf("expected historical values for period %d, got none", period)
		}
	}
}

// TestName tests the generation of the EMAAdapter's name.
func TestName(t *testing.T) {
	periods := []int{3, 5}
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewEMAAdapter(context.Background(), periods, 10, mock)

	expectedName := "EMA_3_5"
	if adapter.Name() != expectedName {
		t.Errorf("expected name %s, got %s", expectedName, adapter.Name())
	}
}

// TestGetSignal tests the signal generation logic of the EMAAdapter.
func TestGetSignal(t *testing.T) {
	periods := []int{3, 5}
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewEMAAdapter(context.Background(), periods, 10, mock)

	// Add initial data points to initialize the EMAs
	initialDataPoints := []model.DataPoint{
		{Close: 10.0},
		{Close: 20.0},
		{Close: 30.0},
		{Close: 40.0},
		{Close: 50.0},
	}
	ctx := context.Background()

	for _, data := range initialDataPoints {
		err := adapter.AddDataPoint(ctx, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Check the initial signal, should be Wait
	signal := adapter.GetSignal(ctx)
	if signal != model.Wait {
		t.Errorf("expected signal %v, got %v", model.Wait, signal)
	}
	test_utils.AssertEqual(t, []float64{20.0, 30.0, 40.0}, adapter.HistoricalValues[3], "HistoricalValues doesn't match for period: 3")
	test_utils.AssertEqual(t, []float64{30.0}, adapter.HistoricalValues[5], "HistoricalValues doesn't match for period: 5")

	// Simulate a buy signal scenario
	adapter.HistoricalValues[3] = []float64{20.0, 30.0, 40.0, 50.0, 60.0}
	adapter.HistoricalValues[5] = test_utils.GiveCrossingLine(adapter.HistoricalValues[3], 5, test_utils.Above)

	signal = adapter.GetSignal(ctx)
	if signal != model.Buy {
		t.Errorf("expected signal %v, got %v", model.Buy, signal)
	}

	// Simulate a sell signal scenario
	adapter.HistoricalValues[3] = []float64{80.0, 70.0, 55.0, 35.0, 25.0}
	adapter.HistoricalValues[5] = test_utils.GiveCrossingLine(adapter.HistoricalValues[3], 5, test_utils.Below)

	signal = adapter.GetSignal(ctx)
	if signal != model.Sell {
		t.Errorf("expected signal %v, got %v", model.Sell, signal)
	}
}
