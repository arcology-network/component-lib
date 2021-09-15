package kafka

import (
	"strings"

	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
)

type KafkaUploader struct {
	actor.WorkerThread
	outgoing  *lib.ComOutgoing
	relations map[string]string
	mqaddr    string
}

//return a Subscriber struct
func NewKafkaUploader(concurrency int, groupid string, relations map[string]string, mqaddr string) *KafkaUploader {
	uploader := KafkaUploader{}
	uploader.Set(concurrency, groupid)
	uploader.relations = relations
	uploader.mqaddr = mqaddr
	return &uploader
}

func (ku *KafkaUploader) OnStart() {
	ku.outgoing = new(lib.ComOutgoing)
	if err := ku.outgoing.Start(strings.Split(ku.mqaddr, ","), ku.relations, ku.Name); err != nil {
		panic(err)
	}
}

func (ku *KafkaUploader) OnMessageArrived(msgs []*actor.Message) error {
	if msgs[0].From == actor.MessageFrom_Remote {
		return nil
	}
	ku.outgoing.Send(msgs[0])
	return nil
}
