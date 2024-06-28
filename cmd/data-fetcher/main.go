package main

import (
	"log"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/datafetcher"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

func main() {
	validateGetHistoricalData()
	//validateFetchFunction()
}

func validateGetHistoricalData() {
	config.InitConfig()

	// Load configuration values
	interval := config.AppConfig.Interval

	timespan := model.Timespan(config.AppConfig.Timespan)
	if !isValidTimespan(timespan) {
		log.Fatalf("Invalid timespan value: %s", timespan)
	}

	request := &model.HistoricalDataRequest{
		Ticker:    config.AppConfig.Ticker,
		Interval:  interval,
		Timespan:  timespan,
		StartDate: timeutil("2023-06-07"),
		EndDate:   timeutil("2023-06-14"),
	}
	response, err := datafetcher.GetHistoricalData(request)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	if "2023-06-15" != utils.TimeFromTimeStamp(response.Results[0].Time).StockFormatDate() {
		log.Fatalf("Error starting date not matching")
	}
	//if "2023-06-17" != datafetcher.GetFormatedDateFromTime(response.Results[len(response.Results)-1].Time) {
	//	log.Fatalf("Error end date not matching")
	//}
}

func validateFetchFunction() {
	config.InitConfig()

	// Load configuration values
	interval := config.AppConfig.Interval

	timespan := model.Timespan(config.AppConfig.Timespan)
	if !isValidTimespan(timespan) {
		log.Fatalf("Invalid timespan value: %s", timespan)
	}

	request := &model.HistoricalDataRequest{
		Ticker:    config.AppConfig.Ticker,
		Interval:  interval,
		Timespan:  timespan,
		StartDate: timeutil(config.AppConfig.StartDate),
		EndDate:   timeutil(config.AppConfig.EndDate),
	}

	response, err := datafetcher.FetchHistoricalData(request)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}

	//filename := "data/" + config.AppConfig.StockAction + "_" + config.AppConfig.StartDate + "_to_" + config.AppConfig.EndDate + ".csv"
	//if err := datafetcher.SaveDataToCSV(response, filename); err != nil {
	//	log.Fatalf("Error saving data to CSV: %v", err)
	//}

	log.Println("Data fetching and saving completed successfully.", response)
}

func isValidTimespan(timespan model.Timespan) bool {
	switch timespan {
	case model.Second, model.Minute, model.Hour, model.Day:
		return true
	default:
		return false
	}
}

func timeutil(t string) utils.TimeUtil {
	format, err := utils.NewTimeUtilFromFormat(t)
	if err != nil {
		panic(err)
	}
	return format
}
