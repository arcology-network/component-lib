package lib

import (
	"fmt"

	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/log"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	//"go.uber.org/zap"
)

type OnMessagReceived func(msg *actor.Message) error // message callback function

// MsgComsumer is the Kafka consumer client used
type ComIncoming struct {
	consumers []Consumer
	cfg       *cluster.Config
	topics    []string
	addrs     []string
	groupid   string

	workThreadName string
	exitChan       []chan bool

	whiteList map[string]string
	buffers   []map[uint64][]KafkaMessage

	onMessagReceived OnMessagReceived
}

// start a consumer for receive data
func (ci *ComIncoming) Start(addrs, topics, whiteList []string, groupid, workThreadName string, callback OnMessagReceived) {
	config := cluster.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	topicSize := len(topics)
	ci.workThreadName = workThreadName
	ci.groupid = groupid
	ci.addrs = addrs
	ci.topics = topics
	ci.whiteList = make(map[string]string)
	ci.buffers = make([]map[uint64][]KafkaMessage, topicSize)
	ci.cfg = config

	ci.consumers = make([]Consumer, topicSize)
	ci.exitChan = make([]chan bool, topicSize)
	ci.onMessagReceived = callback

	for _, v := range whiteList {
		ci.whiteList[v] = v
	}

	for i, topic := range topics {
		// init consumer
		consumer, err := consumerCreator(addrs, groupid, []string{topic}, ci.cfg)
		if err != nil {
			ci.AddLog(actor.NewMessage(), log.LogLevel_Info, "kafka NewConsumer error: "+fmt.Sprintf("%v", err))
			return
		}
		ci.consumers[i] = consumer
		ci.buffers[i] = map[uint64][]KafkaMessage{}

		// consume notifications
		go func(idx int) {
			for ntf := range ci.consumers[idx].Notifications() {
				ci.AddLog(actor.NewMessage(), log.LogLevel_Info, "kafka Consumer notification: "+fmt.Sprintf("%v", ntf))
			}
		}(i)

		// Error notifications
		go func(idx int) {
			for error := range ci.consumers[idx].Errors() {
				ci.AddLog(actor.NewMessage(), log.LogLevel_Info, "kafka Consumer notification: "+fmt.Sprintf("%v", error))
			}
		}(i)

		ci.exitChan[i] = make(chan bool, 1)

		go func(idx int) {
			for {
				select {
				case consumerMsg, ok := <-ci.consumers[idx].Messages():
					if ok {
						if msg := ci.TryReassemble(consumerMsg.Value, idx); msg != nil {
							if _, ok := ci.whiteList[msg.Name]; ok {
								ci.onMessagReceived(msg)
							} else {
								ci.AddLog(msg, log.LogLevel_Debug, "kafka Consumer received message but not in whiltlist(v2):"+msg.Name)
							}
							ci.consumers[idx].MarkOffset(consumerMsg, "") // mark message as processed
						}
					}
				case stop := <-ci.exitChan[idx]:
					if stop {
						return
					}
				}
			}
		}(i)
	}
}

func (ci *ComIncoming) Stop() {
	for i, _ := range ci.topics {
		ci.consumers[i].Close()
		close(ci.exitChan[i])
	}
}

func (ci *ComIncoming) TryReassemble(data []byte, bufIndex int) *actor.Message {
	newKfkMsg := NewKafkaMessage()
	newKfkMsg.Decode(data)

	if _, ok := ci.buffers[bufIndex][newKfkMsg.uuid]; !ok {
		ci.buffers[bufIndex][newKfkMsg.uuid] = make([]KafkaMessage, 0, newKfkMsg.total)
	}
	ci.buffers[bufIndex][newKfkMsg.uuid] = append(ci.buffers[bufIndex][newKfkMsg.uuid], *newKfkMsg)
	msg, err := KafkaMsgToSource(ci.buffers[bufIndex][newKfkMsg.uuid])
	if err != nil {
		ci.AddLog(msg, log.LogLevel_Error, "KafkaMsgToSource  err:"+err.Error())
		return nil
	}
	if msg != nil {
		delete(ci.buffers[bufIndex], newKfkMsg.uuid)
		return msg
	}
	return nil
}

func (ci *ComIncoming) AddLog(msg *actor.Message, level string, content string) {
	log.Logger.AddLog(
		0,
		level,
		msg.Name,
		ci.workThreadName,
		content,
		log.LogType_Msg,
		msg.Height,
		msg.Round,
		msg.Msgid,
		int(msg.Size()),
	)
}
