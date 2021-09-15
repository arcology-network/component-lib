package aggregator

import (
	"reflect"

	"github.com/arcology-network/component-lib/actor"
)

type Aggregator struct {
	clearMsg     string
	dataMsg      string
	pickListMsg  string
	clearListMsg string
	outMsg       string
	broker       MsgBroker
	pool         map[string]interface{}
	missingSet   map[string]struct{}
	clearList    []string
	dataSet      map[string]interface{}
	dataType     reflect.Type
}

func NewAggregator(broker MsgBroker, clearMsg, dataMsg, pickListMsg, clearListMsg, outMsg string, dataType reflect.Type) *Aggregator {
	return &Aggregator{
		broker:       broker,
		clearMsg:     clearMsg,
		dataMsg:      dataMsg,
		pickListMsg:  pickListMsg,
		clearListMsg: clearListMsg,
		outMsg:       outMsg,
		dataType:     dataType,
		pool:         make(map[string]interface{}),
	}
}

func (aggr *Aggregator) HandleMsg(msg *actor.Message) {
	switch msg.Name {
	case aggr.clearMsg:
		aggr.clear()
		break
	case aggr.dataMsg:
		aggr.addData(msg)
		break
	case aggr.pickListMsg:
		if msg.Data == nil {
			aggr.setPickList([]string{})
		} else {
			aggr.setPickList(msg.Data.([]string))
		}
		break
	case aggr.clearListMsg:
		if msg.Data == nil {
			aggr.setClearList([]string{})
		} else {
			aggr.setClearList(msg.Data.([]string))
		}
		break
	default:
		// Log error.
		break
	}
}

func (aggr *Aggregator) clear() {
	if aggr.clearList == nil || len(aggr.clearList) == 0 {
		aggr.pool = make(map[string]interface{})
	} else {
		for _, e := range aggr.clearList {
			delete(aggr.pool, e)
		}
	}
	aggr.missingSet = nil
	aggr.clearList = nil
	aggr.dataSet = nil
}

func (aggr *Aggregator) setPickList(list []string) {
	if len(list) == 0 {
		aggr.missingSet = nil
		aggr.dataSet = nil
		aggr.sendData(aggr.pool)
		return
	}

	aggr.missingSet = make(map[string]struct{})
	aggr.dataSet = make(map[string]interface{})
	for _, e := range list {
		if v, ok := aggr.pool[e]; ok {
			aggr.dataSet[e] = v
		} else {
			aggr.missingSet[e] = struct{}{}
		}
	}

	if len(aggr.missingSet) == 0 {
		aggr.sendData(aggr.dataSet)
	}
}

func (aggr *Aggregator) setClearList(list []string) {
	aggr.clearList = list
}

func (aggr *Aggregator) addData(msg *actor.Message) {
	dataList := reflect.ValueOf(msg.Data)
	for i := 0; i < dataList.Len(); i++ {
		data := dataList.Index(i).Interface()
		hash := Hash(data)
		aggr.pool[hash] = data

		if aggr.missingSet != nil && len(aggr.missingSet) > 0 {
			if _, ok := aggr.missingSet[hash]; ok {
				delete(aggr.missingSet, hash)
				aggr.dataSet[hash] = data
			}

			if len(aggr.missingSet) == 0 {
				aggr.sendData(aggr.dataSet)
			}
		}
	}
}

func (aggr *Aggregator) sendData(set map[string]interface{}) {
	dataList := reflect.MakeSlice(reflect.SliceOf(aggr.dataType), 0, len(set))
	for _, v := range set {
		dataList = reflect.Append(dataList, reflect.ValueOf(v))
	}

	// TODO: fill other fields.
	msg := &actor.Message{
		Name: aggr.outMsg,
		Data: dataList.Interface(),
	}
	aggr.broker.Send(aggr.outMsg, msg)
}
