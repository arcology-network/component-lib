package v2

import (
	"reflect"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/aggregator/v2/aggregator"
)

type AggreSelector struct {
	actor.WorkerThread
	impl *aggregator.Aggregator
}

type brokerAdaptor struct {
	wt *actor.WorkerThread
}

func (adaptor *brokerAdaptor) Send(name string, msg *actor.Message) error {
	adaptor.wt.MsgBroker.Send(name, msg)
	return nil
}

func NewAggreSelector(
	concurrency int,
	groupID string,
	clearMsg, dataMsg, pickListMsg, clearListMsg, outMsg string,
	dataType reflect.Type) *AggreSelector {
	aggreSelector := &AggreSelector{}
	aggreSelector.Set(concurrency, groupID)
	aggreSelector.impl = aggregator.NewAggregator(
		&brokerAdaptor{wt: &aggreSelector.WorkerThread},
		clearMsg, dataMsg, pickListMsg, clearListMsg, outMsg,
		dataType)
	return aggreSelector
}

func (as *AggreSelector) OnStart() {}

func (as *AggreSelector) OnMessageArrived(msgs []*actor.Message) error {
	as.impl.HandleMsg(msgs[0])
	return nil
}
