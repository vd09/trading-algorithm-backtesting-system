package indicator

import (
	"testing"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func TestBollingerBands(t *testing.T) {
	bb := NewBollingerBands(20)

	dataPoints := []model.DataPoint{
		{Time: time.Now().Unix(), Close: 100.0},
		{Time: time.Now().Add(time.Minute * 1).Unix(), Close: 101.0},
		{Time: time.Now().Add(time.Minute * 2).Unix(), Close: 102.0},
		// ... (add more data points to have at least 20 entries)
	}

	// Adding less than 20 data points should not update the Bollinger Bands values
	for _, dp := range dataPoints[:19] {
		err := bb.AddDataPoint(dp)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		values := bb.GetBollingerBands()
		if values.MovingAverage != 0 || values.UpperBand != 0 || values.LowerBand != 0 {
			t.Errorf("expected initial values to be zero")
		}
	}

	// Adding the 20th data point should update the Bollinger Bands values
	err := bb.AddDataPoint(dataPoints[19])
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	values := bb.GetBollingerBands()
	if values.MovingAverage == 0 || values.UpperBand == 0 || values.LowerBand == 0 {
		t.Errorf("expected Bollinger Bands values to be calculated")
	}

	// Adding more data points should keep updating the Bollinger Bands values
	for _, dp := range dataPoints[20:] {
		err := bb.AddDataPoint(dp)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		values = bb.GetBollingerBands()
		if values.MovingAverage == 0 || values.UpperBand == 0 || values.LowerBand == 0 {
			t.Errorf("expected Bollinger Bands values to be updated")
		}
	}

	// Test adding a data point with a timestamp earlier than the last one
	err = bb.AddDataPoint(model.DataPoint{Time: dataPoints[18].Time, Close: 105.0})
	if err == nil {
		t.Errorf("expected error for out-of-order data point, got nil")
	}
}
