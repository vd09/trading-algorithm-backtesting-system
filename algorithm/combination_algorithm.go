package algorithm

import (
	"context"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/indicator_adaptor"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"github.com/vd09/trading-algorithm-backtesting-system/model"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
)

const (
	ALGORITHM_NAME_LABEL = "algorithm_name"
)

type CombinationTradingAlgorithm struct {
	adaptors []indicator_adaptor.IndicatorAdaptor
	logger   logger.LoggerInterface
	metrics  *TradingAlgorithmMetrics
}

type TradingAlgorithmMetrics struct {
	SignalCounter monitor.CounterMetric
	ClosePrice    monitor.CounterMetric
	monitor       monitor.Monitoring
}

func NewCombinationTradingAlgorithm(ctx context.Context, adaptors []indicator_adaptor.IndicatorAdaptor, monitor monitor.Monitoring) *CombinationTradingAlgorithm {
	algo := &CombinationTradingAlgorithm{
		adaptors: adaptors,
		logger:   logger.GetLogger(),
	}
	algo.registerMetrics(ctx, monitor)
	return algo
}

func (ta *CombinationTradingAlgorithm) registerMetrics(ctx context.Context, m monitor.Monitoring) {
	ctx = ta.getUpdateContext(ctx)
	ta.metrics = &TradingAlgorithmMetrics{
		monitor:       m,
		ClosePrice:    m.RegisterCounter(ctx, "algorithm_close_price", "Algorithm close price", nil),
		SignalCounter: m.RegisterCounter(ctx, "algorithm_signals_generated", "Total number of signals generated", []string{constraint.SIGNAL_TYPE_LABEL}),
	}
}

// Name returns the name of the trading algorithm
func (ta *CombinationTradingAlgorithm) Name() string {
	name := ""
	for _, adaptor := range ta.adaptors {
		name += adaptor.Name() + "_"
	}
	return name[:len(name)-1] // Remove the trailing underscore
}

// Evaluate evaluates the signals from all adaptors and determines the action
func (ta *CombinationTradingAlgorithm) Evaluate(ctx context.Context, data model.DataPoint) (result model.TradingSignal) {
	ctx = ta.getUpdateContext(ctx)
	defer func() {
		ta.metrics.SignalCounter.IncrementCounter(ctx, monitor.NewTagsKV(constraint.SIGNAL_TYPE_LABEL, model.StockAction(result.Action)))
	}()
	ta.metrics.ClosePrice.SetValue(ctx, data.Close, nil)

	buyCount, sellCount := 0, 0
	for _, adaptor := range ta.adaptors {
		adaptor.AddDataPoint(ctx, data)
		signal := adaptor.GetSignal(ctx)
		switch signal {
		case model.Buy:
			buyCount++
		case model.Sell:
			sellCount++
		case model.Wait:
			return model.TradingSignal{Time: data.Time, Action: model.Wait}
		}
	}

	if buyCount == len(ta.adaptors) {
		return model.TradingSignal{Time: data.Time, Action: model.Buy}
	} else if sellCount == len(ta.adaptors) {
		return model.TradingSignal{Time: data.Time, Action: model.Sell}
	} else {
		return model.TradingSignal{Time: data.Time, Action: model.Wait}
	}
}

// Function to retrieve and update the slice from context
func (ta *CombinationTradingAlgorithm) getUpdateContext(ctx context.Context) context.Context {
	return getUpdatedCommonLabelsContext(ctx, ta.Name())
}

func CreateCombinationTradingAlgorithms(rootCtx context.Context, adaptors []indicator_adaptor.IndicatorAdaptor) []TradingAlgorithm {
	var tradingAlgorithms []TradingAlgorithm
	var generate func(context.Context, []indicator_adaptor.IndicatorAdaptor, int)
	promMetric := monitor.NewPrometheusMonitoring()

	generate = func(ctx context.Context, currentCombination []indicator_adaptor.IndicatorAdaptor, start int) {
		if len(currentCombination) > 0 {
			combo := append([]indicator_adaptor.IndicatorAdaptor{}, currentCombination...)
			clonedCombo := make([]indicator_adaptor.IndicatorAdaptor, len(combo))
			for i, adaptor := range combo {
				clonedCombo[i] = adaptor.Clone(getUpdatedCommonLabelsContext(ctx, ""))
			}
			tradingAlgorithms = append(tradingAlgorithms, NewCombinationTradingAlgorithm(ctx, clonedCombo, promMetric))
		}
		for i := start; i < len(adaptors); i++ {
			generate(ctx, append(currentCombination, adaptors[i]), i+1)
		}
	}

	generate(rootCtx, []indicator_adaptor.IndicatorAdaptor{}, 0)
	return tradingAlgorithms
}

func getUpdatedCommonLabelsContext(ctx context.Context, name string) context.Context {
	slice, ok := ctx.Value(constraint.COMMON_LABELS_CTX).(monitor.Labels)
	if !ok {
		slice = monitor.Labels{}
	}
	slice = append(slice, ALGORITHM_NAME_LABEL)
	ctx = context.WithValue(ctx, constraint.COMMON_LABELS_CTX, slice)
	return context.WithValue(ctx, ALGORITHM_NAME_LABEL, name)
}
