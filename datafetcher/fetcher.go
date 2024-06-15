package datafetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func FetchHistoricalData(request model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	if !isValidDate(request.StartDate) || !isValidDate(request.EndDate) {
		return nil, fmt.Errorf("invalid date format, must be YYYY-MM-DD")
	}

	url := fmt.Sprintf("%s/%s/range/%d/%s/%s/%s?apiKey=%s",
		config.AppConfig.PolygonAPIURL, request.Ticker, request.Interval, request.Timespan, request.StartDate, request.EndDate, config.AppConfig.PolygonAPIKey)

	log.Printf("Requesting URL: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Response: %s\n", string(body))

	var polygonResponse model.PolygonResponse
	if err := json.Unmarshal(body, &polygonResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &polygonResponse, nil
}

func GetHistoricalData(request model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	existingFiles, err := getExistingDataFiles(request.Ticker)
	if err != nil {
		return nil, err
	}

	var combinedData *model.PolygonResponse
	var missingDateRanges []dateRange

	for _, file := range existingFiles {
		if fileCoversRange(file, request.StartDate, request.EndDate, request.Interval, request.Timespan) {
			data, err := readDataFromCSV(file)
			if err != nil {
				return nil, err
			}
			if combinedData == nil {
				combinedData = data
			} else {
				combinedData = mergeData(combinedData, data)
			}
		}
	}

	if combinedData != nil {
		missingDateRanges = getMissingDateRanges(request, combinedData)
	} else {
		missingDateRanges = append(missingDateRanges, dateRange{request.StartDate, request.EndDate})
	}

	var newData *model.PolygonResponse
	for _, dateRange := range missingDateRanges {
		partialRequest := model.HistoricalDataRequest{
			Ticker:    request.Ticker,
			Interval:  request.Interval,
			Timespan:  request.Timespan,
			StartDate: dateRange.Start,
			EndDate:   dateRange.End,
		}
		partialData, err := FetchHistoricalData(partialRequest)
		if err != nil {
			return nil, err
		}
		if newData == nil {
			newData = partialData
		} else {
			newData = mergeData(newData, partialData)
		}
	}

	if combinedData != nil {
		newData = mergeData(combinedData, newData)
	}

	filename := fmt.Sprintf("data/%s_%s_to_%s_%d_%s.csv", request.Ticker, request.StartDate, request.EndDate, request.Interval, request.Timespan)
	if err := SaveDataToCSV(newData, filename); err != nil {
		return nil, err
	}

	for _, file := range existingFiles {
		if fileCoversRange(file, request.StartDate, request.EndDate, request.Interval, request.Timespan) {
			if err := os.Remove(file); err != nil {
				log.Printf("Error deleting file: %v", err)
			}
		}
	}

	return newData, nil
}

func isValidDate(date string) bool {
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	return re.MatchString(date)
}
