package datafetcher

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

type dateRange struct {
	Start string
	End   string
}

func getExistingDataFiles(ticker string) ([]string, error) {
	files, err := ioutil.ReadDir("data")
	if err != nil {
		return nil, err
	}

	var existingFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if match, _ := regexp.MatchString(fmt.Sprintf(`^%s_.*\.csv$`, ticker), file.Name()); match {
			existingFiles = append(existingFiles, "data/"+file.Name())
		}
	}
	return existingFiles, nil
}

func fileCoversRange(filename, startDate, endDate string, interval int, timespan model.Timespan) bool {
	re := regexp.MustCompile(`_(\d{4}-\d{2}-\d{2})_to_(\d{4}-\d{2}-\d{2})_(\d+)_(\w+)\.csv$`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) != 5 {
		return false
	}

	fileStartDate, fileEndDate := matches[1], matches[2]
	fileInterval, _ := strconv.Atoi(matches[3])
	fileTimespan := model.Timespan(matches[4])

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return false
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return false
	}

	fileStart, err := time.Parse("2006-01-02", fileStartDate)
	if err != nil {
		return false
	}

	fileEnd, err := time.Parse("2006-01-02", fileEndDate)
	if err != nil {
		return false
	}

	switch {
	case fileInterval != interval:
		return false
	case fileTimespan != timespan:
		return false
	case end.Before(fileStart):
		return false
	case start.After(fileEnd):
		return false
	}

	return true
}
