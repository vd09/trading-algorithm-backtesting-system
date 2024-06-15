package model

type Timespan string

const (
	Minute Timespan = "minute"
	Hour   Timespan = "hour"
	Day    Timespan = "day"
)

type HistoricalDataRequest struct {
	Ticker    string
	Interval  int
	Timespan  Timespan
	StartDate string
	EndDate   string
}
