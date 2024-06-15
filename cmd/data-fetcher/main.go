package main

import (
	"log"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/datafetcher"
)

func main() {
	config.InitConfig()

	// Load configuration values
	interval := config.AppConfig.Interval

	timespan := datafetcher.Timespan(config.AppConfig.Timespan)
	if !isValidTimespan(timespan) {
		log.Fatalf("Invalid timespan value: %s", timespan)
	}

	request := datafetcher.HistoricalDataRequest{
		Ticker:    config.AppConfig.Ticker,
		Interval:  interval,
		Timespan:  timespan,
		StartDate: config.AppConfig.StartDate,
		EndDate:   config.AppConfig.EndDate,
	}

	response, err := datafetcher.FetchHistoricalData(request)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}

	filename := "data/" + config.AppConfig.Ticker + "_" + config.AppConfig.StartDate + "_to_" + config.AppConfig.EndDate + ".csv"
	if err := datafetcher.SaveDataToCSV(response, filename); err != nil {
		log.Fatalf("Error saving data to CSV: %v", err)
	}

	log.Println("Data fetching and saving completed successfully.")
}

func isValidTimespan(timespan datafetcher.Timespan) bool {
	switch timespan {
	case datafetcher.Minute, datafetcher.Hour, datafetcher.Day:
		return true
	default:
		return false
	}
}
