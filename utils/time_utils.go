package utils

import "time"

const (
	STOCK_DATE_FORMAT_LAYOUT = "2006-01-02"
	TIME_LAYOUT              = "2006-01-02T15:04:05Z07:00" // RFC3339 format
)

type TimeUtil struct {
	time.Time
}

func NewTimeUtil(t time.Time) TimeUtil {
	return TimeUtil{Time: t.UTC()}
}

func NewTimeUtilFromFormat(dateString string) (TimeUtil, error) {
	t, err := time.Parse(STOCK_DATE_FORMAT_LAYOUT, dateString)
	if err != nil {
		return TimeUtil{}, err
	}
	return NewTimeUtil(t), nil
}

func TimeFromTimeStamp(t int64) TimeUtil {
	// Convert the timestamp to time.Time
	parsedTime := time.Unix(t/1000, 0)
	// Truncate to the start of the day
	truncatedTime := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, time.UTC)
	// Return as TimeUtil
	return NewTimeUtil(truncatedTime)
}

func (tu TimeUtil) StockFormatDate() string {
	return tu.Time.Format(STOCK_DATE_FORMAT_LAYOUT)
}

func (tu TimeUtil) Before(u TimeUtil) bool {
	return tu.Time.Before(u.Time)
}

func (tu TimeUtil) After(u TimeUtil) bool {
	return tu.Time.After(u.Time)
}

func (tu TimeUtil) Equal(u TimeUtil) bool {
	return tu.Time.Equal(u.Time)
}

func (tu TimeUtil) AddDate(years int, months int, days int) TimeUtil {
	return NewTimeUtil(tu.Time.AddDate(years, months, days))
}

func (tu TimeUtil) Unix() int64 {
	return tu.Time.Unix() * 1000
}

//func (tu TimeUtil) MarshalJSON() ([]byte, error) {
//	stamp := tu.Time.Format(TIME_LAYOUT)
//	return []byte(`"` + stamp + `"`), nil
//}
//
//func (tu *TimeUtil) UnmarshalJSON(data []byte) error {
//	parsedTime, err := time.Parse(`"`+TIME_LAYOUT+`"`, string(data))
//	if err != nil {
//		return err
//	}
//	tu.Time = parsedTime
//	return nil
//}

func (tu TimeUtil) String() string {
	return tu.StockFormatDate()
}
