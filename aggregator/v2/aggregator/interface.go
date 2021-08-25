package aggregator

import (
	"github.com/HPISTechnologies/component-lib/actor"
)

type MsgBroker interface {
	Send(string, *actor.Message) error
}
