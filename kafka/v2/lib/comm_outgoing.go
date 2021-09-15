package lib

import (
	//"errors"
	"fmt"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"github.com/Shopify/sarama"
)

type ComOutgoing struct {
	producer    sarama.AsyncProducer
	topics      []string
	addrs       []string
	topicLookup map[string]string // topic to message type topicLookup

	outQueue       chan *actor.Message
	exitChan       chan bool
	workThreadName string
}

// Start the producer.
func (co *ComOutgoing) Start(addrs []string, relations map[string]string, workThreadName string) error {
	co.topics = common.GetUniqueValue(&relations)
	co.addrs = addrs
	co.topicLookup = relations
	co.workThreadName = workThreadName

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// config.Producer.Return.Successes = true
	config.Producer.Flush.Messages = 1
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_0_0_0

	if producer, err := asyncProducerCreator(co.addrs, config); err != nil {
		co.AddLog(actor.NewMessage(), log.LogLevel_Error, "kafka NewAsyncProducer err: "+fmt.Sprintf("%v", err))
		return err
	} else {
		co.producer = producer
	}

	go func() {
		for err := range co.producer.Errors() {
			co.AddLog(actor.NewMessage(), log.LogLevel_Error, "kafka NewAsyncProducer err: "+fmt.Sprintf("%v", err))
		}
	}()

	//start queue
	co.outQueue = make(chan *actor.Message, 1000)
	co.exitChan = make(chan bool, 0)

	go func() {
		for {
			select {
			case msg := <-co.outQueue:
				co.Send(msg)
			case <-co.exitChan:
				break
			}
		}
	}()

	return nil
}

// Stop the producer.
func (co *ComOutgoing) Stop() {
	close(co.exitChan)
	close(co.outQueue)
	co.producer.Close()
}

func (co *ComOutgoing) Send(msg *actor.Message) {
	if topic, ok := co.topicLookup[msg.Name]; !ok {
		co.AddLog(msg, log.LogLevel_Error, "Topic: "+topic+" not found")
		return
	}
	co.AddLog(msg, log.LogLevel_Info, "Start sending messages")

	kfkMsgs, err := SourceToKafkaMsg(msg, KafkaPackageMaxSize-KafkaHeaderSize)
	if err != nil {
		co.AddLog(msg, log.LogLevel_Error, fmt.Sprintf("send messages err : %v", err))
		return
	}
	for _, v := range kfkMsgs {
		co.producer.Input() <- &sarama.ProducerMessage{
			Topic: co.topicLookup[msg.Name],
			Value: sarama.ByteEncoder(*v.Encode()),
		}
	}
	co.AddLog(msg, log.LogLevel_Info, fmt.Sprintf("Message sent %v", len(kfkMsgs)))

	return
}

func (co *ComOutgoing) AddLog(msg *actor.Message, level string, content string) {
	log.Logger.AddLog(
		0,
		level,
		msg.Name,
		co.workThreadName,
		content,
		log.LogType_Msg,
		msg.Height,
		msg.Round,
		msg.Msgid,
		(int)(msg.Size()),
	)
}
