package main

import (
	"github.com/rolandhe/saber/example/nfour/rpc/server/handler"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/rpc/proto"
)

func main() {

	working, router := proto.NewJsonRpcSrvWorking(handler.JsonRpcErrHandler)
	handler.RegisterAll(router)
	conf := nfour.NewSrvConf(working, handler.TransErrHandler, 2000)

	nfour.Startup(11011, conf)
}
