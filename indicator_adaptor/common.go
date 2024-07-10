// common.go
package indicator_adaptor

import (
	"context"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
	"github.com/vd09/trading-algorithm-backtesting-system/monitor"
)

const (
	PERIOD_LABEL       = "period"
	ADAPTOR_NAME_LABEL = "adaptor_name"
)

func getUpdatedCommonLabelsContext(ctx context.Context) context.Context {
	slice, ok := ctx.Value(constraint.COMMON_LABELS_CTX).(monitor.Labels)
	if !ok {
		slice = monitor.Labels{}
	}
	slice = append(slice, ADAPTOR_NAME_LABEL)
	return context.WithValue(ctx, constraint.COMMON_LABELS_CTX, slice)
}
