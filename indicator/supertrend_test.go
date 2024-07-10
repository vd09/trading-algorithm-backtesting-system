package indicator

import (
	"context"
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func TestSuperTrend(t *testing.T) {
	// Example data points
	dataPoints := []model.DataPoint{
		{Time: 1, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Time: 2, Open: 105, High: 115, Low: 95, Close: 110, Volume: 1100},
		{Time: 3, Open: 110, High: 120, Low: 100, Close: 115, Volume: 1200},
		{Time: 4, Open: 115, High: 125, Low: 105, Close: 120, Volume: 1300},
		{Time: 5, Open: 120, High: 130, Low: 110, Close: 125, Volume: 1400},
		{Time: 6, Open: 125, High: 135, Low: 115, Close: 130, Volume: 1500},
		{Time: 7, Open: 130, High: 140, Low: 120, Close: 135, Volume: 1600},
		{Time: 8, Open: 135, High: 145, Low: 125, Close: 140, Volume: 1700},
		{Time: 9, Open: 140, High: 150, Low: 130, Close: 145, Volume: 1800},
		{Time: 10, Open: 145, High: 155, Low: 135, Close: 150, Volume: 1900},
	}

	st := NewSuperTrend(7, 3.0)

	ctx := context.Background()
	for _, dp := range dataPoints {
		err := st.AddDataPoint(ctx, dp)
		if err != nil {
			t.Errorf("error adding data point: %v", err)
		}

		if len(st.History) >= st.Period {
			superTrendLine, isUpTrend := st.CalculateSuperTrend()
			t.Logf("SuperTrend: %v, IsUpTrend: %v", superTrendLine, isUpTrend)
		}
	}
}
