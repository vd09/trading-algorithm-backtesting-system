package indicator_adaptor

import (
	"context"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

type IndicatorAdaptor interface {
	Name() string
	Clone(ctx context.Context) IndicatorAdaptor
	AddDataPoint(ctx context.Context, data model.DataPoint) error
	GetSignal(ctx context.Context) model.StockAction
}
