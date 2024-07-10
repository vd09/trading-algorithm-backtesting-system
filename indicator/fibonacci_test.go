package indicator

import (
	"context"
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils/test_utils"
)

func TestFibonacci(t *testing.T) {
	fib := NewFibonacci(5)

	dataPoints := []model.DataPoint{
		{Time: 1, High: 100, Low: 90, Close: 95},
		{Time: 2, High: 110, Low: 95, Close: 105},
		{Time: 3, High: 120, Low: 100, Close: 115},
		{Time: 4, High: 130, Low: 110, Close: 125},
		{Time: 5, High: 140, Low: 120, Close: 135},
	}

	ctx := context.Background()
	for _, dp := range dataPoints {
		err := fib.AddDataPoint(ctx, dp)
		if err != nil {
			t.Fatalf("Failed to add data point: %v", err)
		}
	}

	levels := fib.GetFibonacciLevels()
	expectedLevels := map[FibonacciLevel]float64{
		Zero:        140,
		TwentyThree: 127.2,
		ThirtyEight: 122.36,
		Fifty:       115,
		SixtyOne:    107.64,
		Hundred:     90,
	}

	for level, value := range expectedLevels {
		test_utils.AssertEqual(t, levels[level], value, "")
	}
}
