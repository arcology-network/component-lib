package lib

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

const (
	KafkaHeaderLength  = 1 + 8 + 4 + 4 + 4 // version (1byte)+ msgid(8byte)+total packages(4bytes)+package idx (4byte)
	KafkaPackageLength = 999000            //512 * 1024    //default 1M / package
	KafkaVersion       = byte(0)
)

type KafkaMessageHeader struct {
	version          byte
	uuid             uint64
	subPackagerNums  uint32
	subPackagerIndex uint32
	totalSize        uint32
}

func NewKafkaMessageHeader() *KafkaMessageHeader {
	return &KafkaMessageHeader{
		version:          KafkaVersion,
		uuid:             0,
		subPackagerNums:  0,
		subPackagerIndex: 0,
		totalSize:        0,
	}
}

type ComOutgoing struct {
	producer   sarama.AsyncProducer
	topics     []string
	addrs      []string
	relations  map[string]string
	partitions map[string]uint8

	chanSend       chan *actor.Message
	exitChan       chan bool
	workThreadName string
	header         *KafkaMessageHeader
}

func getTopics(relations *map[string]string) []string {
	topics := []string{}
	mtopic := map[string]int{}
	for _, topic := range *relations {
		mtopic[topic] = 0
	}
	for topic, _ := range mtopic {
		topics = append(topics, topic)
	}
	return topics
}

// Start the producer.
func (co *ComOutgoing) Start(addrs []string, relations map[string]string, workThreadName string) error {
	co.topics = getTopics(&relations)
	co.addrs = addrs
	co.relations = relations

	co.workThreadName = workThreadName

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	//config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Partitioner = sarama.NewHashPartitioner
	// config.Producer.Return.Successes = true

	config.Producer.Flush.Messages = 1
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_0_0_0
	producer, err := sarama.NewAsyncProducer(co.addrs, config)
	if err != nil {
		log.Logger.AddLog(0, log.LogLevel_Error, "kafka NewAsyncProducer err", co.workThreadName, fmt.Sprintf("%v", err), log.LogType_Msg, 0, 0, 0, 0)
		return err
	}

	co.partitions = map[string]uint8{}
	co.partitions["local-txs"] = 0
	co.partitions["remote-txs"] = 1
	co.partitions["block-txs"] = 2

	co.producer = producer

	go func() {
		for err := range producer.Errors() {
			log.Logger.AddLog(0, log.LogLevel_Error, "kafka producer err", co.workThreadName, fmt.Sprintf("%v", err), log.LogType_Msg, 0, 0, 0, 0)
		}
	}()

	//start queue
	co.chanSend = make(chan *actor.Message, 1000)
	co.exitChan = make(chan bool, 0)

	go func() {
		for {
			select {
			case si := <-co.chanSend:
				co.sendPack(si)
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
	close(co.chanSend)
	co.producer.Close()
}

//send pack with queue
func (co *ComOutgoing) Send(msg *actor.Message) {
	co.chanSend <- msg
}
func (co *ComOutgoing) sendPack(msg *actor.Message) error {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("GobEncode err data")
	// 		common.Serialization("recoveredData", *msg)
	// 		panic(err)
	// 	}
	// }()
	topic, ok := co.relations[msg.Name]
	if !ok {
		log.Logger.AddLog(0, log.LogLevel_Error, msg.Name, co.workThreadName, "send tpoic not defined", log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, 0)
		return errors.New("send tpoic not defined")
	}

	logid := log.Logger.AddLog(0, log.LogLevel_Info, msg.Name, co.workThreadName, "start send pack", log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, 0)

	startTime := time.Now()
	msg.Msgid = logid
	sendData, err := common.GobEncode(msg)
	if err != nil {
		log.Logger.AddLog(0, log.LogLevel_Error, msg.Name, co.workThreadName, fmt.Sprintf("encode %v", err), log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, 0)
		return err
	}
	elapsed := time.Now().Sub(startTime)
	sendingPackages := co.splitPackage(sendData)
	totalPackages := len(sendingPackages)

	key := []byte{}
	if _, ok := co.partitions[topic]; ok {
		hash := sha256.Sum256(sendData)
		key = hash[:]
	}

	for i := 0; i < totalPackages; i++ {
		header := NewKafkaMessageHeader()
		header.uuid = logid
		header.subPackagerNums = uint32(totalPackages)
		header.subPackagerIndex = uint32(i)
		co.flush(topic, header, sendingPackages[i], key)
		log.Logger.AddLog(0, log.LogLevel_Info, msg.Name, co.workThreadName, "subpack send  completed", log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, len(sendingPackages[i]), zap.Uint64("uuid", header.uuid), zap.Int("packIdx", i))
	}
	log.Logger.AddLog(0, log.LogLevel_Info, msg.Name, co.workThreadName, "pack send completed", log.LogType_Msg, msg.Height, msg.Round, msg.Msgid, len(sendData), zap.Duration("encodeTime", elapsed), zap.Int("totals", totalPackages))
	return nil
}

//transport layer protocol
// version                                                                      byte		default 0
// uuid	                                                                           uint64
// total packages  														uint32
// index of current package in all packages  uint32
// actual data ,the encoded result of sending message

// Send an data to Kafka broker.
func (co *ComOutgoing) flush(topic string, header *KafkaMessageHeader, data, key []byte) {
	totalsize := KafkaHeaderLength + len(data)
	sendData := make([]byte, totalsize)
	header.totalSize = uint32(totalsize)
	bz := 0

	bz += copy(sendData[bz:], []byte{KafkaVersion}) //version

	bysUuid := common.Uint64ToBytes(header.uuid)
	bz += copy(sendData[bz:], bysUuid) //total

	bysTotal := common.Uint32ToBytes(header.subPackagerNums)
	bz += copy(sendData[bz:], bysTotal) //total

	bysIdx := common.Uint32ToBytes(header.subPackagerIndex)
	bz += copy(sendData[bz:], bysIdx) //idx

	totalSizeBys := common.Uint32ToBytes(header.totalSize)
	bz += copy(sendData[bz:], totalSizeBys) //idx

	bz += copy(sendData[bz:], data) //data

	producerMessage := sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(sendData),
	}
	if len(key) > 0 {
		producerMessage.Key = sarama.ByteEncoder(key)
	}

	co.producer.Input() <- &producerMessage
	return
}

func (co *ComOutgoing) splitPackage(data []byte) [][]byte {
	dataLength := KafkaPackageLength - KafkaHeaderLength
	totalSize := len(data)
	packs := totalSize / dataLength
	if totalSize%dataLength > 0 {
		packs = packs + 1
	}
	subPackages := make([][]byte, packs)
	for i := 0; i < packs; i++ {
		idxEnd := (i + 1) * dataLength
		if idxEnd > totalSize {
			idxEnd = totalSize
		}
		subPackages[i] = data[i*dataLength : idxEnd]
	}
	return subPackages
}
