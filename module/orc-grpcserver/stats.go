package orcgrpcserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRequestsInspected = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grpc",
		Name:      "hijacker_requests_inspected",
		Help:      "Number of gRPC requests inspected for hijacking (by whether they were routed to gRPC or not)",
	},
		[]string{"grpc"},
	)
)
