package indicator

import (
	"context"
	"errors"

	"github.com/vd09/trading-algorithm-backtesting-system/model"
)

// MACD represents the state of the MACD indicator.
type MACD struct {
	ShortEMA     float64
	LongEMA      float64
	SignalLine   float64
	History      []model.DataPoint
	ShortPeriod  int
	LongPeriod   int
	SignalPeriod int
	Initialized  bool
}

// MACDResult holds the results of the MACD calculation.
type MACDResult struct {
	MACDLine      float64
	MACDHistogram float64
	MACDSignal    float64
}

// NewMACD initializes a new MACD instance.
func NewMACD(shortPeriod, longPeriod, signalPeriod int) *MACD {
	return &MACD{
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		SignalPeriod: signalPeriod,
		Initialized:  false,
	}
}

// AddDataPoint adds a new data point and updates the MACD calculation.
func (m *MACD) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	if len(m.History) > 0 && data.Time <= m.History[len(m.History)-1].Time {
		return errors.New("data point is not in chronological order")
	}

	m.History = append(m.History, data)
	if len(m.History) > m.LongPeriod {
		m.History = m.History[1:]
	} else if len(m.History) == m.LongPeriod {
		m.Initialized = true
	} else {
		return nil
	}

	m.calculateEMA()
	return nil
}

func (m *MACD) calculateEMA() {
	if len(m.History) < m.LongPeriod {
		return
	}

	if len(m.History) == m.LongPeriod {
		m.ShortEMA = m.simpleMovingAverage(m.ShortPeriod)
		m.LongEMA = m.simpleMovingAverage(m.LongPeriod)
	} else {
		kShort := 2.0 / (float64(m.ShortPeriod) + 1)
		kLong := 2.0 / (float64(m.LongPeriod) + 1)
		m.ShortEMA = ((m.History[len(m.History)-1].Close - m.ShortEMA) * kShort) + m.ShortEMA
		m.LongEMA = ((m.History[len(m.History)-1].Close - m.LongEMA) * kLong) + m.LongEMA
	}

	m.SignalLine = ((m.ShortEMA-m.LongEMA)-m.SignalLine)*(2.0/(float64(m.SignalPeriod)+1)) + m.SignalLine
}

func (m *MACD) simpleMovingAverage(period int) float64 {
	if len(m.History) < period {
		return 0.0
	}
	sum := 0.0
	for i := len(m.History) - period; i < len(m.History); i++ {
		sum += m.History[i].Close
	}
	return sum / float64(period)
}

// CalculateMACD returns the MACD line and the MACD histogram as an object.
func (m *MACD) CalculateMACD() MACDResult {
	macdLine := m.ShortEMA - m.LongEMA
	macdHistogram := macdLine - m.SignalLine
	return MACDResult{
		MACDLine:      macdLine,
		MACDHistogram: macdHistogram,
		MACDSignal:    m.SignalLine,
	}
}
