package streamer

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type streamBuffer interface {
	Add(newItem interface{})
	RegisterListener(l streamListener)
	Serve()
}

type defaultStreamBuffer struct {
	name      string
	buf       chan interface{}
	listeners []streamListener
}

func newDefaultStreamBuffer(name string, length int) *defaultStreamBuffer {
	if length < 1 {
		panic("length of defaultStreamBuffer should >= 1")
	}

	return &defaultStreamBuffer{
		name: name,
		buf:  make(chan interface{}, length),
	}
}

func (dsb *defaultStreamBuffer) Add(newItem interface{}) {
	dsb.buf <- newItem
	promProduceCounter.With(prometheus.Labels{"channel": dsb.name}).Inc()
}

func (dsb *defaultStreamBuffer) RegisterListener(l streamListener) {
	dsb.listeners = append(dsb.listeners, l)
}

func (dsb *defaultStreamBuffer) Serve() {
	for item := range dsb.buf {
		for _, l := range dsb.listeners {
			l.Notify(item)
		}
		promConsumeCounter.With(prometheus.Labels{"channel": dsb.name}).Inc()
	}
}

type Comparable interface {
	Equals(rhs Comparable) bool
}

type staticStreamBuffer struct {
	name      string
	buf       chan Comparable
	state     Comparable
	listeners []streamListener
}

func newStaticStreamBuffer(name string) *staticStreamBuffer {
	return &staticStreamBuffer{
		name: name,
		buf:  make(chan Comparable),
	}
}

func (ssb *staticStreamBuffer) Add(newItem interface{}) {
	c, ok := newItem.(Comparable)
	if !ok {
		panic("non comparable object got")
	}

	ssb.buf <- c
	promProduceCounter.With(prometheus.Labels{"channel": ssb.name}).Inc()
}

func (ssb *staticStreamBuffer) RegisterListener(l streamListener) {
	ssb.listeners = append(ssb.listeners, l)
}

func (ssb *staticStreamBuffer) Serve() {
	for item := range ssb.buf {
		promConsumeCounter.With(prometheus.Labels{"channel": ssb.name}).Inc()
		if ssb.state != nil && ssb.state.Equals(item) {
			log.WithFields(log.Fields{
				"channel": ssb.name,
			}).Warn("duplicate state got")
			continue
		}

		ssb.state = item
		for _, l := range ssb.listeners {
			l.Notify(item)
		}
	}
}
