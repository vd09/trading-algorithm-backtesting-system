package indicator_adaptor

import "github.com/vd09/trading-algorithm-backtesting-system/model"

type IndicatorAdaptor interface {
	Name() string
	Clone() IndicatorAdaptor
	AddDataPoint(data model.DataPoint) error
	GetSignal() model.StockAction
}
