package datafetcher

import (
	"sort"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

func mergeData(existing, new *model.PolygonResponse) *model.PolygonResponse {
	combined := make(map[int64]model.DataPoint)

	for _, result := range existing.Results {
		combined[result.Time] = result
	}
	for _, result := range new.Results {
		combined[result.Time] = result
	}

	var mergedResults []model.DataPoint
	for _, result := range combined {
		mergedResults = append(mergedResults, result)
	}

	sort.Slice(mergedResults, func(i, j int) bool {
		return mergedResults[i].Time < mergedResults[j].Time
	})

	return &model.PolygonResponse{
		Ticker:  existing.Ticker,
		Results: mergedResults,
	}
}

func getMissingDateRanges(request model.HistoricalDataRequest, existingData *model.PolygonResponse) []dateRange {
	requestStart, _ := time.Parse("2006-01-02", request.StartDate)
	requestEnd, _ := time.Parse("2006-01-02", request.EndDate)

	existingDates := make(map[string]bool)
	for _, result := range existingData.Results {
		date := time.Unix(result.Time/1000, 0).Format("2006-01-02")
		existingDates[date] = true
	}

	var missingRanges []dateRange
	currentStart := ""
	currentEnd := ""

	for date := requestStart; !date.After(requestEnd); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")
		if !existingDates[dateStr] {
			if currentStart == "" {
				currentStart = dateStr
			}
			currentEnd = dateStr
		} else {
			if currentStart != "" {
				missingRanges = append(missingRanges, dateRange{currentStart, currentEnd})
				currentStart = ""
				currentEnd = ""
			}
		}
	}

	if currentStart != "" {
		missingRanges = append(missingRanges, dateRange{currentStart, currentEnd})
	}

	return missingRanges
}

func nextDate(date string) string {
	parsedDate, _ := time.Parse("2006-01-02", date)
	return parsedDate.AddDate(0, 0, 1).Format("2006-01-02")
}
