// rpc implementation basing rpc abstraction
// Copyright 2023 The saber Authors. All rights reserved.

// Package proto 协议层运行在rpc与tcp之间，它提供了业务对象与底层二进制之间的转换协议。
// 当前实现的是json协议。
// 使用方式：
//
//	  working, router := proto.NewJsonRpcSrvWorking(handler.JsonRpcErrHandler)
//
//	  // 注册方法名称和业务处理函数
//	  router.Register("rpc.test", func(req *proto.JsonProtoReq) (*proto.JsonProtoRes, error) {
//
//			rpcResult, _ := json.Marshal(&Result[string]{
//				Code: 200,
//				Data: string(req.Body),
//			})
//			res := &proto.JsonProtoRes{
//				Key:  req.Key,
//				Body: rpcResult,
//			}
//			return res, nil
//		})
//
//	 conf := nfour.NewSrvConf(working, func (err error, interfaceName any) *proto.JsonProtoRes {
//	       // 处理err
//	       return &proto.JsonProtoRes {}
//	  }, 10000)
//
//		duplex.Startup(11011, conf)
package proto

import (
	"encoding/json"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/duplex"
	"github.com/rolandhe/saber/nfour/rpc"
)

// json协议实现，业务对象被封装成可以打包的json对象，经过json转换后在网络上传输

// NewJsonRpcSrvWorking 构建duplex层需要的 nfour.WorkingFunc, 编解码基于json实现
// nfour.WorkingFunc 会被设置到 nfour.SrvConf中
func NewJsonRpcSrvWorking(errToRes rpc.HandleErrorFunc[JsonProtoRes]) (nfour.WorkingFunc, nfour.HandleError, *rpc.SrvRouter[JsonProtoReq, JsonProtoRes]) {
	codec := &jsonSerCodec[JsonProtoReq, JsonProtoRes]{}
	kExtractor := func(req *JsonProtoReq) any {
		return req.Key
	}
	heFunc := func(err error) []byte {
		protoBuf := errToRes(err, nil)
		body, _ := json.Marshal(protoBuf)
		return body
	}
	wf, router := rpc.NewRpcWorking[JsonProtoReq, JsonProtoRes](codec, kExtractor, errToRes)
	return wf, heFunc, router
}

// JsonClient 基于json编解码协议的客户端抽象
type JsonClient interface {
	SendRequest(req *JsonProtoReq, reqTimeout *duplex.ReqTimeout) (*JsonProtoRes, error)
	Shutdown(source string)
}

// NewJsonRpcClient 构建JsonClient客户端
func NewJsonRpcClient(trans *duplex.Trans) JsonClient {
	codec := &jsonClientCodec[JsonProtoReq, JsonProtoRes]{}
	c := rpc.NewClient[JsonProtoReq, JsonProtoRes](codec, trans)
	return c
}

// JsonProtoReq json协议封装请求
type JsonProtoReq struct {
	// rpc方法名称
	Key string `json:"key"`
	// 请求对象被编解码协议编码成二进制
	Body []byte `json:"body"`
}

// JsonProtoRes json协议的响应对象，在json协议中req和res对象相同，但由于泛型的原因，二者命名需要不同，所以应用req的别名
type JsonProtoRes JsonProtoReq

// JsonHandleBiz json协议的业务处理函数，用于server端，它被 FactoryHandleBiz 使用
type JsonHandleBiz[T any, V any] func(tIns *T) (*V, error)

// FactoryHandleBiz 包装JsonHandleBiz 为rpc.HandleBiz
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

// JsonSameTypeHandleBiz 与 JsonHandleBiz 类似，request和response采用相同的数据类型
type JsonSameTypeHandleBiz[T any] func(tIns *T) (*T, error)

// FactorySameTypeHandleBiz 与 FactoryHandleBiz 类似
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

// ParseJsonProtoRes 解析json协议数据，解析json数据为业务对象
// factory 生成业务对象实例的回调
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
