package datafetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/datasaver"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func FetchHistoricalData(request *model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	url := fmt.Sprintf("%s/%s/range/%d/%s/%s/%s?apiKey=%s",
		config.AppConfig.PolygonAPIURL, request.Ticker, request.Interval, request.Timespan, request.StartDate.StockFormatDate(), request.EndDate.StockFormatDate(), config.AppConfig.PolygonAPIKey)

	log.Printf("Requesting URL: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching data: %v\n", err)
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v\n", err)
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Response: %s\n", string(body))

	var polygonResponse model.PolygonResponse
	if err := json.Unmarshal(body, &polygonResponse); err != nil {
		log.Printf("Error unmarshalling response: %v\n", err)
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	log.Printf("Fetched data for %s from %s to %s\n", request.Ticker, request.StartDate.StockFormatDate(), request.EndDate.StockFormatDate())

	return &polygonResponse, nil
}

func FetchHistoricalDataDummy(request *model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	body := []byte("{\"ticker\":\"AAPL\",\"queryCount\":3,\"resultsCount\":3,\"adjusted\":true,\"results\":[{\"v\":5.7462782e+07,\"vw\":183.7101,\"o\":183.37,\"c\":183.95,\"h\":184.39,\"l\":182.02,\"t\":1686715200000,\"n\":614097},{\"v\":6.5433166e+07,\"vw\":185.5672,\"o\":183.96,\"c\":186.01,\"h\":186.52,\"l\":183.78,\"t\":1686801600000,\"n\":697003},{\"v\":1.01151225e+08,\"vw\":185.4644,\"o\":186.73,\"c\":184.92,\"h\":186.99,\"l\":184.27,\"t\":1686888000000,\"n\":616272}],\"status\":\"OK\",\"request_id\":\"8c6e1f9a5e931cdfb283ccc755f7c5ba\",\"count\":3}")

	log.Printf("Using dummy data: %s\n", string(body))

	var polygonResponse model.PolygonResponse
	if err := json.Unmarshal(body, &polygonResponse); err != nil {
		log.Printf("Error unmarshalling dummy data: %v\n", err)
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	log.Printf("Fetched dummy data for %s\n", request.Ticker)

	return &polygonResponse, nil
}

func GetHistoricalData(request *model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	log.Printf("Getting historical data for %s from %s to %s\n", request.Ticker, request.StartDate.StockFormatDate(), request.EndDate.StockFormatDate())

	sfd := datasaver.NewStocksFetcherData()

	// Check for missing date ranges
	missingRanges := sfd.GetMissingDateRanges(request)
	log.Printf("Missing date ranges for %s: %v\n", request.Ticker, missingRanges)

	existingData, err := sfd.ReadDataForRequest(request)
	if err != nil {
		log.Printf("Error reading existing data: %v\n", err)
		return nil, err
	}
	if len(missingRanges) == 0 {
		// No missing data, read from existing data
		log.Printf("No missing data for %s. Returning existing data.\n", request.Ticker)
		return filterDataByDateRange(request, existingData), nil
	}

	// Fetch missing data from API
	var combinedData *model.PolygonResponse
	for _, dateRange := range missingRanges {
		apiRequest := &model.HistoricalDataRequest{
			Ticker:    request.Ticker,
			Interval:  request.Interval,
			Timespan:  request.Timespan,
			StartDate: dateRange.Start,
			EndDate:   dateRange.End,
		}

		data, err := FetchHistoricalData(apiRequest)
		//data, err := FetchHistoricalDataDummy(apiRequest)
		if err != nil {
			log.Printf("Error fetching data from API: %v\n", err)
			continue
		}

		if combinedData == nil {
			combinedData = data
		} else {
			combinedData.MergeResponse(data)
		}
	}

	if existingData != nil {
		combinedData.MergeResponse(existingData)
	}
	// Save the fetched data to CSV and update the index
	if err := sfd.SaveDataForRequest(request, combinedData); err != nil {
		log.Printf("Error saving data: %v\n", err)
	}
	return filterDataByDateRange(request, combinedData), nil
}

// Helper function to filter data by date range
func filterDataByDateRange(request *model.HistoricalDataRequest, data *model.PolygonResponse) *model.PolygonResponse {
	if data == nil {
		return nil
	}

	startUnix := request.StartDate.Unix()
	endUnix := request.EndDate.Unix()
	var filteredResults []model.DataPoint // Assuming the data.Results is of type []model.Result
	for _, result := range data.Results {
		if result.Time >= startUnix && result.Time <= endUnix { // Assuming result.Time is the correct attribute to compare dates
			filteredResults = append(filteredResults, result)
		}
	}
	data.Results = filteredResults
	log.Printf("Filtered data for %s from %s to %s. Total results: %d\n", request.Ticker, request.StartDate.StockFormatDate(), request.EndDate.StockFormatDate(), len(filteredResults))
	return data
}
