package aggregator

import (
	"github.com/arcology-network/component-lib/actor"
)

type MsgBroker interface {
	Send(string, *actor.Message) error
}
