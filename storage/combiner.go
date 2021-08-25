package storage

import (
	"github.com/arcology/component-lib/actor"
)

type CombinerElements struct {
	Msgs map[string]*actor.Message
}

func (cl *CombinerElements) Get(msgname string) *actor.Message {
	if msg, ok := cl.Msgs[msgname]; ok {
		return msg
	} else {
		return nil
	}
}

type Combiner struct {
	actor.WorkerThread
	inMsgs []string
	outMsg string
	els    *CombinerElements
}

//return a Subscriber struct
func NewCombiner(concurrency int, groupid string, outMsg string) *Combiner {
	c := Combiner{}
	c.outMsg = outMsg
	c.els = &CombinerElements{
		Msgs: map[string]*actor.Message{},
	}
	return &c
}

func (c *Combiner) OnStart() {
}

func (c *Combiner) Stop() {

}

func (c *Combiner) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		c.els.Msgs[v.Name] = v
	}
	c.MsgBroker.Send(c.outMsg, c.els)
	c.els = &CombinerElements{
		Msgs: map[string]*actor.Message{},
	}
	return nil
}
