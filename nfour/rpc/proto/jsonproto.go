// rpc implementation basing rpc abstraction
// Copyright 2023 The saber Authors. All rights reserved.

package proto

import (
	"encoding/json"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/duplex"
	"github.com/rolandhe/saber/nfour/rpc"
)

func NewJsonRpcSrvWorking(errToRes rpc.HandleErrorFunc[JsonProtoRes]) (nfour.Working, *rpc.SrvRouter[JsonProtoReq, JsonProtoRes]) {
	codec := &jsonSerCodec[JsonProtoReq, JsonProtoRes]{}
	kExtractor := func(req *JsonProtoReq) any {
		return req.Key
	}
	return rpc.NewRpcWorking[JsonProtoReq, JsonProtoRes](codec, kExtractor, errToRes)
}

type JsonClient interface {
	SendRequest(req *JsonProtoReq, reqTimeout *duplex.ReqTimeout) (*JsonProtoRes, error)
	Shutdown(source string)
}

func NewJsonRpcClient(trans *duplex.Trans) JsonClient {
	codec := &jsonClientCodec[JsonProtoReq, JsonProtoRes]{}
	c := rpc.NewClient[JsonProtoReq, JsonProtoRes](codec, trans)
	return c
}

type JsonProtoReq struct {
	Key  string `json:"key"`
	Body []byte `json:"body"`
}

type JsonProtoRes JsonProtoReq

type JsonHandleBiz[T any, V any] func(tIns *T) (*V, error)

func FactoryHandleBiz[T any, V any](handle JsonHandleBiz[T, V]) rpc.HandleBiz[JsonProtoReq, JsonProtoRes] {
	return func(req *JsonProtoReq) (*JsonProtoRes, error) {
		tIns := new(T)
		err := json.Unmarshal(req.Body, tIns)
		if err != nil {
			return nil, err
		}

		v, err := handle(tIns)
		if err != nil {
			return nil, err
		}
		jv, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		res := &JsonProtoRes{
			Key:  req.Key,
			Body: jv,
		}
		return res, nil
	}
}

type JsonSameTypeHandleBiz[T any] func(tIns *T) (*T, error)

func FactorySameTypeHandleBiz[T any](handle JsonSameTypeHandleBiz[T]) rpc.HandleBiz[JsonProtoReq, JsonProtoRes] {
	return func(req *JsonProtoReq) (*JsonProtoRes, error) {
		tIns := new(T)
		err := json.Unmarshal(req.Body, tIns)
		if err != nil {
			return nil, err
		}

		v, err := handle(tIns)
		if err != nil {
			return nil, err
		}
		jv, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		res := &JsonProtoRes{
			Key:  req.Key,
			Body: jv,
		}
		return res, nil
	}
}

type JsonStringTypeHandleBiz func(s string) (string, error)

func FactoryStringTypeHandleBiz(handle JsonStringTypeHandleBiz) rpc.HandleBiz[JsonProtoReq, JsonProtoRes] {
	return func(req *JsonProtoReq) (*JsonProtoRes, error) {
		tIns := string(req.Body)
		v, err := handle(tIns)
		if err != nil {
			return nil, err
		}

		res := &JsonProtoRes{
			Key:  req.Key,
			Body: []byte(v),
		}
		return res, nil
	}
}

func ParseJsonProtoRes[T any](res *JsonProtoRes, factory func() *T) (*T, error) {
	tIns := factory()
	if err := json.Unmarshal(res.Body, tIns); err != nil {
		return nil, err
	} else {
		return tIns, nil
	}
}

func ParseStringValueJsonProtoRes(res *JsonProtoRes) (string, error) {
	return string(res.Body), nil
}

type jsonClientCodec[REQ JsonProtoReq, RES JsonProtoRes] struct {
}

func (jc *jsonClientCodec[JsonProto, JsonProtoRes]) Decode(payload []byte) (*JsonProtoRes, error) {
	o := new(JsonProtoRes)
	if err := json.Unmarshal(payload, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (jc *jsonClientCodec[JsonProto, JsonProtoRes]) Encode(req *JsonProto) ([]byte, error) {
	return json.Marshal(req)
}

type jsonSerCodec[REQ JsonProtoReq, RES JsonProtoRes] struct {
}

func (jc *jsonSerCodec[JsonProto, JsonProtoRes]) Decode(payload []byte) (*JsonProto, error) {
	o := new(JsonProto)
	if err := json.Unmarshal(payload, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (jc *jsonSerCodec[JsonProto, JsonProtoRes]) Encode(req *JsonProtoRes) ([]byte, error) {
	return json.Marshal(req)
}
