package lib

import (
	"fmt"
	"time"

	"github.com/arcology/common-lib/common"
	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/log"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"go.uber.org/zap"
)

// MsgComsumer is the Kafka consumer client used
type ComIncoming struct {
	consumers []*cluster.Consumer
	cfg       *cluster.Config
	topics    []string
	addrs     []string
	groupid   string

	msgTypes         map[string]int
	chStops          []chan bool
	onMessagReceived OnMessagReceived
	buffers          []map[uint64][]byte
	workThreadName   string
}

// one package is received,call this func
type OnMessagReceived func(msg *actor.Message) error

// start a consumer for receive data
func (ci *ComIncoming) Start(addrs, topics, msgs []string, groupid, workThreadName string, callback OnMessagReceived) {
	ci.msgTypes = map[string]int{}
	for i, msgType := range msgs {
		ci.msgTypes[msgType] = i
	}

	ci.workThreadName = workThreadName
	ci.groupid = groupid
	ci.addrs = addrs
	ci.topics = topics
	ci.onMessagReceived = callback

	config := cluster.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	ci.cfg = config

	topicSize := len(topics)
	ci.consumers = make([]*cluster.Consumer, topicSize)
	ci.buffers = make([]map[uint64][]byte, topicSize)
	ci.chStops = make([]chan bool, topicSize)

	for i, topic := range topics {
		ci.buffers[i] = map[uint64][]byte{}
		// init consumer
		consumer, err := cluster.NewConsumer(addrs, groupid+"-"+topic, []string{topic}, ci.cfg)
		if err != nil {
			log.Logger.AddLog(0, log.LogLevel_Error, "kafka NewConsumer err", ci.workThreadName, fmt.Sprintf("%v", err), log.LogType_Msg, 0, 0, 0, 0)
			return
		}

		ci.consumers[i] = consumer

		// consume errors
		go func(idx int) {
			for err := range ci.consumers[idx].Errors() {
				log.Logger.AddLog(0, log.LogLevel_Error, "kafka Consumer err", ci.workThreadName, fmt.Sprintf("%v", err), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", ci.topics[idx]))
			}
		}(i)

		// consume notifications
		go func(idx int) {
			for ntf := range ci.consumers[idx].Notifications() {
				log.Logger.AddLog(0, log.LogLevel_Info, "kafka Consumer notification", ci.workThreadName, fmt.Sprintf("%v", ntf), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", ci.topics[idx]))
			}
		}(i)

		ci.chStops[i] = make(chan bool, 1)

		go func(idx int) {
			for {
				select {
				case msg, ok := <-ci.consumers[idx].Messages():
					if ok {
						ci.stripMsgHeader(msg, idx)
						ci.consumers[idx].MarkOffset(msg, "") // mark message as processed
					}
				case stop := <-ci.chStops[idx]:
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
		close(ci.chStops[i])
	}

}

func (ci *ComIncoming) stripMsgHeader(msg *sarama.ConsumerMessage, bufIdx int) {

	//transport layer protocol
	// version                                                                      byte		default 0
	// uuid	                                                                           uint64
	// total packages  														uint32
	// index of current package in all packages  uint32
	// actual data ,the encoded result of sending message

	data := msg.Value
	if len(data) < KafkaHeaderLength {
		log.Logger.AddLog(0, log.LogLevel_Error, "received data err", ci.workThreadName, fmt.Sprintf("received raw data too shorter,data length must be greater than %d", KafkaHeaderLength), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", msg.Topic))
		return
	}
	if data[0] != KafkaVersion {
		log.Logger.AddLog(0, log.LogLevel_Error, "received data err", ci.workThreadName, fmt.Sprintf("protocol version error,except %d,but get %d", KafkaVersion, data[0]), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", msg.Topic))
		return
	}
	header := NewKafkaMessageHeader()
	header.version = data[0]
	header.uuid = common.BytesToUint64(data[1:9])               //8byte uuid uint64
	header.subPackagerNums = common.BytesToUint32(data[9:13])   //4 total 	uint32
	header.subPackagerIndex = common.BytesToUint32(data[13:17]) //4  idx	uint32
	header.totalSize = common.BytesToUint32(data[17:21])        //4 totalsize uint32
	actualData := data[21:]

	if header.totalSize != uint32(len(data)) {
		log.Logger.AddLog(0, log.LogLevel_Error, "received data err", ci.workThreadName, fmt.Sprintf("protocol version error,except %d,but get %d", KafkaVersion, data[0]), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", msg.Topic), zap.Uint32("send data", header.totalSize), zap.Int("received data", len(data)))
		return
	}

	ci.assembleData(header, msg.Topic, bufIdx, actualData)
}

func (ci *ComIncoming) assembleData(header *KafkaMessageHeader, topic string, bufIdx int, actualData []byte) {
	savedData := ci.buffers[bufIdx][header.uuid]
	if savedData == nil || len(savedData) == 0 {
		log.Logger.AddLog(0, log.LogLevel_Debug, "unknown", ci.workThreadName, "start received package data", log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", ci.topics[bufIdx]))
		savedData = actualData
	} else {
		savedData = append(savedData, actualData...)
	}
	ci.buffers[bufIdx][header.uuid] = savedData

	if header.subPackagerNums == header.subPackagerIndex+1 {
		var msg actor.Message
		startTime := time.Now()
		err := common.GobDecode(savedData, &msg)
		delete(ci.buffers[bufIdx], header.uuid)
		if err != nil {
			log.Logger.AddLog(0, log.LogLevel_Error, "received data err", ci.workThreadName, fmt.Sprintf("message  decode err=%v", err), log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", ci.topics[bufIdx]), zap.Int("savedData length", len(savedData)))
			return
		}

		if _, ok := ci.msgTypes[msg.Name]; !ok {
			return
		}

		logid := log.Logger.AddLog(0, log.LogLevel_Info, msg.Name, ci.workThreadName, "received data completed", log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, len(actualData), zap.String("topic", ci.topics[bufIdx]), zap.Duration("decodeTime", time.Now().Sub(startTime)))
		msg.Msgid = logid
		ci.onMessagReceived(&msg)
	} else {
		log.Logger.AddLog(0, log.LogLevel_Debug, "unknown", ci.workThreadName, "this is a subpackage", log.LogType_Msg, 0, 0, 0, 0, zap.String("topic", ci.topics[bufIdx]))

	}
}
