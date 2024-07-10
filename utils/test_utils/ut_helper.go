package test_utils

import (
	"testing"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/mocks/mock_monitor"
	"go.uber.org/mock/gomock"
)

func NewMockMetricsCollector(t *testing.T) *mock_monitor.MockMonitoring {
	config.InitConfig()
	ctl := gomock.NewController(t)
	metricsCollector := mock_monitor.NewMockMonitoring(ctl)

	counter := mock_monitor.NewMockCounterMetric(ctl)
	counter.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	counter.EXPECT().SetCounter(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	metricsCollector.EXPECT().RegisterCounter(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(counter).AnyTimes()

	gauge := mock_monitor.NewMockGaugeMetric(ctl)
	gauge.EXPECT().SetGauge(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	metricsCollector.EXPECT().RegisterGauge(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(gauge).AnyTimes()

	histogram := mock_monitor.NewMockHistogramMetrics(ctl)
	histogram.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	metricsCollector.EXPECT().RegisterHistogram(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(histogram).AnyTimes()

	return metricsCollector
}
