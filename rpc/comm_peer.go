package rpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

type RpcServerPeer struct {
	rpcSvc  *server.Server
	svcAddr string
}

func NewRpcServerPeer(svcAddr string) *RpcServerPeer {
	share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
	return &RpcServerPeer{
		rpcSvc:  server.NewServer(),
		svcAddr: svcAddr,
	}
}

func (rs *RpcServerPeer) RegisterSvc(name string, rcvr interface{}, metadata string) {
	rs.rpcSvc.RegisterName(name, rcvr, metadata)
}

func (rs *RpcServerPeer) StartServer() {
	rs.rpcSvc.Serve(net, rs.svcAddr)
}

type RpcClientPeer struct {
	rpcClient client.XClient
}

func NewClientPeer(svcAddr, svcname string) *RpcClientPeer {
	share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
	d := client.NewPeer2PeerDiscovery(net+"@"+svcAddr, "")
	opt := client.DefaultOption
	opt.SerializeType = protocol.SerializeType(4)

	xclient := client.NewXClient(svcname, client.Failtry, client.RandomSelect, d, opt)

	return &RpcClientPeer{
		rpcClient: xclient,
	}
}

func (rc *RpcClientPeer) Stop() {
	rc.rpcClient.Close()
}

func (rc *RpcClientPeer) Call(serviceMethod string, request *actor.Message, reply interface{}) {
	err := rc.rpcClient.Call(context.Background(), serviceMethod, request, reply)
	if err != nil {
		fmt.Printf("err=%v\n", err)

	}
}

type GobCodec struct {
}

func (c *GobCodec) Decode(data []byte, i interface{}) error {
	enc := gob.NewDecoder(bytes.NewBuffer(data))
	err := enc.Decode(i)
	return err
}

func (c *GobCodec) Encode(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(i)
	return buf.Bytes(), err
}
