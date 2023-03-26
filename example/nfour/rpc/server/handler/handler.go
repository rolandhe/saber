package handler

import (
	"encoding/json"
	"github.com/rolandhe/saber/nfour/rpc/proto"
)

type Result[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func JsonRpcErrHandler(err error, interfaceName any) *proto.JsonProtoRes {
	ret := &Result[string]{
		Code:    500,
		Message: err.Error(),
	}

	method := "can't parse request, so don't know"
	if interfaceName != nil {
		method = interfaceName.(string)
	}

	body, _ := json.Marshal(ret)
	return &proto.JsonProtoRes{
		Key:  method,
		Body: body,
	}
}
