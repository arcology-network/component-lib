package storage

import (
	"time"

	"github.com/arcology-network/component-lib/actor"
	ccurl "github.com/arcology-network/concurrenturl/v2"

	"go.uber.org/zap"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/log"

	urlcommon "github.com/arcology-network/concurrenturl/v2/common"
	urltype "github.com/arcology-network/concurrenturl/v2/type"
	"github.com/arcology-network/concurrenturl/v2/type/commutative"
)

type ApcWorker struct {
	actor.WorkerThread
	DB       urlcommon.DB
	height   uint64
	url      *ccurl.ConcurrentUrl
	executed *[]*types.EuResult
}

//return a Subscriber struct
func NewApcWorker(lanes int, groupid string) *ApcWorker {
	awt := ApcWorker{}
	awt.Set(lanes, groupid)
	return &awt
}

func (awt *ApcWorker) OnStart() {
	platform := urlcommon.NewPlatform()
	persistentDB := urlcommon.NewDataStore()
	meta, _ := commutative.NewMeta(platform.Eth10Account())
	persistentDB.Save(platform.Eth10Account(), meta)
	awt.DB = persistentDB

	awt.url = ccurl.NewConcurrentUrl(awt.DB)
	awt.executed = &[]*types.EuResult{}
}
func GetTransitionIds(euresults []*types.EuResult) []uint32 {
	txIds := make([]uint32, len(euresults))
	for i, euresult := range euresults {
		txIds[i] = euresult.ID
	}
	return txIds
}
func GetTransitions(euresults []*types.EuResult) ([]uint32, []urlcommon.UnivalueInterface) {
	txIds := make([]uint32, len(euresults))
	transitionsize := 0
	for i, euresult := range euresults {
		transitionsize = transitionsize + len(euresult.Transitions)
		txIds[i] = euresult.ID
	}
	threadNum := 6
	transitionses := make([][]urlcommon.UnivalueInterface, threadNum)
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			univalues := urltype.Univalues{}
			transitionses[index] = append(transitionses[index], univalues.DecodeV2(euresults[i].Transitions)...)
		}
	}
	common.ParallelWorker(len(euresults), threadNum, worker)

	transitions := make([]urlcommon.UnivalueInterface, 0, transitionsize)
	for _, trans := range transitionses {
		transitions = append(transitions, trans...)
	}
	return txIds, transitions
}
func (awt *ApcWorker) OnMessageArrived(msgs []*actor.Message) error {
	result := ""
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgEuResults:
			t := time.Now()
			data := msgs[0].Data.(*types.Euresults)
			_, transitions := GetTransitions(*data)
			awt.url.Import(transitions)
			awt.AddLog(log.LogLevel_Info, "euresult import completed ---->", zap.Duration("time", time.Since(t)))
		case actor.MsgBlockCompleted:
			result = v.Data.(string)
			awt.height = v.Height
			if actor.MsgBlockCompleted_Success == result {
				awt.AddLog(log.LogLevel_Info, "start compute final state---->")
				txIds := GetTransitionIds(*awt.executed)
				awt.url.Commit(txIds)
				awt.AddLog(log.LogLevel_Info, "4=ApcWorkerThread send MsgApcHandle---->", zap.Uint64("height", awt.height))
				awt.MsgBroker.Send(actor.MsgApcHandle, &awt.DB)
			}
			awt.executed = &[]*types.EuResult{}
		case actor.MsgExecuted:
			awt.executed = v.Data.(*[]*types.EuResult)
			if v.Height == 0 {
				_, transitions := GetTransitions(*awt.executed)
				awt.url.Import(transitions)
				awt.MsgBroker.Send(actor.MsgClearCompletedEuresults, "true")
			}
		}
	}

	// switch result {
	// case actor.MsgBlockCompleted_Success:
	// 	if euresults != nil {
	// 		url := ccurl.NewConcurrentUrl(awt.DB)
	// 		txIds, transitions := GetTransitions(*euresults)
	// 		awt.AddLog(log.LogLevel_Info, "=====================================", zap.Int("transitions", len(transitions)), zap.Int("txIds", len(txIds)))
	// 		url.Commit(transitions, txIds)
	// 	}
	// case actor.MsgBlockCompleted_Failed:

	// }

	return nil
}
