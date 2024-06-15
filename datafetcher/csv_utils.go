package datafetcher

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func SaveDataToCSV(data *model.PolygonResponse, filename string) error {
	if err := ensureDataDir(); err != nil {
		return fmt.Errorf("error ensuring data directory exists: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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

	return nil
}

func readDataFromCSV(filename string) (*model.PolygonResponse, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var polygonResponse model.PolygonResponse
	for _, record := range records[1:] { // Skip header
		timestamp, err := time.Parse(time.RFC3339, record[0])
		if err != nil {
			return nil, err
		}
		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}
		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, err
		}
		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
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

		polygonResponse.Results = append(polygonResponse.Results, struct {
			Time   int64   `json:"t"`
			Open   float64 `json:"o"`
			High   float64 `json:"h"`
			Low    float64 `json:"l"`
			Close  float64 `json:"c"`
			Volume float64 `json:"v"`
		}{
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

func ensureDataDir() error {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		return os.Mkdir("data", os.ModePerm)
	}
	return nil
}
