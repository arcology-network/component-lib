package storage

import (
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/concurrenturl"

	"go.uber.org/zap"

	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/log"

	urlcommon "github.com/HPISTechnologies/concurrenturl/common"
	"github.com/HPISTechnologies/concurrenturl/type/commutative"
)

type ApcWorker struct {
	actor.WorkerThread
	DB     urlcommon.DB
	height uint64
}

//return a Subscriber struct
func NewApcWorker(lanes int, groupid string) *ApcWorker {
	awt := ApcWorker{}
	awt.Set(lanes, groupid)
	return &awt
}

func (awt *ApcWorker) OnStart() {
	persistentDB := urlcommon.NewDataStore()
	meta, _ := commutative.NewMeta(urlcommon.ACCOUNT_BASE_URL)
	persistentDB.Save(urlcommon.ACCOUNT_BASE_URL, meta)
	awt.DB = persistentDB
}

func GetTransitions(euresults []*types.EuResult) ([]uint32, []urlcommon.UnivalueInterface) {
	txIds := make([]uint32, len(euresults))
	transitionsize := 0
	for i, euresult := range euresults {
		transitionsize = transitionsize + len(euresult.Transitions)
		txIds[i] = euresult.ID
	}
	transitions := make([]urlcommon.UnivalueInterface, 0, transitionsize)
	for _, euresult := range euresults {
		for _, transition := range euresult.Transitions {
			transitions = append(transitions, transition)
		}
	}
	return txIds, transitions
}
func (awt *ApcWorker) OnMessageArrived(msgs []*actor.Message) error {
	var euresults *[]*types.EuResult
	result := ""
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgBlockCompleted:
			result = v.Data.(string)
			awt.height = v.Height
		case actor.MsgExecuted:
			euresults = v.Data.(*[]*types.EuResult)
			if v.Height == 0 {
				awt.MsgBroker.Send(actor.MsgClearCompletedEuresults, "true")
			}
		}
	}

	switch result {
	case actor.MsgBlockCompleted_Success:
		if euresults != nil {
			url := concurrenturl.NewConcurrentUrl(awt.DB)
			txIds, transitions := GetTransitions(*euresults)
			awt.AddLog(log.LogLevel_Info, "=====================================", zap.Int("transitions", len(transitions)), zap.Int("txIds", len(txIds)))
			url.Commit(transitions, txIds)
		}
	case actor.MsgBlockCompleted_Failed:

	}

	awt.AddLog(log.LogLevel_Info, "4=ApcWorkerThread send MsgApcHandle---->", zap.Uint64("height", awt.height))
	awt.MsgBroker.Send(actor.MsgApcHandle, &awt.DB)
	return nil
}
