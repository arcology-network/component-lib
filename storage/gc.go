package storage

import (
	"runtime"
	"time"

	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"go.uber.org/zap"
)

type Gc struct {
	actor.WorkerThread
}

func NewGc(lanes int, groupid string) *Gc {
	gc := Gc{}
	gc.Set(lanes, groupid)
	return &gc
}

func (gc *Gc) OnStart() {
}

func (gc *Gc) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgGc:
			t := time.Now()
			runtime.GC()
			gc.AddLog(log.LogLevel_Info, "gc completed ---->", zap.Duration("time", time.Since(t)))
		}
	}
	return nil
}
