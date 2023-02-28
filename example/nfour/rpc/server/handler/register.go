package handler

import (
	"github.com/rolandhe/saber/nfour/rpc"
	"github.com/rolandhe/saber/nfour/rpc/proto"
)

func RegisterAll(router *rpc.SrvRouter[proto.JsonProtoReq, proto.JsonProtoRes]) {
	router.Register("rpc.test", func(req *proto.JsonProtoReq) (*proto.JsonProtoRes, error) {
		res := proto.JsonProtoRes(*req)
		return &res, nil
	})
}
