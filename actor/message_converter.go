package actor

import (
	"fmt"

	"github.com/arcology/component-lib/log"
	"github.com/arcology/component-lib/streamer"
)

type Actor struct {
	sender   *MessageWrapper
	receiver chan interface{}

	Producer    streamer.StreamProducer
	name        string
	subscribeTo []string
	workerThd   IWorker
	broker      *streamer.StatefulStreamer
}

func NewActor(name string, broker *streamer.StatefulStreamer, subscribeTo []string, publishTo []string, bufferLen []int, workerThd IWorker) *Actor {
	actor := &Actor{
		receiver: make(chan interface{}),
		sender: &MessageWrapper{
			MsgBroker:      broker,
			WorkThreadName: name,
		},
		name:        name,
		subscribeTo: subscribeTo,
		broker:      broker,
	}
	if publishTo != nil || len(publishTo) > 0 {
		actor.Producer = streamer.NewDefaultProducer(name+"-producer", publishTo, bufferLen)
	}

	actor.SetWorker(workerThd)
	workerThd.Init(name, broker)
	workerThd.OnStart()

	for _, subscribe := range subscribeTo {
		log.Metas.Add(name, "sm_"+subscribe, false)
	}
	for _, publish := range publishTo {
		log.Metas.Add(name, "sm_"+publish, true)
	}

	go actor.Serve()
	return actor
}

func (actor *Actor) Connect(controller streamer.StreamController) {
	if actor.Producer != nil {
		actor.broker.RegisterProducer(actor.Producer)
	}

	actor.broker.RegisterConsumer(streamer.NewDefaultConsumer(actor.name+"-consumer", actor.subscribeTo, controller))
}

// SetWorkUint set one uint for work
func (actor *Actor) SetWorker(worker IWorker) {
	actor.workerThd = worker
}

func (actor *Actor) Consume(data interface{}) {
	actor.receiver <- data
}

func (actor *Actor) Serve() {
	for v := range actor.receiver {
		msgs := []*Message{}
		switch v.(type) {
		case []interface{}:
			params := v.([]interface{})
			for _, p := range params {
				actor.parseParams(&msgs, p)
			}
		case streamer.Aggregated:
			param := v.(streamer.Aggregated)
			actor.parseParams(&msgs, param.Data)
		}

		idx := actor.findMaxHeight(&msgs)
		refid := actor.log(&msgs)
		lstMessage := msgs[idx].CopyHeader()
		lstMessage.Msgid = refid
		actor.workerThd.ChangeEnvironment(lstMessage)
		actor.workerThd.OnMessageArrived(msgs)
	}

}

func (actor *Actor) log(msgs *[]*Message) uint64 {
	workthreadname := actor.name
	var latestMsg *Message
	source := ""
	for _, v := range *msgs {
		latestMsg = v
		log.Logger.AddLog(
			0,
			log.LogLevel_Info,
			v.Name,
			workthreadname,
			"received msg "+v.Name,
			"msg",
			v.Height,
			v.Round,
			v.Msgid,
			0,
		)
		source = source + "," + v.Name
	}

	interRefId := log.Logger.AddLog(
		0,
		log.LogLevel_Info,
		source[1:],
		workthreadname,
		fmt.Sprintf("%d messages enter workthread %v", len(*msgs), workthreadname),
		log.LogType_Act,
		latestMsg.Height,
		latestMsg.Round,
		latestMsg.Msgid,
		0,
	)
	return interRefId
}

func (actor *Actor) findMaxHeight(msgs *[]*Message) int {
	heightPlus := uint64(0)
	roundPlus := uint64(0)
	idx := 0
	for i, msg := range *msgs {
		if msg.Height > heightPlus || (msg.Height == heightPlus && msg.Round > roundPlus) {
			heightPlus = msg.Height
			roundPlus = msg.Round
			idx = i
		}
	}
	return idx
}

func (actor *Actor) parseParams(msgs *[]*Message, data interface{}) {
	switch data.(type) {
	case []interface{}:
		pparams := data.([]interface{})
		for _, pp := range pparams {
			*msgs = append(*msgs, pp.(*Message))
		}
	case streamer.Aggregated:
		param := data.(streamer.Aggregated)
		switch param.Data.(type) {
		case streamer.Aggregated:
			pparam := param.Data.(streamer.Aggregated)

			*msgs = append(*msgs, pparam.Data.(*Message))
		case []interface{}:
			pparams := param.Data.([]interface{})
			for _, pp := range pparams {
				*msgs = append(*msgs, pp.(*Message))
			}
		}
	case *Message:
		*msgs = append(*msgs, data.(*Message))
	}
}
