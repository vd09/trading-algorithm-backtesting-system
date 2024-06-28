package indicator_adaptor_test

import (
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func TestAddDataPoint_NonChronological(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data1 := model.DataPoint{Time: 2, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}
	data2 := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}

	err := sta.AddDataPoint(data1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = sta.AddDataPoint(data2)
	if err == nil || err.Error() != "data point is not in chronological order" {
		t.Fatalf("Expected error for non-chronological data point, got: %v", err)
	}
}

func TestAddDataPoint_DuplicateTimestamps(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data1 := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}
	data2 := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}

	err := sta.AddDataPoint(data1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = sta.AddDataPoint(data2)
	if err == nil || err.Error() != "data point is not in chronological order" {
		t.Fatalf("Expected error for duplicate timestamp, got: %v", err)
	}
}

func TestAddDataPoint_InsufficientData(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}

	for i := 1; i < 14; i++ {
		data.Time = int64(i)
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	if sta.SuperTrend.Initialized {
		t.Fatalf("Expected SuperTrend to be uninitialized")
	}
}

func TestAddDataPoint_ExtremeValues(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)

	for i := 1; i <= 14; i++ {
		data := model.DataPoint{
			Time:   int64(i),
			Open:   1e10,
			High:   1e10 + 10,
			Low:    1e10 - 10,
			Close:  1e10,
			Volume: 1000,
		}
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	if !sta.SuperTrend.Initialized {
		t.Fatalf("Expected SuperTrend to be initialized")
	}
}

func TestAddDataPoint_NegativeOrZeroPrices(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data := model.DataPoint{Time: 1, Open: -100, High: 110, Low: 90, Close: 105, Volume: 1000}

	err := sta.AddDataPoint(data)
	if err == nil {
		t.Fatalf("Expected error for negative prices, got nil")
	} else if err.Error() != "data point contains negative or zero prices" {
		t.Fatalf("Unexpected error: %v", err)
	}

	data = model.DataPoint{Time: 2, Open: 0, High: 0, Low: 0, Close: 0, Volume: 1000}
	err = sta.AddDataPoint(data)
	if err == nil {
		t.Fatalf("Expected error for zero prices, got nil")
	} else if err.Error() != "data point contains negative or zero prices" {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAddDataPoint_MissingFields(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data := model.DataPoint{Time: 1, Open: 100, High: 0, Low: 0, Close: 105, Volume: 1000}

	err := sta.AddDataPoint(data)
	if err == nil {
		t.Fatalf("Expected error for missing high and low prices, got nil")
	}
}

func TestAddDataPoint_RapidSuccession(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}

	for i := 1; i <= 1000; i++ {
		data.Time = int64(i)
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	if !sta.SuperTrend.Initialized {
		t.Fatalf("Expected SuperTrend to be initialized")
	}
}

func TestAddDataPoint_NullDataPoint(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	var data *model.DataPoint = nil

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for nil data point, got nil")
		}
	}()

	sta.AddDataPoint(*data)
}

func TestGetSignal_BeforeInitialization(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	signal := sta.GetSignal()
	if signal != model.Wait {
		t.Fatalf("Expected signal to be Wait before initialization, got %v", signal)
	}
}

func TestGetSignal_AfterInitialization_NoTrendChange(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)
	data := model.DataPoint{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}

	// Add 14 data points to initialize
	for i := 1; i <= 14; i++ {
		data.Time = int64(i)
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	signal := sta.GetSignal()
	if signal != model.Wait {
		t.Fatalf("Expected signal to be Wait after initialization with no trend change, got %v", signal)
	}
}

func TestGetSignal_TrendChangeToUptrend(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 1.5)

	// Add 14 data points to initialize
	for i := 1; i <= 14; i++ {
		data := model.DataPoint{
			Time:   int64(i),
			Open:   100,
			High:   110,
			Low:    90,
			Close:  95, // Set close price to below the calculated SuperTrend line
			Volume: 1000,
		}
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Add data point to switch to uptrend
	data := model.DataPoint{Time: 15, Open: 100, High: 160, Low: 60, Close: 156, Volume: 1000}
	err := sta.AddDataPoint(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	signal := sta.GetSignal()
	if signal != model.Buy {
		t.Fatalf("Expected signal to be Buy, got %v", signal)
	}
}

func TestGetSignal_TrendChangeToDowntrend(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 1)

	// Add 14 data points to initialize
	for i := 1; i <= 14; i++ {
		data := model.DataPoint{
			Time:   int64(i),
			Open:   100,
			High:   110,
			Low:    90,
			Close:  105, // Set close price to above the calculated SuperTrend line
			Volume: 1000,
		}
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
	err := sta.AddDataPoint(model.DataPoint{Time: 15, Open: 100, High: 160, Low: 60, Close: 156, Volume: 1000})

	// Add data point to switch to downtrend
	data := model.DataPoint{Time: 16, Open: 100, High: 180, Low: 50, Close: 80, Volume: 1000}
	err = sta.AddDataPoint(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	signal := sta.GetSignal()
	if signal != model.Sell {
		t.Fatalf("Expected signal to be Sell, got %v", signal)
	}
}

func TestGetSignal_NoTrendChange(t *testing.T) {
	sta := indicator_adaptor.NewSuperTrendAdapter(14, 3.0)

	// Add 14 data points to initialize
	for i := 1; i <= 14; i++ {
		data := model.DataPoint{
			Time:   int64(i),
			Open:   100,
			High:   110,
			Low:    90,
			Close:  105, // Set close price to above the calculated SuperTrend line
			Volume: 1000,
		}
		err := sta.AddDataPoint(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Add data point with no trend change
	data := model.DataPoint{Time: 15, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000}
	err := sta.AddDataPoint(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	signal := sta.GetSignal()
	if signal != model.Wait {
		t.Fatalf("Expected signal to be Wait, got %v", signal)
	}
}
