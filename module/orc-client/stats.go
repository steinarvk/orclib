package orcclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	orcprometheus "github.com/steinarvk/orclib/module/orc-prometheus"
)

const (
	statsNamespace = "orcclient"
)

var (
	metricRequestsRejected = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: statsNamespace,
		Name:      "requests_rejected",
		Help:      "Number of attempted requests rejected by sanity-checks, by reason",
	},
		[]string{"client", "reason"},
	)

	metricRequestsBegun = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: statsNamespace,
		Name:      "requests_sent",
		Help:      "Number of requests sent",
	},
		[]string{"client", "method", "target"},
	)

	metricRequestsFinished = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: statsNamespace,
		Name:      "requests_finished",
		Help:      "Number of outgoing requests fully processed",
	},
		[]string{"client", "method", "target", "ok", "code"},
	)

	metricRequestsFinishedLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: statsNamespace,
		Name:      "request_latency_histogram",
		Help:      "Outgoing request latency",
		Buckets:   orcprometheus.DefTimeBuckets,
	},
		[]string{"client", "method", "target", "ok", "code"},
	)
)
