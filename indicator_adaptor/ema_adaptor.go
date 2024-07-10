package indicator_adaptor

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
	"github.com/vd09/trading-algorithm-backtesting-system/utils"
	"go.uber.org/zap"
)

type EMAMetrics struct {
	SignalCounter monitor.CounterMetric
	ValueCounter  monitor.CounterMetric
	monitor       monitor.Monitoring
}

type EMAAdapter struct {
	EMAs                   map[int]*indicator.EMA
	CurrentData            model.DataPoint
	MaxTotalHistoricalData int
	HistoricalValues       map[int][]float64
	periods                []int
	periodsString          string // Store periods as string
	logger                 logger.LoggerInterface
	metrics                *EMAMetrics
}

func NewEMAAdapter(ctx context.Context, periods []int, maxTotalHistoricalData int, monitor monitor.Monitoring) *EMAAdapter {
	emas := make(map[int]*indicator.EMA)
	historicalValues := make(map[int][]float64)
	for _, period := range periods {
		emas[period] = indicator.NewEMA(period)
		historicalValues[period] = []float64{}
	}
	sort.Ints(periods)

	emaAdapter := &EMAAdapter{
		EMAs:                   emas,
		HistoricalValues:       historicalValues,
		MaxTotalHistoricalData: maxTotalHistoricalData,
		periods:                periods,
		periodsString:          getPeriodsString(periods), // Store periods as string
		logger:                 logger.GetLogger(),
	}
	emaAdapter.registerMetrics(ctx, monitor)
	return emaAdapter
}

func (ea *EMAAdapter) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = ea.getUpdateContext(ctx)
	ea.metrics = &EMAMetrics{
		monitor:       m,
		SignalCounter: m.RegisterCounter(ctx, "ema_signals_generated", "Total number of signals generated", []string{constraint.SIGNAL_TYPE_LABEL, ADAPTOR_NAME_LABEL}),
		ValueCounter:  m.RegisterCounter(ctx, "ema_value", "EMA line", []string{PERIOD_LABEL, ADAPTOR_NAME_LABEL}),
	}
}

func (ea *EMAAdapter) Clone(ctx context.Context) IndicatorAdaptor {
	return NewEMAAdapter(ctx, ea.periods, ea.MaxTotalHistoricalData, ea.metrics.monitor)
}

func (ea *EMAAdapter) AddDataPoint(ctx context.Context, data model.DataPoint) error {
	ctx = ea.getUpdateContext(ctx)
	ea.logger.Debug(ctx, "Adding data point to EMAAdapter", zap.Int64("timestamp", data.Time))

	ea.CurrentData = data
	for period, ema := range ea.EMAs {
		ema.AddDataPoint(ctx, data)
		if ema.Initialized {
			tags := monitor.NewTagsKV(PERIOD_LABEL, period)
			tags.Add(ADAPTOR_NAME_LABEL, ea.Name())
			ea.metrics.ValueCounter.SetValue(ctx, ema.Value, tags)

			ea.HistoricalValues[period] = append(ea.HistoricalValues[period], ema.Value)
		}
		if len(ea.HistoricalValues[period]) > ea.MaxTotalHistoricalData {
			ea.HistoricalValues[period] = ea.HistoricalValues[period][1:]
		}
	}

	return nil
}

func (ea *EMAAdapter) Name() string {
	return fmt.Sprintf("EMA_%s", ea.periodsString) // Use the stored periods string
}

func (ea *EMAAdapter) GetSignal(ctx context.Context) (result model.StockAction) {
	ctx = ea.getUpdateContext(ctx)
	defer func() {
		tags := monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, string(result))
		tags.Add(ADAPTOR_NAME_LABEL, ea.Name())
		ea.metrics.SignalCounter.IncrementCounter(ctx, tags)
	}()

	zaps := []zap.Field{zap.String("adapter", ea.Name()), zap.Any("time", ea.CurrentData.Time)}
	for _, ema := range ea.EMAs {
		if !ema.Initialized {
			ea.logger.Debug(ctx, "EMA not initialized", zaps...)
			return model.Wait
		}
	}

	highestPeriod := ea.periods[len(ea.periods)-1]
	if len(ea.HistoricalValues[highestPeriod]) < 2 {
		ea.logger.Debug(ctx, "Not enough historical data for signal generation", zaps...)
		return model.Wait
	}

	if ea.checkForBuySignal() {
		ea.logger.Info(ctx, "Buy signal detected", zaps...)
		return model.Buy
	} else if ea.checkForSellSignal() {
		ea.logger.Info(ctx, "Sell signal detected", zaps...)
		return model.Sell
	}

	ea.logger.Debug(ctx, "No trading signal detected", zaps...)
	return model.Wait
}

func getPeriodsString(periods []int) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(periods)), "_"), "[]")
}

func (ea *EMAAdapter) checkForBuySignal() bool {
	smallestPeriod := ea.periods[0]
	for _, period := range ea.periods[1:] {
		if ea.HistoricalValues[smallestPeriod][0] >= ea.HistoricalValues[period][0] {
			return false
		}
		if !utils.IsLineIntersect(ea.HistoricalValues[smallestPeriod], ea.HistoricalValues[period]) {
			return false
		}
	}
	return true
}

func (ea *EMAAdapter) checkForSellSignal() bool {
	smallestPeriod := ea.periods[0]
	for _, period := range ea.periods[1:] {
		if ea.HistoricalValues[smallestPeriod][0] <= ea.HistoricalValues[period][0] {
			return false
		}
		if !utils.IsLineIntersect(ea.HistoricalValues[smallestPeriod], ea.HistoricalValues[period]) {
			return false
		}
	}
	return true
}

// Function to retrieve and update the slice from context
func (ea *EMAAdapter) getUpdateContext(ctx context.Context) context.Context {
	ctx = getUpdatedCommonLabelsContext(ctx)
	return context.WithValue(ctx, ADAPTOR_NAME_LABEL, ea.Name())
}
