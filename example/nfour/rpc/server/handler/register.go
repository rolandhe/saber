package handler

import (
	"encoding/json"
	"github.com/rolandhe/saber/nfour/rpc"
	"github.com/rolandhe/saber/nfour/rpc/proto"
)

func RegisterAll(router *rpc.SrvRouter[proto.JsonProtoReq, proto.JsonProtoRes]) {
	router.Register("rpc.test", func(req *proto.JsonProtoReq) (*proto.JsonProtoRes, error) {

		rpcResult, _ := json.Marshal(&Result[string]{
			Code: 200,
			Data: string(req.Body),
		})
		res := &proto.JsonProtoRes{
			Key:  req.Key,
			Body: rpcResult,
		}
		return res, nil
	})
}
