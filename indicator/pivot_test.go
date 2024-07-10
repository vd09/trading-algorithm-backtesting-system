package indicator

import (
	"context"
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func TestPivotPoint(t *testing.T) {
	pp := NewPivotPoint()

	data := []model.DataPoint{
		{Time: 1, Open: 10, High: 15, Low: 5, Close: 10},
		{Time: 2, Open: 10, High: 15, Low: 5, Close: 12},
		{Time: 3, Open: 12, High: 18, Low: 8, Close: 16},
	}

	ctx := context.Background()
	for _, dp := range data {
		err := pp.AddDataPoint(ctx, dp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	expectedPivot := float64((18 + 8 + 16) / 3)
	expectedSupport1 := 2*expectedPivot - float64(18)
	expectedSupport2 := expectedPivot - (float64(18) - float64(8))
	expectedSupport3 := float64(8) - 2*(float64(18)-expectedPivot)
	expectedResistance1 := 2*expectedPivot - float64(8)
	expectedResistance2 := expectedPivot + (float64(18) - float64(8))
	expectedResistance3 := float64(18) + 2*(expectedPivot-float64(8))

	pivotLevels := pp.GetPivotLevels()

	if pivotLevels.Pivot != expectedPivot {
		t.Errorf("expected pivot %v, got %v", expectedPivot, pivotLevels.Pivot)
	}
	if pivotLevels.Support1 != expectedSupport1 {
		t.Errorf("expected support1 %v, got %v", expectedSupport1, pivotLevels.Support1)
	}
	if pivotLevels.Support2 != expectedSupport2 {
		t.Errorf("expected support2 %v, got %v", expectedSupport2, pivotLevels.Support2)
	}
	if pivotLevels.Support3 != expectedSupport3 {
		t.Errorf("expected support3 %v, got %v", expectedSupport3, pivotLevels.Support3)
	}
	if pivotLevels.Resistance1 != expectedResistance1 {
		t.Errorf("expected resistance1 %v, got %v", expectedResistance1, pivotLevels.Resistance1)
	}
	if pivotLevels.Resistance2 != expectedResistance2 {
		t.Errorf("expected resistance2 %v, got %v", expectedResistance2, pivotLevels.Resistance2)
	}
	if pivotLevels.Resistance3 != expectedResistance3 {
		t.Errorf("expected resistance3 %v, got %v", expectedResistance3, pivotLevels.Resistance3)
	}
}
