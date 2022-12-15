package txspool

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
)

type Metrics struct {
	Size         metrics.Gauge
	TxsSizeBytes metrics.Histogram
	FailedTxs    metrics.Counter
}

func TxsPoolMetrics() *Metrics {
	return &Metrics{
		Size:         discard.NewGauge(),
		TxsSizeBytes: discard.NewHistogram(),
		FailedTxs:    discard.NewCounter(),
	}
}
