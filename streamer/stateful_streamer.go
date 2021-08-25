package streamer

import (
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promProduceCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sstreamer_received_msgs_total",
			Help: "The total number of received messages.",
		},
		[]string{"channel"},
	)
	promConsumeCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sstreamer_sent_msgs_total",
			Help: "The total number of sent messages.",
		},
		[]string{"channel"},
	)
)

type StatefulStreamer struct {
	producers []StreamProducer
	consumers []StreamConsumer
	buffers   map[string]streamBuffer
}

type StreamProducer interface {
	GetName() string
	GetOutputs() []string
	GetBufferLengths() []int
}

type StreamConsumer interface {
	GetName() string
	GetInputs() []string
	GetStreamController() StreamController
}

func NewStatefulStreamer() *StatefulStreamer {
	return &StatefulStreamer{
		buffers: make(map[string]streamBuffer),
	}
}

func (ss *StatefulStreamer) RegisterProducer(p StreamProducer) {
	ss.producers = append(ss.producers, p)
}

func (ss *StatefulStreamer) RegisterConsumer(c StreamConsumer) {
	ss.consumers = append(ss.consumers, c)
}

func (ss *StatefulStreamer) Send(name string, data interface{}) {
	ss.buffers[name].Add(data)
}

func (ss *StatefulStreamer) Serve() {
	for _, p := range ss.producers {
		outputs := p.GetOutputs()
		lengths := p.GetBufferLengths()
		for i := range outputs {
			if lengths[i] != 0 {
				ss.buffers[outputs[i]] = newDefaultStreamBuffer(outputs[i], lengths[i])
			} else {
				ss.buffers[outputs[i]] = newStaticStreamBuffer(outputs[i])
			}
		}
	}

	for _, c := range ss.consumers {
		inputs := c.GetInputs()
		for i := range inputs {
			if inputs[i][0] == '#' {
				c.GetStreamController().GetListener(inputs[i])
				continue
			}
			ss.buffers[inputs[i]].RegisterListener(c.GetStreamController().GetListener(inputs[i]))
		}
		go c.GetStreamController().Serve()
	}

	for _, buf := range ss.buffers {
		go buf.Serve()
	}
}

func (ss *StatefulStreamer) GenerateDot() string {
	str := "digraph g {\n"
	str += "\tStreamer [shape=record]\n"
	for _, p := range ss.producers {
		outputs := p.GetOutputs()
		for _, o := range outputs {
			str += "\t" + p.GetName() + " -> Streamer [label=\"" + o + "\", shape=record]\n"
			str += "\t" + p.GetName() + " [shape=record]\n"
		}
	}
	intermediate := make(map[string]struct{})
	for _, c := range ss.consumers {
		inputs := c.GetInputs()
		for _, i := range inputs {
			if i[0] == '#' {
				intermediate[string(i[1:])] = struct{}{}
				continue
			}
			str += "\tStreamer -> " + c.GetName() + " [label=\"" + i + "\"]\n"
		}
		worker := c.GetStreamController().GetActor()
		switch w := worker.(type) {
		case *ShortCircuitActor:
			str += "\t" + c.GetName() + " -> " + w.consumer.GetName() + " [label=\"" + w.name + "\"]\n"
		default:
			str += "\t" + c.GetName() + " -> " + reflect.TypeOf(w).Elem().Name() + " [shape=circle]\n"
		}
	}
	rank := "\t{ rank=same "
	for _, c := range ss.consumers {
		str += "\t" + c.GetName() + " [label=<<b>" + c.GetName() + "</b><br/>"
		str += c.GetStreamController().GenerateDotLabel() + ">,"
		if _, ok := intermediate[c.GetName()]; ok {
			str += " shape=record style=rounded color=grey bgcolor=grey"
		} else {
			str += " shape=record"
			rank += c.GetName() + " "
		}
		str += "]\n"
	}
	str += rank + "}\n"
	str += "}"
	return str
}
