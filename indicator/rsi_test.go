package indicator

import (
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func TestRSI(t *testing.T) {
	rsi := NewRSI(14)

	data := []model.DataPoint{
		{Time: 1, Close: 44.34},
		{Time: 2, Close: 44.09},
		{Time: 3, Close: 44.15},
		{Time: 4, Close: 43.61},
		{Time: 5, Close: 44.33},
		{Time: 6, Close: 44.83},
		{Time: 7, Close: 45.10},
		{Time: 8, Close: 45.42},
		{Time: 9, Close: 45.84},
		{Time: 10, Close: 46.08},
		{Time: 11, Close: 45.89},
		{Time: 12, Close: 46.03},
		{Time: 13, Close: 45.61},
		{Time: 14, Close: 46.28},
		{Time: 15, Close: 46.28},
	}

	for _, dp := range data {
		err := rsi.AddDataPoint(dp)
		if err != nil {
			t.Errorf("Error adding data point: %v", err)
		}
	}

	expectedRSI := 70.464
	actualRSI := rsi.CalculateRSI()
	if actualRSI != expectedRSI {
		t.Errorf("Expected RSI: %v, got: %v", expectedRSI, actualRSI)
	}
}
