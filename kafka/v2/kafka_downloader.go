package kafka

import (
	"strings"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/kafka/v2/lib"
	"github.com/HPISTechnologies/component-lib/log"

	"github.com/spf13/viper"
)

type KafkaDownloader struct {
	actor.WorkerThread

	inComing     *lib.ComIncoming
	topics       []string
	messageTypes []string
}

//return a Subscriber struct
func NewKafkaDownloader(concurrency int, groupid string, topics, messageTypes []string) *KafkaDownloader {
	downloader := KafkaDownloader{}
	downloader.Set(concurrency, groupid)
	downloader.topics = topics
	downloader.messageTypes = messageTypes
	return &downloader
}

func (kd *KafkaDownloader) OnStart() {
	kd.AddLog(log.LogLevel_Info, "start exec subscriber ")

	mqaddr := viper.GetString("mqaddr")
	kd.inComing = new(lib.ComIncoming)
	kd.inComing.Start(strings.Split(mqaddr, ","), kd.topics, kd.messageTypes, kd.Groupid, kd.Name, kd.onKafkaMessageArrived)

}

func (kd *KafkaDownloader) OnMessageArrived(msgs []*actor.Message) error {
	return nil
}

func (kd *KafkaDownloader) onKafkaMessageArrived(msg *actor.Message) error {
	kd.MsgBroker.SendMessage(msg, actor.MessageFrom_Remote)
	return nil
}
