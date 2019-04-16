package sectiontraceapi

import (
	"sync"

	"github.com/steinarvk/orclib/lib/circularbuffer"
	"github.com/steinarvk/sectiontrace"
)

type Collector struct {
	mu     sync.Mutex
	circle *circularbuffer.Circular
	buf    []*sectiontrace.Record
}

func (k *Collector) Collect(rec *sectiontrace.Record) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.buf[k.circle.AppendIndex()] = rec
}

func NewCollector(size int) *Collector {
	return &Collector{
		circle: circularbuffer.New(size),
		buf:    make([]*sectiontrace.Record, size),
	}
}

func (k *Collector) GetRecords() []*sectiontrace.Record {
	k.mu.Lock()
	defer k.mu.Unlock()

	a, b, c, d := k.circle.SliceIndices()
	return append(k.buf[a:b], k.buf[c:d]...)
}
