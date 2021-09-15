package workers

import (
	ethCommon "github.com/arcology-network/3rd-party/eth/common"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/aggregator/aggregator"
	"github.com/arcology-network/component-lib/log"
	"go.uber.org/zap"
)

type EuResultsAggreSelector struct {
	actor.WorkerThread
	aggregator   *aggregator.Aggregator
	savedMessage *actor.Message
}

//return a Subscriber struct
func NewEuResultsAggreSelector(concurrency int, groupid string) *EuResultsAggreSelector {
	agg := EuResultsAggreSelector{}
	agg.Set(concurrency, groupid)
	agg.aggregator = aggregator.NewAggregator()
	return &agg
}

func (a *EuResultsAggreSelector) OnStart() {

}

func (a *EuResultsAggreSelector) OnMessageArrived(msgs []*actor.Message) error {
	switch msgs[0].Name {
	case actor.MsgClearEuresults:
		remainingQuantity := a.aggregator.OnClearInfoReceived()
		a.AddLog(log.LogLevel_Info, " ...... clear pool", zap.Uint64("msgHeight", msgs[0].Height), zap.Int("remainingQuantity", remainingQuantity))
		a.MsgBroker.Send(actor.MsgClearCompletedEuresults, "true")
	case actor.MsgInclusiveEuresults:
		msg := msgs[0]
		a.savedMessage = msg.CopyHeader()
		inclusive := msg.Data.(*types.InclusiveList)
		a.AddLog(log.LogLevel_Info, " ...... execMsgInclusive", zap.Uint64("msgHeight", msg.Height), zap.Int("inclusive.HashList", len(inclusive.HashList)))
		inclusive.Mode = types.InclusiveMode_Results
		copyInclusive := inclusive.CopyListAddHeight(msg.Height, msg.Round)
		result, _ := a.aggregator.OnListReceived(copyInclusive)
		a.SendMsg(result, a.MsgBroker)
	case actor.MsgEuResults:
		data := msgs[0].Data.(*types.Euresults)
		if data != nil && len(*data) > 0 {
			for _, v := range *data {
				euresult := v
				hash := ethCommon.BytesToHash([]byte(euresult.H))
				newhash := common.ToNewHash(hash, msgs[0].Height, msgs[0].Round)
				result := a.aggregator.OnDataReceived(newhash, euresult)
				a.SendMsg(result, a.MsgBroker)
			}
		}
	}
	return nil
}
func (a *EuResultsAggreSelector) SendMsg(selectedData *[]*interface{}, sender *actor.MessageWrapper) {
	if selectedData != nil {
		euresults := make([]*types.EuResult, len(*selectedData))
		for i, euresult := range *selectedData {
			euresults[i] = (*euresult).(*types.EuResult)
		}

		a.LatestMessage = a.savedMessage
		a.MsgBroker.LatestMessage = a.savedMessage

		a.AddLog(log.LogLevel_Info, "send gather result", zap.Int("counts", len(euresults)))
		sender.Send(actor.MsgExecuted, &euresults)
		a.MsgBroker.Send(actor.MsgSelectCompletedEuresults, "true")

	}
}
