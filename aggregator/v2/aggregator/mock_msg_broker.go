package aggregator

import (
	"github.com/HPISTechnologies/component-lib/actor"
)

type mockMsgBroker struct {
	recvName string
	recvMsg  *actor.Message
}

func (broker *mockMsgBroker) Send(name string, msg *actor.Message) error {
	broker.recvName = name
	broker.recvMsg = msg
	return nil
}
