package model

// StockAction represents the possible actions in the stock market.
type StockAction string

const (
	Wait StockAction = "wait"
	Sell             = "sell"
	Buy              = "buy"
)

type TradingSignal struct {
	Time   int64
	Action StockAction
}
