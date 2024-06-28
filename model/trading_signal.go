package model

// StockAction represents the possible actions in the stock market.
type StockAction int

const (
	Buy StockAction = iota
	Sell
	Wait
)

type TradingSignal struct {
	Time   int64
	Action StockAction
}
