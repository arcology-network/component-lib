package rpc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

var service interface{}

func InitRPCServer(serviceAddr, basepath string, zkAddrs []string, rcvrs, fns []interface{}) {
	t.Log("InitRPCServer")
	service = rcvrs[0]
}

func Call(f string, args, reply interface{}) {
	method := reflect.ValueOf(service).MethodByName(f)
	if method.Type().NumIn() != 3 {
		panic(fmt.Sprintf("expected 3 arguments, got %d", method.Type().NumIn()))
	}

	method.Call([]reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(args), reflect.ValueOf(reply)})
}

var t testing.TB

func NewRPCServerInitializer(logger testing.TB) func(string, string, []string, []interface{}, []interface{}) {
	t = logger
	return InitRPCServer
}
