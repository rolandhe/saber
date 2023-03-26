package main

import (
	"github.com/rolandhe/saber/example/nfour/rpc/server/handler"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/duplex"
	"github.com/rolandhe/saber/nfour/rpc/proto"
)

func main() {

	working, handlerErrFunc, router := proto.NewJsonRpcSrvWorking(handler.JsonRpcErrHandler)
	handler.RegisterAll(router)
	conf := nfour.NewSrvConf(working, handlerErrFunc, 10000)

	duplex.Startup(11011, conf)
}
