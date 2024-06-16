package datasaver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/fileutils"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
)

const (
	STOCKS_FETCHER_DATA_PATH       = "data/stocks_fetcher"
	STOCKS_FETCHER_DATA_INDEX_PATH = "data/stocks_fetcher/index.json"
)

type dateRange struct {
	Filename string
	Start    utils.TimeUtil
	End      utils.TimeUtil
}

type MissingDateRange struct {
	Start utils.TimeUtil
	End   utils.TimeUtil
}

type stocksFetcherIndexKey struct {
	Ticker   string
	Interval int
	Timespan model.Timespan
}

type fileIndex struct {
	Files map[stocksFetcherIndexKey][]dateRange
}

type StocksFetcherData struct {
	index fileIndex
}

func (key stocksFetcherIndexKey) String() string {
	return fmt.Sprintf("%s_%d_%s", key.Ticker, key.Interval, key.Timespan)
}

func parseIndexKey(keyStr string) (stocksFetcherIndexKey, error) {
	parts := strings.SplitN(keyStr, "_", 3)
	if len(parts) != 3 {
		return stocksFetcherIndexKey{}, fmt.Errorf("invalid key format: %s", keyStr)
	}
	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return stocksFetcherIndexKey{}, fmt.Errorf("invalid interval: %s", parts[1])
	}
	return stocksFetcherIndexKey{
		Ticker:   parts[0],
		Interval: interval,
		Timespan: model.Timespan(parts[2]),
	}, nil
}

func (fi *fileIndex) MarshalJSON() ([]byte, error) {
	type Alias fileIndex
	aux := make(map[string][]dateRange)
	for k, v := range fi.Files {
		aux[k.String()] = v
	}
	return json.Marshal(&struct {
		Files map[string][]dateRange `json:"files"`
	}{
		Files: aux,
	})
}

func (fi *fileIndex) UnmarshalJSON(data []byte) error {
	aux := struct {
		Files map[string][]dateRange `json:"files"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	fi.Files = make(map[stocksFetcherIndexKey][]dateRange)
	for k, v := range aux.Files {
		key, err := parseIndexKey(k)
		if err != nil {
			return err
		}
		fi.Files[key] = v
	}
	return nil
}

func NewStocksFetcherData() *StocksFetcherData {
	log.Println("Initializing StocksFetcherData")
	df := &StocksFetcherData{
		index: fileIndex{
			Files: make(map[stocksFetcherIndexKey][]dateRange),
		},
	}
	df.loadIndex()
	return df
}

func (df *StocksFetcherData) loadIndex() {
	log.Println("Loading index from file")
	file, err := os.Open(STOCKS_FETCHER_DATA_INDEX_PATH)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Index file does not exist, creating a new one")
			return
		}
		log.Printf("Error loading index: %v\n", err)
		return
	}
	defer func() {
		log.Println("Closing index file")
		file.Close()
	}()

	if err := json.NewDecoder(file).Decode(&df.index); err != nil {
		log.Printf("Error decoding index: %v\n", err)
	}
}

func (df *StocksFetcherData) saveIndex() {
	log.Println("Saving index to file")
	file, err := os.Create(STOCKS_FETCHER_DATA_INDEX_PATH)
	if err != nil {
		log.Printf("Error saving index: %v\n", err)
		return
	}
	defer func() {
		log.Println("Closing index file")
		file.Close()
	}()

	if err := json.NewEncoder(file).Encode(&df.index); err != nil {
		log.Printf("Error encoding index: %v\n", err)
	}
}

func (df *StocksFetcherData) updateIndex(key stocksFetcherIndexKey, dr dateRange) {
	log.Printf("Updating index with new date range for key: %+v\n", key)

	// Filter out dateRanges that fall within dr.Start and dr.End
	updatedRanges := []dateRange{dr}
	for _, existingDR := range df.index.Files[key] {
		switch {
		case existingDR.Start.After(dr.End):
			fallthrough
		//case existingDR.Start.Equal(dr.End):
		//	fallthrough
		case existingDR.End.Before(dr.Start):
			//fallthrough
			//case existingDR.End.Equal(dr.Start):
			updatedRanges = append(updatedRanges, existingDR)
		default:
			log.Printf("Removing dateRange from index: %+v\n", existingDR)
		}
	}

	df.index.Files[key] = updatedRanges
	df.saveIndex()
}

func (df *StocksFetcherData) saveDataToCSVFile(data *model.PolygonResponse, filename string) error {
	log.Printf("Ensuring directory exists: %s\n", STOCKS_FETCHER_DATA_PATH)
	if err := fileutils.EnsureDir(STOCKS_FETCHER_DATA_PATH); err != nil {
		return fmt.Errorf("error ensuring data directory exists: %v", err)
	}

	log.Printf("Creating CSV file: %s\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer func() {
		log.Println("Closing CSV file")
		file.Close()
	}()

	writer := csv.NewWriter(file)
	defer func() {
		log.Println("Flushing CSV writer")
		writer.Flush()
	}()

	header := []string{"Time", "Open", "High", "Low", "Close", "Volume"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	for _, result := range data.Results {
		record := []string{
			time.Unix(result.Time/1000, 0).Format(time.RFC3339),
			fmt.Sprintf("%f", result.Open),
			fmt.Sprintf("%f", result.High),
			fmt.Sprintf("%f", result.Low),
			fmt.Sprintf("%f", result.Close),
			fmt.Sprintf("%f", result.Volume),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing CSV record: %v", err)
		}
	}
	log.Printf("Data saved to CSV file: %s\n", filename)
	return nil
}

func (df *StocksFetcherData) SaveDataForRequest(request *model.HistoricalDataRequest, data *model.PolygonResponse) error {
	log.Printf("Saving data for request: %+v\n", request)
	newKey, newDr := df.getIndexKeyAndDataRange(request, data)

	for _, dr := range df.index.Files[newKey] {
		if df.fileCoversRange(dr, request.StartDate, request.EndDate) {
			log.Printf("Deleting file covering range: %s\n", dr.Filename)
			if err := os.Remove(dr.Filename); err != nil {
				log.Printf("Error deleting file: %v\n", err)
			}
		}
	}

	err := df.saveDataToCSVFile(data, newDr.Filename)
	if err != nil {
		log.Printf("Error saving data to CSV file: %v\n", err)
		return err
	}
	df.updateIndex(newKey, newDr)
	log.Println("Data successfully saved and index updated")
	return nil
}

func (df *StocksFetcherData) ReadDataForRequest(request *model.HistoricalDataRequest) (*model.PolygonResponse, error) {
	log.Printf("Reading data for request: %+v\n", request)
	newKey := df.getIndexKey(request)
	var combinedData *model.PolygonResponse
	for _, dr := range df.index.Files[newKey] {
		if df.fileCoversRange(dr, request.StartDate, request.EndDate) {
			log.Printf("Reading data from file: %s\n", dr.Filename)
			data, err := df.readDataFromCSV(dr.Filename)
			if err != nil {
				log.Printf("Error reading data from file: %v\n", err)
				return nil, err
			}
			if combinedData == nil {
				combinedData = data
			} else {
				combinedData.MergeResponse(data)
			}
		}
	}
	log.Println("Data successfully read and combined")
	return combinedData, nil
}

func (df *StocksFetcherData) getIndexKeyAndDataRange(request *model.HistoricalDataRequest, data *model.PolygonResponse) (stocksFetcherIndexKey, dateRange) {
	// After saving the new data
	actualStartDate := utils.TimeFromTimeStamp(data.Results[0].Time)
	actualEndDate := utils.TimeFromTimeStamp(data.Results[len(data.Results)-1].Time)
	filename := fmt.Sprintf("%s/%s_%s_to_%s_%d_%s.csv", STOCKS_FETCHER_DATA_PATH, request.Ticker, actualStartDate.StockFormatDate(), actualEndDate.StockFormatDate(), request.Interval, request.Timespan)
	dr := dateRange{Start: actualStartDate, End: actualEndDate, Filename: filename}
	log.Printf("Generated filename: %s, start date: %s, end date: %s\n", filename, actualStartDate, actualEndDate)

	return df.getIndexKey(request), dr
}

func (df *StocksFetcherData) getIndexKey(request *model.HistoricalDataRequest) stocksFetcherIndexKey {
	log.Printf("Generating index key for request: %+v\n", request)
	return stocksFetcherIndexKey{Ticker: request.Ticker, Interval: request.Interval, Timespan: request.Timespan}
}

func (df *StocksFetcherData) fileCoversRange(dr dateRange, fileStart, fileEnd utils.TimeUtil) bool {
	covers := !(dr.End.Before(fileStart) || dr.Start.After(fileEnd))
	log.Printf("Checking if file %s covers range %s to %s: %v\n", dr.Filename, fileStart, fileEnd, covers)
	return covers
}

func (df *StocksFetcherData) readDataFromCSV(filename string) (*model.PolygonResponse, error) {
	log.Printf("Opening CSV file for reading: %s\n", filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return nil, err
	}
	defer func() {
		log.Println("Closing CSV file")
		file.Close()
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v\n", err)
		return nil, err
	}

	var polygonResponse model.PolygonResponse
	for _, record := range records[1:] { // Skip header
		timestamp, err := time.Parse(time.RFC3339, record[0])
		if err != nil {
			log.Printf("Error parsing timestamp: %v\n", err)
			return nil, err
		}
		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Error parsing open price: %v\n", err)
			return nil, err
		}
		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Error parsing high price: %v\n", err)
			return nil, err
		}
		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Error parsing low price: %v\n", err)
			return nil, err
		}
		closePrice, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, err
		}
		volume, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return nil, err
		}

		polygonResponse.Results = append(polygonResponse.Results, model.DataPoint{
			Time:   timestamp.Unix() * 1000,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}

	return &polygonResponse, nil
}

func (df *StocksFetcherData) GetMissingDateRanges(request *model.HistoricalDataRequest) []MissingDateRange {
	log.Printf("Getting missing date ranges for request: %+v\n", request)
	coveredDates := make(map[string]bool)
	newKey := df.getIndexKey(request)

	log.Printf("Checking covered dates for key: %+v\n", newKey)
	for _, dr := range df.index.Files[newKey] {
		for date := dr.Start; !date.After(dr.End); date = date.AddDate(0, 0, 1) {
			if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday {
				coveredDates[date.StockFormatDate()] = true
			}
		}
	}

	var missingRanges []MissingDateRange
	var currentMissingRange *MissingDateRange
	log.Println("Identifying missing date ranges")
	for date := request.StartDate; !date.After(request.EndDate); date = date.AddDate(0, 0, 1) {
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue // Skip weekends
		}

		if !coveredDates[date.StockFormatDate()] {
			if currentMissingRange == nil {
				log.Printf("New missing range started at %s\n", date.StockFormatDate())
				currentMissingRange = &MissingDateRange{Start: date, End: date}
			} else {
				currentMissingRange.End = date
			}
		} else {
			if currentMissingRange != nil {
				log.Printf("Missing range identified: %s to %s\n", currentMissingRange.Start.StockFormatDate(), currentMissingRange.End.StockFormatDate())
				missingRanges = append(missingRanges, *currentMissingRange)
				currentMissingRange = nil
			}
		}
	}

	if currentMissingRange != nil {
		log.Printf("Final missing range identified: %s to %s\n", currentMissingRange.Start.StockFormatDate(), currentMissingRange.End.StockFormatDate())
		missingRanges = append(missingRanges, *currentMissingRange)
	}
	log.Println("Completed identification of missing date ranges")
	return missingRanges
}

// {"files":{"AAPL_1_day":[{"Filename":"data/stocks_fetcher/AAPL_2023-06-14_to_2023-06-16_1_day.csv","Start":"2023-06-14T09:30:00+05:30","End":"2023-06-16T09:30:00+05:30"},{"Filename":"data/stocks_fetcher/AAPL_2023-06-06_to_2023-06-12_1_day.csv","Start":"2023-06-06T09:30:00+05:30","End":"2023-06-12T09:30:00+05:30"}]}}
