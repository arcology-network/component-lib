package lib

import (
	"math"
	"sort"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
)

const (
	KafkaHeaderSize     = 1 + 8 + 4 + 4 // version (1byte)+ msgid(8byte)+total packages(4bytes)+package idx (4byte)
	KafkaPackageMaxSize = 512 * 1024    //default 1MB / package
	KafkaVersion        = uint8(0)
)

type KafkaMessage struct {
	version  uint8
	uuid     uint64
	total    uint32
	serialID uint32
	data     []byte
}

func NewKafkaMessage() *KafkaMessage {
	return &KafkaMessage{
		version:  KafkaVersion,
		uuid:     0,
		total:    0,
		serialID: 0,
		data:     nil,
	}
}

type KafkaMessages []KafkaMessage

func (msgs KafkaMessages) Len() int           { return len(msgs) }
func (msgs KafkaMessages) Swap(i, j int)      { msgs[i], msgs[j] = msgs[j], msgs[i] }
func (msgs KafkaMessages) Less(i, j int) bool { return msgs[i].serialID < msgs[j].serialID }

func SourceToKafkaMsg(msg *actor.Message, maxSize int) ([]KafkaMessage, error) {
	data, err := msg.Encode()
	if err != nil {
		return []KafkaMessage{}, err
	}

	kfkMsgs := make([]KafkaMessage, uint32(math.Ceil(float64(len(data))/float64(maxSize))))

	uuid := common.GenerateUUID()
	for i := range kfkMsgs {
		kfkMsgs[i].uuid = uuid
		kfkMsgs[i].total = uint32(len(kfkMsgs))
		kfkMsgs[i].serialID = uint32(i)
		kfkMsgs[i].data = data[i*maxSize : int(math.Min(float64((i+1)*maxSize), float64(len(data))))]
	}

	return kfkMsgs, nil

}

func KafkaMsgToSource(buffer []KafkaMessage) (*actor.Message, error) {
	if len(buffer) == int(buffer[0].total) { // all received
		sort.Sort(KafkaMessages(buffer))
		data := make([]byte, 0)
		for _, v := range buffer {
			data = append(data, v.data...)
		}

		msg := actor.NewMessage()
		err := msg.Decode(data)
		return msg, err

	}
	if len(buffer) > int(buffer[0].total) {
		panic("Kafka decoding buffer overflowed !")
	}
	return nil, nil
}

func (kfkMsg *KafkaMessage) Encode() *[]byte {
	buffer := make([]byte, KafkaHeaderSize+len(kfkMsg.data))

	copy(buffer[0:], []byte{kfkMsg.version})
	copy(buffer[1:], common.Uint64ToBytes(kfkMsg.uuid))
	copy(buffer[9:], common.Uint32ToBytes(kfkMsg.total))
	copy(buffer[13:], common.Uint32ToBytes(kfkMsg.serialID))
	copy(buffer[KafkaHeaderSize:], kfkMsg.data)

	return &buffer
}

func (kfkMsg *KafkaMessage) Decode(data []byte) {
	kfkMsg.version = data[0]
	kfkMsg.uuid = common.BytesToUint64(data[1:9])       //8byte uuid uint64
	kfkMsg.total = common.BytesToUint32(data[9:13])     //4 total 	uint32
	kfkMsg.serialID = common.BytesToUint32(data[13:17]) //4  idx	uint32
	kfkMsg.data = data[KafkaHeaderSize:]
}
