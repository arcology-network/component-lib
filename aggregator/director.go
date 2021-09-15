package aggregator

import (
	"github.com/arcology-network/component-lib/actor"
)

type Director struct {
	actor.WorkerThread
	msgRelation map[string]string
}

//return a Subscriber struct
func NewDirector(concurrency int, groupid string, msgRelation map[string]string) *Director {
	d := Director{}
	d.Set(concurrency, groupid)
	d.msgRelation = msgRelation
	return &d
}

func (d *Director) OnStart() {
}

func (d *Director) Stop() {

}

func (d *Director) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		if sendMsg, ok := d.msgRelation[v.Name]; ok {
			d.LatestMessage = v
			d.MsgBroker.LatestMessage = v
			d.MsgBroker.Send(sendMsg, v.Data)
		}
	}
	return nil
}
