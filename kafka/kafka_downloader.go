package kafka

import (
	"strings"

	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
	"github.com/arcology-network/component-lib/log"
)

type KafkaDownloader struct {
	actor.WorkerThread

	inComing     *lib.ComIncoming
	topics       []string
	messageTypes []string
	mqaddr       string
}

//return a Subscriber struct
func NewKafkaDownloader(concurrency int, groupid string, topics, messageTypes []string, mqaddr string) actor.IWorker {
	downloader := KafkaDownloader{}
	downloader.Set(concurrency, groupid)
	downloader.topics = topics
	downloader.messageTypes = messageTypes
	downloader.mqaddr = mqaddr
	return &downloader
}

func (kd *KafkaDownloader) OnStart() {
	kd.AddLog(log.LogLevel_Info, "start exec subscriber ")
	kd.inComing = new(lib.ComIncoming)
	kd.inComing.Start(strings.Split(kd.mqaddr, ","), kd.topics, kd.messageTypes, kd.Groupid, kd.Name, kd.onKafkaMessageArrived)

}

func (kd *KafkaDownloader) OnMessageArrived(msgs []*actor.Message) error {
	return nil
}

func (kd *KafkaDownloader) onKafkaMessageArrived(msg *actor.Message) error {
	kd.MsgBroker.SendMessage(msg, actor.MessageFrom_Remote)
	return nil
}
