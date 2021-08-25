package kafka

import (
	"fmt"
	"testing"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/streamer"
)

// Download implements IWorker interface.
type Downloader struct {
	messageTypes []string
	broker       *actor.MessageWrapper
}

func NewDownloader(_ int, _ string, _, messageTypes []string, _ string) actor.IWorker {
	t.Log("NewDownloader")
	return &Downloader{
		messageTypes: messageTypes,
	}
}

func (d *Downloader) Init(wtName string, broker *streamer.StatefulStreamer) {
	t.Log("Downloader.Init")
	d.broker = &actor.MessageWrapper{
		MsgBroker:      broker,
		LatestMessage:  actor.NewMessage(),
		WorkThreadName: wtName,
	}
}

func (d *Downloader) ChangeEnvironment(_ *actor.Message) {

}

func (d *Downloader) OnStart() {
	t.Log("Downloader.OnStart")
}

func (d *Downloader) OnMessageArrived(_ []*actor.Message) error {
	return nil
}

// Receive is used for testing.
func (d *Downloader) Receive(msg *actor.Message) {
	for _, typ := range d.messageTypes {
		if typ == msg.Name {
			d.broker.SendMessage(msg, actor.MessageFrom_Remote)
			return
		}
	}
	panic(fmt.Sprintf("unknown message type got: %v", msg.Name))
}

var t testing.TB

func NewDownloaderCreator(logger testing.TB) func(int, string, []string, []string, string) actor.IWorker {
	t = logger
	return NewDownloader
}
