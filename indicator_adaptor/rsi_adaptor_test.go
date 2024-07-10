package indicator_adaptor_test

import (
	"context"
	"testing"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils/test_utils"
)

// TestNewRSIAdapter tests the initialization of a new RSIAdapter instance.
func TestNewRSIAdapter(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewRSIAdapter(context.Background(), 14, 100, 70.0, 30.0, mock)

	if adapter.RSI.Period != 14 {
		t.Errorf("expected Period 14, got %d", adapter.RSI.Period)
	}
	if adapter.MaxTotalHistoricalData != 100 {
		t.Errorf("expected MaxTotalHistoricalData 100, got %d", adapter.MaxTotalHistoricalData)
	}
	if adapter.OverboughtThreshold != 70.0 {
		t.Errorf("expected OverboughtThreshold 70.0, got %v", adapter.OverboughtThreshold)
	}
	if adapter.OversoldThreshold != 30.0 {
		t.Errorf("expected OversoldThreshold 30.0, got %v", adapter.OversoldThreshold)
	}
}

// TestAddDataPoint tests adding a new data point to the RSIAdapter.
func TestRSIAdapterAddDataPoint(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewRSIAdapter(context.Background(), 14, 100, 70.0, 30.0, mock)

	dataPoint := model.DataPoint{
		Time:  time.Now().Unix(),
		Close: 120.0,
	}

	ctx := context.Background()
	err := adapter.AddDataPoint(ctx, dataPoint)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if adapter.RSI.Initialized {
		t.Errorf("expected Initialized to be false, got true")
	}

	// Add more data points to initialize the RSI
	for i := 1; i < 14; i++ {
		dataPoint = model.DataPoint{
			Time:  time.Now().Add(time.Minute * time.Duration(i)).Unix(),
			Close: 120.0 + float64(i),
		}
		err = adapter.AddDataPoint(ctx, dataPoint)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if !adapter.RSI.Initialized {
		t.Errorf("expected Initialized to be true, got false")
	}
}

// TestGetSignal tests the signal generation logic of the RSIAdapter.
func TestRSIAdapterGetSignal(t *testing.T) {
	mock := test_utils.NewMockMetricsCollector(t)
	adapter := indicator_adaptor.NewRSIAdapter(context.Background(), 14, 100, 70.0, 30.0, mock)

	ctx := context.Background()
	dataPoints := []model.DataPoint{}
	for i := 0; i < 20; i++ {
		dataPoints = append(dataPoints, model.DataPoint{
			Time:  time.Now().Add(time.Minute * time.Duration(i)).Unix(),
			Close: 120.0 + float64(i*3),
		})
	}

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

	// Simulate an overbought scenario for sell signal
	adapter.HistoricalValues = make([]float64, 15)
	for i := 0; i < 14; i++ {
		adapter.HistoricalValues[i] = 75.0
	}
	adapter.HistoricalValues[14] = 65.0 // Cross below the overbought threshold

	signal = adapter.GetSignal(ctx)
	if signal != model.Sell {
		t.Errorf("expected signal %v, got %v", model.Sell, signal)
	}

	// Simulate an oversold scenario for buy signal
	adapter.HistoricalValues = make([]float64, 15)
	for i := 0; i < 14; i++ {
		adapter.HistoricalValues[i] = 25.0
	}
	adapter.HistoricalValues[14] = 35.0 // Cross above the oversold threshold

	signal = adapter.GetSignal(ctx)
	if signal != model.Buy {
		t.Errorf("expected signal %v, got %v", model.Buy, signal)
	}
}
