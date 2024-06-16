package model

import "github.com/vd09/trading-algorithm-backtesting-system/utils"

type Timespan string

const (
	Second Timespan = "second"
	Minute Timespan = "minute"
	Hour   Timespan = "hour"
	Day    Timespan = "day"
)

type HistoricalDataRequest struct {
	Ticker    string
	Interval  int
	Timespan  Timespan
	StartDate utils.TimeUtil
	EndDate   utils.TimeUtil
}
