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

// TestNewMACDAdapter tests the initialization of a new MACDAdapter instance.
func TestNewMACDAdapter(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewMACDAdapter(context.Background(), 12, 26, 9, 100, mock)

	if adapter.MACD.ShortPeriod != 12 {
		t.Errorf("expected ShortPeriod 12, got %d", adapter.MACD.ShortPeriod)
	}
	if adapter.MACD.LongPeriod != 26 {
		t.Errorf("expected LongPeriod 26, got %d", adapter.MACD.LongPeriod)
	}
	if adapter.MACD.SignalPeriod != 9 {
		t.Errorf("expected SignalPeriod 9, got %d", adapter.MACD.SignalPeriod)
	}
}

// TestMACDAddDataPointNotInitialized tests adding data points when the MACD is not yet initialized.
func TestMACDAddDataPointNotInitialized(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewMACDAdapter(context.Background(), 12, 26, 9, 10, mock)

	dataPoints := []model.DataPoint{
		{Time: time.Now().Unix(), Close: 10.0},
		{Time: time.Now().Add(time.Minute * 1).Unix(), Close: 20.0},
		{Time: time.Now().Add(time.Minute * 2).Unix(), Close: 30.0},
		{Time: time.Now().Add(time.Minute * 3).Unix(), Close: 40.0},
		{Time: time.Now().Add(time.Minute * 4).Unix(), Close: 50.0},
		{Time: time.Now().Add(time.Minute * 5).Unix(), Close: 60.0},
	}

	ctx := context.Background()
	for _, data := range dataPoints {
		err := adapter.AddDataPoint(ctx, data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if adapter.MACD.Initialized {
		t.Errorf("expected MACD to be not initialized")
	}

	if len(adapter.HistoricalValues) != 0 {
		t.Errorf("expected zero historical values, got %d", len(adapter.HistoricalValues))
	}
}

// TestMACDAddDataPointInitialization tests the initialization of the MACD after adding enough data points.
func TestMACDAddDataPointInitialization(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewMACDAdapter(context.Background(), 12, 26, 9, 10, mock)

	ctx := context.Background()
	for i := 0; i < 30; i++ {
		err := adapter.AddDataPoint(ctx, model.DataPoint{Time: time.Now().Add(time.Minute * time.Duration(i)).Unix(), Close: (10.0 * float64(i+1))})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if !adapter.MACD.Initialized {
		t.Errorf("expected MACD to be initialized")
	}

	if len(adapter.HistoricalValues) == 0 {
		t.Errorf("expected non-zero historical values")
	}
}

// TestMACDName tests the generation of the MACDAdapter's name.
func TestMACDName(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewMACDAdapter(context.Background(), 12, 26, 9, 10, mock)

	expectedName := "MACD_12_26_9"
	if adapter.Name() != expectedName {
		t.Errorf("expected name %s, got %s", expectedName, adapter.Name())
	}
}

// TestMACDGetSignal tests the signal generation logic of the MACDAdapter.
func TestMACDGetSignal(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewMACDAdapter(context.Background(), 12, 26, 9, 10, mock)

	dataPoints := []model.DataPoint{}
	for i := 0; i < 30; i++ {
		dataPoints = append(dataPoints, model.DataPoint{Time: time.Now().Add(time.Minute * time.Duration(i)).Unix(), Close: (10.0 * float64(i+1))})
	}

	ctx := context.Background()
	for _, data := range dataPoints {
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

	// Simulate a buy signal scenario
	adapter.HistoricalValues = make([]indicator.MACDResult, 0)
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 5.0, MACDSignal: 10.0, MACDHistogram: 5.0,
	})
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 10.0, MACDSignal: 15.0, MACDHistogram: 5.0,
	})
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 25.0, MACDSignal: 20.0, MACDHistogram: 5.0,
	})
	signal = adapter.GetSignal(ctx)
	if signal != model.Buy {
		t.Errorf("expected signal %v, got %v", model.Buy, signal)
	}

	// Simulate a sell signal scenario
	adapter.HistoricalValues = make([]indicator.MACDResult, 0)
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 10.0, MACDSignal: 5.0, MACDHistogram: 5.0,
	})
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 15.0, MACDSignal: 10.0, MACDHistogram: 5.0,
	})
	adapter.HistoricalValues = append(adapter.HistoricalValues, indicator.MACDResult{
		MACDLine: 20.0, MACDSignal: 25.0, MACDHistogram: 5.0,
	})
	signal = adapter.GetSignal(ctx)
	if signal != model.Sell {
		t.Errorf("expected signal %v, got %v", model.Sell, signal)
	}
}
