package jsonapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	orcprometheus "github.com/steinarvk/orclib/module/orc-prometheus"
)

var (
	metricRequestsBegun = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "jsonapi",
		Name:      "api_requests_begun",
		Help:      "Number of API requests for which processing was started.",
	},
		[]string{"endpoint", "method"},
	)

	metricRequestsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "jsonapi",
		Name:      "api_requests_processed",
		Help:      "Number of API requests for which processing finished.",
	},
		[]string{"endpoint", "method", "code"},
	)

	metricRequestsProcessedLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "jsonapi",
		Name:      "api_request_processing_latency_histogram",
		Help:      "Latency between starting and finishing processing of a request.",
		Buckets:   orcprometheus.DefTimeBuckets,
	},
		[]string{"endpoint", "method", "code"},
	)
)

func wrapForStats(endpoint string, apihandler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		fields := logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}
		labels := prometheus.Labels{
			"endpoint": endpoint,
			"method":   r.Method,
		}
		logrus.WithFields(fields).Infof("Received API request")
		metricRequestsBegun.With(labels).Inc()

		err := apihandler(w, r)
		code := getErrorCode(err)

		duration := time.Since(t0)

		fields["duration"] = duration
		fields["code"] = fmt.Sprintf("%d", code)

		labels["code"] = fmt.Sprintf("%d", code)

		metricRequestsProcessed.With(labels).Inc()
		metricRequestsProcessedLatencyHistogram.With(labels).Observe(duration.Seconds())

		if err != nil {
			logrus.WithFields(fields).Infof("Request finished with error: %v", err)
		} else {
			logrus.WithFields(fields).Infof("Request succeeded")
		}
	})
}
