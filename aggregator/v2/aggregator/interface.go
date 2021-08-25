package aggregator

import (
	"github.com/arcology/component-lib/actor"
)

type MsgBroker interface {
	Send(string, *actor.Message) error
}
