package storage

import (
	"github.com/HPISTechnologies/component-lib/actor"
)

type Switcher struct {
	actor.WorkerThread
	inTopic  string
	outTopic string
}

//return a Subscriber struct
func NewSwitcher(concurrency int, groupid, in, out string) *Switcher {
	switcher := Switcher{}
	switcher.Set(concurrency, groupid)
	switcher.inTopic = in
	switcher.outTopic = out
	return &switcher
}

func (switcher *Switcher) OnStart() {

}

func (switcher *Switcher) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case switcher.inTopic:
			switcher.MsgBroker.Send(switcher.outTopic, v.Data)
		}
	}

	return nil
}
