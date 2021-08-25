package lib

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/arcology/component-lib/actor"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	cluster "github.com/bsm/sarama-cluster"
)

var Workload []byte
var Msg *actor.Message

//callback when all finaish
func onMessageReceived(msg *actor.Message) error {
	fmt.Printf("msg received\n")
	if Msg.Msgid != msg.Msgid ||
		Msg.Name != msg.Name ||
		Msg.Height != msg.Height ||
		Msg.Round != msg.Round ||
		bytes.Compare(Msg.Data.([]byte), msg.Data.([]byte)) != 0 {
		panic("Failed to check msg.")
	}
	return nil
}

func TestBasicFunction(t *testing.T) {
	Workload = []byte{123, 43, 67, 98}
	Msg = &actor.Message{
		Msgid:  99999,
		Name:   "tx",
		Height: 21,
		Round:  1,
		Data:   Workload,
	}
	var mockProducer *mocks.AsyncProducer
	asyncProducerCreator = func(brokers []string, config *sarama.Config) (sarama.AsyncProducer, error) {
		config.Producer.Return.Successes = true
		mockProducer = mocks.NewAsyncProducer(t, config)
		return mockProducer, nil
	}
	consumers := make(map[string]Consumer)
	consumerCreator = func(brokers []string, groupID string, topics []string, config *cluster.Config) (Consumer, error) {
		consumers[topics[0]] = createMockConsumer(brokers, groupID, topics, config)
		return consumers[topics[0]], nil
	}

	addr := "localhost:9092"
	relations := map[string]string{
		"tx":       "log",
		"noderole": "msgexch",
	}

	outgoing := new(ComOutgoing)
	if err := outgoing.Start([]string{addr}, relations, "uploader"); err != nil {
		panic(err)
	}
	mockProducer.ExpectInputAndSucceed()

	fmt.Printf("send start time %v\n", time.Now())
	outgoing.Send(Msg)

	kMsg := <-mockProducer.Successes()
	bArray, _ := kMsg.Value.Encode()
	if kMsg.Topic != "log" {
		t.Error("Failed to send message to kafka.")
		return
	}

	incoming := new(ComIncoming)
	incoming.Start([]string{addr}, []string{"log", "msgexch"}, []string{"tx", "noderole"}, "groupids", "downloader", onMessageReceived)
	consumers["log"].(*mockConsumer).msgChan <- &sarama.ConsumerMessage{
		Value: bArray,
	}

	time.Sleep(time.Second * 3)
}

func TestBigWorkload(t *testing.T) {
	Workload = make([]byte, 1024*1024)
	for i := range Workload {
		Workload[i] = byte(i % 256)
	}
	Msg = &actor.Message{
		Msgid:  99999,
		Name:   "tx",
		Height: 21,
		Round:  1,
		Data:   Workload,
	}
	var mockProducer *mocks.AsyncProducer
	asyncProducerCreator = func(brokers []string, config *sarama.Config) (sarama.AsyncProducer, error) {
		config.Producer.Return.Successes = true
		mockProducer = mocks.NewAsyncProducer(t, config)
		return mockProducer, nil
	}
	consumers := make(map[string]Consumer)
	consumerCreator = func(brokers []string, groupID string, topics []string, config *cluster.Config) (Consumer, error) {
		consumers[topics[0]] = createMockConsumer(brokers, groupID, topics, config)
		return consumers[topics[0]], nil
	}

	addr := "localhost:9092"
	relations := map[string]string{
		"tx":       "log",
		"noderole": "msgexch",
	}

	outgoing := new(ComOutgoing)
	if err := outgoing.Start([]string{addr}, relations, "uploader"); err != nil {
		panic(err)
	}
	mockProducer.ExpectInputAndSucceed()
	mockProducer.ExpectInputAndSucceed()
	mockProducer.ExpectInputAndSucceed()

	fmt.Printf("send start time %v\n", time.Now())
	outgoing.Send(Msg)
	fmt.Printf("send start time end %v\n", time.Now())
	kMsg1 := <-mockProducer.Successes()
	kMsg2 := <-mockProducer.Successes()
	kMsg3 := <-mockProducer.Successes()
	bArray1, _ := kMsg1.Value.Encode()
	bArray2, _ := kMsg2.Value.Encode()
	bArray3, _ := kMsg3.Value.Encode()
	if kMsg1.Topic != "log" || kMsg2.Topic != "log" || kMsg3.Topic != "log" {
		t.Error("Failed to send message to kafka.")
		return
	}

	incoming := new(ComIncoming)
	incoming.Start([]string{addr}, []string{"log", "msgexch"}, []string{"tx", "noderole"}, "groupids", "downloader", onMessageReceived)
	consumers["log"].(*mockConsumer).msgChan <- &sarama.ConsumerMessage{
		Value: bArray1,
	}
	consumers["log"].(*mockConsumer).msgChan <- &sarama.ConsumerMessage{
		Value: bArray2,
	}
	consumers["log"].(*mockConsumer).msgChan <- &sarama.ConsumerMessage{
		Value: bArray3,
	}

	time.Sleep(time.Second * 3)
}
