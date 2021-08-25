package rpc

import (
	"context"
	"fmt"
	"log"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/spf13/viper"
)

const (
	net = "tcp"
)

type RpcServer struct {
	rpcSvc     *server.Server
	svcAddress string
	zkAddress  string
	basePath   string
}

func (rs *RpcServer) RegisterSvc(name string, rcvr interface{}, metadata string) {
	rs.rpcSvc.RegisterName(name, rcvr, metadata)
}

func (rs *RpcServer) StartServer() {
	rs.rpcSvc.Serve(net, rs.svcAddress)
}

var rpcServerInstance *RpcServer

func GetRpcServer() *RpcServer {
	if rpcServerInstance == nil {
		rpcServerInstance = &RpcServer{
			rpcSvc:     server.NewServer(),
			svcAddress: viper.GetString("svcaddr"),
			zkAddress:  viper.GetString("zkaddr"),
			basePath:   viper.GetString("base"),
		}

		rpcServerInstance.addRegistryPlugin(rpcServerInstance.rpcSvc)
	}
	return rpcServerInstance
}

func (rs *RpcServer) addRegistryPlugin(s *server.Server) {
	r := &serverplugin.ZooKeeperRegisterPlugin{
		ServiceAddress:   net + "@" + rs.svcAddress,
		ZooKeeperServers: []string{rs.zkAddress},
		BasePath:         rs.basePath,
		Metrics:          metrics.NewRegistry(),
		UpdateInterval:   time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}

type RpcClient struct {
	rpcClient client.XClient
	zkAddress string
	basePath  string
}

func NewClient(svcname string) *RpcClient {
	// zkAddress := viper.GetString("zkaddr")
	// basePath := viper.GetString("base")

	zkAddress := "localhost:2181"
	basePath := "/rpcx_test"

	d := client.NewZookeeperDiscovery(basePath, svcname, []string{zkAddress}, nil)
	xclient := client.NewXClient(svcname, client.Failtry, client.RoundRobin, d, client.DefaultOption)

	return &RpcClient{
		rpcClient: xclient,
		zkAddress: zkAddress,
		basePath:  basePath,
	}
}

func (rc *RpcClient) Stop() {
	rc.rpcClient.Close()
}

func (rc *RpcClient) Call(serviceMethod string, args interface{}, reply interface{}) {
	err := rc.rpcClient.Call(context.Background(), serviceMethod, args, reply)
	if err != nil {
		//log.Fatalf("failed to call: %v", err)
		fmt.Printf("err=%v\n", err)
	}
}
