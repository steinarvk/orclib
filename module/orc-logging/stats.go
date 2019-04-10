package logging

import (
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricLogEntries = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "log",
		Name:      "entries",
		Help:      "Number of (logrus) log entries generated",
	}, []string{"level"})
)

type logStatHook struct{}

func (_ logStatHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (_ logStatHook) Fire(entry *logrus.Entry) error {
	metricLogEntries.With(prometheus.Labels{
		"level": entry.Level.String(),
	}).Inc()

	return nil
}
