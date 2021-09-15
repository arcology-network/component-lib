// +build !CI

package lib

import (
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	ethCommon "github.com/arcology-network/3rd-party/eth/common"
	tmCommon "github.com/arcology-network/3rd-party/tm/common"
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"github.com/spf13/viper"
)

func Test_SendPack(t *testing.T) {
	addr := "localhost:9092"

	relations := map[string]string{}
	relations["tx"] = "log"
	relations["noderole"] = "msgexch"
	logcfg := viper.GetString("logcfg")
	log.InitLog("test.log", logcfg, "unitTest", "node1", 1)

	outgoing := new(ComOutgoing)
	if err := outgoing.Start([]string{addr}, relations, "uploader"); err != nil {
		panic(err)
	}

	msg := &actor.Message{
		Msgid:  99999,
		Name:   "tx",
		Height: 21,
		Round:  1,
		Data:   &[]byte{123, 43, 67, 98},
	}
	fmt.Printf("send start time %v\n", time.Now())
	outgoing.Send(msg)

	gob.Register(types.ParentInfo{}) //encode must register this type

	msg = &actor.Message{
		Msgid:  8888,
		Name:   "noderole",
		Height: 22,
		Round:  2,
		Data: &types.ParentInfo{
			ParentHash: ethCommon.BytesToHash([]byte{1, 2, 3, 4}),
			ParentRoot: ethCommon.BytesToHash([]byte{5, 62, 7, 48}),
		},
	}

	fmt.Printf("send start time %v\n", time.Now())
	outgoing.Send(msg)

	//pdcTx.Stop()
	tmCommon.TrapSignal(func() {
		// Cleanup
		//svr.Stop()
		//csmTx.Stop()
	})
}

func Test_received(t *testing.T) {
	addr := "localhost:9092"
	gob.Register(types.ParentInfo{}) //encode must register this type
	logcfg := viper.GetString("logcfg")
	log.InitLog("test.log", logcfg, "unitTest", "node1", 1)
	//
	csmTx := new(ComIncoming)

	csmTx.Start([]string{addr}, []string{"log", "msgexch"}, []string{"tx", "noderole"}, "groupids", "downloader", receiveFinish)

	tmCommon.TrapSignal(func() {
		// Cleanup
		//svr.Stop()
		//csmTx.Stop()
	})
}

//callback when all finaish
func receiveFinish(msg *actor.Message) error {
	fmt.Printf("received msg %v\n", *msg)

	return nil
}

func TestSendRecv(t *testing.T) {
	n := 100000
	l := 100
	uploader := &ComOutgoing{}
	uploader.Start([]string{"localhost:9092"}, map[string]string{actor.MsgTxBlocks: "topic"}, "uploader")

	done := make(chan bool, 0)
	count := 0
	downloader := &ComIncoming{}
	downloader.Start([]string{"localhost:9092"}, []string{"topic"}, []string{actor.MsgTxBlocks}, "tester", "downloader", func(msg *actor.Message) error {
		count++
		if count == n {
			close(done)
		}
		return nil
	})

	// b.ResetTimer()
	begin := time.Now()
	for i := 0; i < n; i++ {
		uploader.Send(&actor.Message{
			Name: actor.MsgTxBlocks,
			Data: make([]byte, l),
		})
	}

	<-done
	t.Log(time.Now().Sub(begin))
	// uploader.Stop()
	downloader.Stop()
}
