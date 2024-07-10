package indicator_adaptor_test

import (
	"context"
	"testing"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils/test_utils"
)

// TestNewPivotPointAdapter tests the initialization of a new PivotPointAdapter instance.
func TestNewPivotPointAdapter(t *testing.T) {
	maxTotalHistoricalData := 100
	threshold := 3

	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewPivotPointAdapter(context.Background(), maxTotalHistoricalData, threshold, mock)

	if adapter.MaxTotalHistoricalData != maxTotalHistoricalData {
		t.Errorf("expected MaxTotalHistoricalData %d, got %d", maxTotalHistoricalData, adapter.MaxTotalHistoricalData)
	}
	if adapter.Threshold != threshold {
		t.Errorf("expected Threshold %d, got %d", threshold, adapter.Threshold)
	}
	if adapter.PivotPoint == nil {
		t.Errorf("expected PivotPoint to be initialized")
	}
}

// TestAddDataPoint tests adding a new data point to the PivotPointAdapter.
func TestPivotPointAddDataPoint(t *testing.T) {
	maxTotalHistoricalData := 100
	threshold := 3

	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewPivotPointAdapter(context.Background(), maxTotalHistoricalData, threshold, mock)

	dataPoint := model.DataPoint{
		Time:  time.Now().Unix(),
		High:  150.0,
		Low:   100.0,
		Close: 120.0,
	}

	ctx := context.Background()
	if err := adapter.AddDataPoint(ctx, dataPoint); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !adapter.PivotPoint.Initialized {
		t.Errorf("expected PivotPoint to be initialized")
	}

	if len(adapter.HistoricalValues) != 1 {
		t.Errorf("expected 1 historical value, got %d", len(adapter.HistoricalValues))
	}

	if adapter.CurrentData != dataPoint {
		t.Errorf("expected CurrentData to be %v, got %v", dataPoint, adapter.CurrentData)
	}
}

// TestGetSignal tests the signal generation logic of the PivotPointAdapter.
func TestPivotPointGetSignal(t *testing.T) {
	maxTotalHistoricalData := 5
	threshold := 0

	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewPivotPointAdapter(context.Background(), maxTotalHistoricalData, threshold, mock)

	dataPoints := generateTestDataPoints(10)

	ctx := context.Background()
	for _, data := range dataPoints {
		if err := adapter.AddDataPoint(ctx, data); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Check the initial signal, should be Wait
	expectedSignal := model.Wait
	verifySignal(t, adapter, expectedSignal)

	// Simulate a buy signal scenario
	adapter.PivotPoint.Levels = indicator.PivotLevels{
		Pivot: 120, Resistance1: 145, Resistance2: 170, Resistance3: 220.0,
	}
	expectedSignal = model.Buy
	verifySignal(t, adapter, expectedSignal)

	// Simulate a sell signal scenario
	adapter.PivotPoint.Levels = indicator.PivotLevels{
		Support3: 120, Support2: 145, Support1: 170, Pivot: 220.0,
	}
	expectedSignal = model.Sell
	verifySignal(t, adapter, expectedSignal)
}

// generateTestDataPoints generates a slice of test DataPoint instances.
func generateTestDataPoints(count int) []model.DataPoint {
	dataPoints := make([]model.DataPoint, count)
	for i := 0; i < count; i++ {
		dataPoints[i] = model.DataPoint{
			Time:  time.Now().Add(time.Minute * time.Duration(i)).Unix(),
			High:  150.0 + (float64(i) * 3.0),
			Low:   100.0 + (float64(i) * 3.0),
			Close: 120.0 + (float64(i) * 3.0),
		}
	}
	return dataPoints
}

// verifySignal checks the signal from the adapter and compares it with the expected signal.
func verifySignal(t *testing.T, adapter *indicator_adaptor.PivotPointAdapter, expectedSignal model.StockAction) {
	ctx := context.Background()
	signal := adapter.GetSignal(ctx)
	if signal != expectedSignal {
		t.Errorf("expected signal %v, got %v", expectedSignal, signal)
	}
}
