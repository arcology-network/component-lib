package rpc

import (
	"time"

	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"

	"github.com/smallnest/rpcx/client"
	rpcxserver "github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

const (
	servicePath = "backend.service"
)

func InitZookeeperRpcClient(basepath string, zkAddrs []string) client.XClient {
	share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
	opt := client.DefaultOption
	opt.SerializeType = protocol.SerializeType(4)

	d := client.NewZookeeperDiscovery("/"+basepath, servicePath, zkAddrs, nil)
	return client.NewXClient(servicePath, client.Failover, client.RoundRobin, d, opt)
}

func InitZookeeperRpcServer(serviceAddr, basepath string, zkAddrs []string, rcvrs, fns []interface{}) {
	go func() {
		// rpcx service
		share.Codecs[protocol.SerializeType(4)] = &GobCodec{}
		rpcxServer := rpcxserver.NewServer()
		register := &serverplugin.ZooKeeperRegisterPlugin{
			ServiceAddress:   "tcp@" + serviceAddr,
			ZooKeeperServers: zkAddrs,
			BasePath:         "/" + basepath,
			UpdateInterval:   time.Minute,
		}
		register.Start()
		rpcxServer.Plugins.Add(register)
		if rcvrs != nil {
			for _, svc := range rcvrs {
				rpcxServer.RegisterName(servicePath, svc, "")
			}
		}
		if fns != nil {
			for _, fn := range fns {
				rpcxServer.RegisterFunction(servicePath, fn, "")
			}
		}

		rpcxServer.Serve("tcp", serviceAddr)
	}()
}
