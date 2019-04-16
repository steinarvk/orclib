package orcsectiontrace

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orclib/module/orc-logging"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/sectiontrace"

	identity "github.com/steinarvk/orclib/module/orc-identity"
)

const (
	recordChannelCapacity = 1000
)

type Module struct {
	collectors   []func(*sectiontrace.Record)
	collectorsMu sync.Mutex
	recordChan   chan *sectiontrace.Record
}

var M = &Module{}

func (m *Module) ModuleName() string { return "SectionTracing" }

func (m *Module) AddCollector(collector func(rec *sectiontrace.Record)) {
	m.collectorsMu.Lock()
	defer m.collectorsMu.Unlock()

	m.collectors = append(m.collectors, collector)
}

func (m *Module) collect(rec *sectiontrace.Record) {
	labels := prometheus.Labels{
		"phase": string(rec.Phase),
	}
	select {
	case m.recordChan <- rec:
		labels["fate"] = "sent"
	default:
		labels["fate"] = "dropped"
	}
	metricTraceRecordsReceivedByCollector.With(labels).Inc()
}

func (m *Module) getCollectors() []func(*sectiontrace.Record) {
	m.collectorsMu.Lock()
	defer m.collectorsMu.Unlock()

	return m.collectors
}

func (m *Module) processor() {
	for rec := range m.recordChan {
		colls := m.getCollectors()
		labels := prometheus.Labels{
			"phase":      string(rec.Phase),
			"collectors": fmt.Sprintf("%d", len(colls)),
		}

		for _, coll := range colls {
			coll(rec)
		}
		metricTraceRecordsReceivedByProcessor.With(labels).Inc()
	}
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(logging.M)
	})

	hooks.OnStart(func() error {
		sectiontrace.DefaultScope = identity.Identity().ShortEphemeralID
		sectiontrace.DefaultOtherData["exporterIdentity"] = identity.Identity()

		// The debug mode checks are a good-enough tradeoff for the Orc server use case.
		// In particular, failing fast if there are duplicate sections.
		// We should always create sections at the outermost scope possible, which
		// should mean creating them all at startup.
		// Could expose a flag choice, but don't think it's worth it at the moment.
		sectiontrace.DebugMode = true

		m.recordChan = make(chan *sectiontrace.Record, recordChannelCapacity)
		// TODO channelstats

		go m.processor()

		sectiontrace.OnNodeGenerated = func() {
			metricNodesGenerated.Inc()
		}
		sectiontrace.OnPanic = func(err error) {
			logrus.Fatalf("panic from sectiontrace: %v", err)
		}
		sectiontrace.OnTimeSpent = func(overhead, internal time.Duration, hadParent bool) {
			if !hadParent {
				// Don't double-count.
				metricTotalTimeInAnySection.Add((overhead + internal).Seconds())
			}
			metricOverheadTimeInAnySection.Add(overhead.Seconds())
		}
		sectiontrace.OnBegin = func(begin *sectiontrace.Record) {
			m.collect(begin)
		}
		sectiontrace.OnEnd = func(begin, end *sectiontrace.Record) {
			m.collect(end)

			if begin.Name != end.Name {
				logrus.WithFields(logrus.Fields{
					"begin.name": begin.Name,
					"end.name":   end.Name,
				}).Errorf("Inconsistent fields from begin/end from sectiontrace")
				return
			}

			name := begin.Name

			duration := time.Duration(end.TimestampMicros-begin.TimestampMicros) * time.Microsecond

			labels := prometheus.Labels{
				"section": name,
			}

			if ok, present := end.Args["ok"]; present {
				labels["ok"] = fmt.Sprintf("%v", ok)
			} else {
				labels["ok"] = "unknown"
			}

			durationSecs := duration.Seconds()

			metricExecutionsOfSectionTotal.With(labels).Inc()
			metricTimeSpentInSectionTotal.With(labels).Add(durationSecs)
			metricTimeSpentInSectionHistogram.With(labels).Observe(durationSecs)
		}

		return nil
	})
}
