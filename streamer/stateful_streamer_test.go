package streamer

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type mockState struct {
	v string
}

func (ms *mockState) Equals(rhs Comparable) bool {
	return ms.v == rhs.(*mockState).v
}

type mockActor1 struct {
	c chan interface{}
}

func (mw1 *mockActor1) Consume(data interface{}) {
	mw1.c <- data
}

func (mw1 *mockActor1) serve() {
	for v := range mw1.c {
		fmt.Printf("State1: %v, State2: %v\n", v.([]interface{})[0].(*mockState), v.([]interface{})[1].(*mockState))
	}
}

type mockActor2 struct {
	c chan interface{}
}

func (mw2 *mockActor2) Consume(data interface{}) {
	mw2.c <- data
}

func (mw2 *mockActor2) serve() {
	for v := range mw2.c {
		fmt.Printf("%v: %v\n", v.(Aggregated).Name, v.(Aggregated).Data)
	}
}

func TestStateAggregator(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	ss := NewStatefulStreamer()
	p1 := NewDefaultProducer("A", []string{"State1"}, []int{0})
	p2 := NewDefaultProducer("B", []string{"State2"}, []int{0})
	actor := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"State1", "State2"}, NewConjunctions(&actor))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterConsumer(c)
	ss.Serve()
	go actor.serve()

	ss.Send("State1", &mockState{"1"})
	ss.Send("State2", &mockState{"2"})
	time.Sleep(time.Second)

	ss.Send("State1", &mockState{"1"})
	ss.Send("State1", &mockState{"2"})
	ss.Send("State2", &mockState{"1"})
	time.Sleep(time.Second)
}

func TestStreamAggregator(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	ss := NewStatefulStreamer()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 := NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	actor := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewConjunctions(&actor))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterConsumer(c)
	ss.Serve()
	go actor.serve()

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)
}

func TestDisjunctions(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	ss := NewStatefulStreamer()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 := NewDefaultProducer("A", []string{"Stream2"}, []int{1})
	actor := mockActor2{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewDisjunctions(&actor, 3))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterConsumer(c)
	ss.Serve()
	fmt.Println(ss.GenerateDot())
	go actor.serve()

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"2"})
	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream1", &mockState{"1"})
	time.Sleep(time.Second)
}

func TestCustomizableController(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	ss := NewStatefulStreamer()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{5})
	p2 := NewDefaultProducer("B", []string{"Stream2"}, []int{5})
	actor := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewCustomizableController(
		&actor, func(inputs interface{}) bool {
			return inputs.([]interface{})[0] != nil && inputs.([]interface{})[1] != nil &&
				inputs.([]interface{})[0].(*mockState).Equals(inputs.([]interface{})[1].(*mockState))
		},
	))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterConsumer(c)
	ss.Serve()
	fmt.Println(ss.GenerateDot())
	go actor.serve()

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"1"})
	time.Sleep(time.Second)

	ss.Send("Stream1", &mockState{"2"})
	ss.Send("Stream2", &mockState{"2"})
	ss.Send("Stream2", &mockState{"2"})
	ss.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	ss.Send("Stream1", &mockState{"3"})
	ss.Send("Stream2", &mockState{"1"})
	ss.Send("Stream2", &mockState{"3"})
	time.Sleep(time.Second)
}

func TestGenerateDot(t *testing.T) {
	var p1, p2, p3 StreamProducer
	var c1, c2 StreamConsumer
	ss := NewStatefulStreamer()
	p1 = NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 = NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	p3 = NewDefaultProducer("C", []string{"Stream3"}, []int{1})

	actor2 := mockActor2{
		c: make(chan interface{}),
	}
	c2 = NewDefaultConsumer("E", []string{"#D", "Stream3"}, NewDisjunctions(&actor2, 3))
	actor1 := NewShortCircuitActor(c2, "#D")
	c1 = NewDefaultConsumer("D", []string{"Stream1", "Stream2"}, NewConjunctions(actor1))
	actor3 := mockActor1{
		c: make(chan interface{}),
	}
	c3 := NewDefaultConsumer("H", []string{"Stream2", "Stream3"}, NewDisjunctions(&actor3, 3))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterProducer(p3)
	ss.RegisterConsumer(c1)
	ss.RegisterConsumer(c2)
	ss.RegisterConsumer(c3)
	ss.Serve()
	fmt.Println(ss.GenerateDot())
}

func TestPromMetrics(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":19002", nil)

	ss := NewStatefulStreamer()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 := NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	actor := mockActor2{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewDisjunctions(&actor, 3))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterConsumer(c)
	ss.Serve()
	go actor.serve()

	go func() {
		for {
			ss.Send("Stream1", &mockState{"1"})
			time.Sleep(time.Millisecond * 500)
		}
	}()
	go func() {
		for {
			ss.Send("Stream2", &mockState{"2"})
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second * 10)
}

func TestCompoundController(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	var p1, p2, p3 StreamProducer
	var c1, c2 StreamConsumer
	ss := NewStatefulStreamer()
	p1 = NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 = NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	p3 = NewDefaultProducer("C", []string{"Stream3"}, []int{1})

	actor2 := mockActor2{
		c: make(chan interface{}),
	}
	c2 = NewDefaultConsumer("E", []string{"#D", "Stream3"}, NewDisjunctions(&actor2, 3))
	actor1 := NewShortCircuitActor(c2, "#D")
	c1 = NewDefaultConsumer("D", []string{"Stream1", "Stream2"}, NewConjunctions(actor1))

	ss.RegisterProducer(p1)
	ss.RegisterProducer(p2)
	ss.RegisterProducer(p3)
	ss.RegisterConsumer(c1)
	ss.RegisterConsumer(c2)
	ss.Serve()
	fmt.Println(ss.GenerateDot())
	go actor1.Serve()
	go actor2.serve()

	ss.Send("Stream1", &mockState{"1"})
	ss.Send("Stream2", &mockState{"2"})
	ss.Send("Stream3", &mockState{"3"})
	time.Sleep(time.Second)
}
