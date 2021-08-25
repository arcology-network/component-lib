package kafka

import (
	"strings"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/kafka/v2/lib"
	"github.com/spf13/viper"
)

type KafkaUploader struct {
	actor.WorkerThread
	outgoing  *lib.ComOutgoing
	relations map[string]string
}

//return a Subscriber struct
func NewKafkaUploader(concurrency int, groupid string, relations map[string]string) *KafkaUploader {
	uploader := KafkaUploader{}
	uploader.Set(concurrency, groupid)
	uploader.relations = relations
	return &uploader
}

func (ku *KafkaUploader) OnStart() {
	mqaddr := viper.GetString("mqaddr")
	ku.outgoing = new(lib.ComOutgoing)
	if err := ku.outgoing.Start(strings.Split(mqaddr, ","), ku.relations, ku.Name); err != nil {
		panic(err)
	}
}

func (ku *KafkaUploader) OnMessageArrived(msgs []*actor.Message) error {
	ku.outgoing.Send(msgs[0])
	return nil
}
