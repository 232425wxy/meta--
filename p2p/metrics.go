package p2p

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
)

type Metrics struct {
	Peers                 metrics.Gauge
	PeerReceiveBytesTotal metrics.Counter
	PeerSendBytesTotal    metrics.Counter
	PeerPendingSendBytes  metrics.Gauge
	NumTxs                metrics.Gauge
}

func P2PMetrics() *Metrics {
	return &Metrics{
		Peers:                 discard.NewGauge(),
		PeerReceiveBytesTotal: discard.NewCounter(),
		PeerSendBytesTotal:    discard.NewCounter(),
		PeerPendingSendBytes:  discard.NewGauge(),
		NumTxs:                discard.NewGauge(),
	}
}
