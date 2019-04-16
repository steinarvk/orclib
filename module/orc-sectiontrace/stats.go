package orcsectiontrace

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	orcprometheus "github.com/steinarvk/orclib/module/orc-prometheus"
)

var (
	metricNodesGenerated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "nodes_generated",
		Help:      "Number of section trace nodes generated",
	})

	metricExecutionsOfSectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "section_executions",
		Help:      "Time (in seconds) spent inside section",
	}, []string{"section", "ok"})

	metricTimeSpentInSectionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "section_total_time",
		Help:      "Time (in seconds) spent inside section",
	}, []string{"section", "ok"})

	metricTimeSpentInSectionHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "trace",
		Name:      "section_execution_time_histogram",
		Help:      "Time (in seconds) spent inside section for each execution",
		Buckets:   orcprometheus.DefTimeBuckets,
	}, []string{"section", "ok"})

	metricTotalTimeInAnySection = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "any_section_total_time",
		Help:      "Time (in seconds) spent inside any section, including overhead",
	})

	metricOverheadTimeInAnySection = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "any_section_overhead_time",
		Help:      "Time (in seconds) spent processing section tracing inside any section",
	})

	metricTraceRecordsReceivedByCollector = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "records_collector_received",
		Help:      "Number of records received by the collector",
	}, []string{"phase", "fate"})

	metricTraceRecordsReceivedByProcessor = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "trace",
		Name:      "records_processor_received",
		Help:      "Number of records received by the collector",
	}, []string{"phase", "collectors"})
)
