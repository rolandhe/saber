// rpc abstraction basing nfour
// Copyright 2023 The saber Authors. All rights reserved.

// Package rpc 基于 duplex基础之上实现的高性能的rpc框架，它操作的是接口和业务对象数据，相比基于二进制的duplex有更好的抽象，更容易规范化。
// rpc框架的底层仍然是二进制通信，因此它抽象了编解码层。
package rpc

import (
	"errors"
	"github.com/rolandhe/saber/nfour"
	"sync"
)

var (
	badReqErr = errors.New("bad request")
)

// SrvCodec rpc服务端编解码抽象
type SrvCodec[REQ any, RES any] interface {
	// Decode 解码二进制为业务对象
	Decode(payload []byte) (*REQ, error)
	// Encode 编码业务对象为二进制
	Encode(res *RES) ([]byte, error)
}

// HandleBiz 用于处理业务请求的函数，它能够接收业务请求对象，然后进行业务处理，返回响应业务对象
// 与 nfour.WorkingFunc 不同的是 HandleBiz 处理的是业务对象，nfour.WorkingFunc 处理的是二进制，
type HandleBiz[REQ any, RES any] func(req *REQ) (*RES, error)

// HandleErrorFunc 转换err为需要返回的业务响应对象
type HandleErrorFunc[RES any] func(err error, interfaceName any) *RES

// NewRpcWorking 基于业务对象构建处底层需要的 nfour.WorkingFunc
//
// codec 对象的编解码工具
//
// kExtractor rpc 方法名称抽取工具。rpc的是在方法级别被调用的，每个方法在服务端对应一个业务处理函数，服务端需要根据这个 "方法名称" 来路由到方法处理函数，"方法名称"会在请求中设置，kExtractor 负责从请求对象中提取"方法名称"
//
// 返回值 *SrvRouter ， 返回一个服务端的路由对象，它能帮助正确的路由到方法出来函数
func NewRpcWorking[REQ any, RES any](codec SrvCodec[REQ, RES], kExtractor func(req *REQ) any, errToRes HandleErrorFunc[RES]) (nfour.WorkingFunc, *SrvRouter[REQ, RES]) {
	router := &SrvRouter[REQ, RES]{
		codec:        codec,
		keyExtractor: kExtractor,
		errorToRes:   errToRes,
	}
	return func(task *nfour.Task) ([]byte, error) {
		payload := task.PayLoad
		return workingCore(router, payload)
	}, router
}

func workingCore[REQ any, RES any](router *SrvRouter[REQ, RES], payload []byte) ([]byte, error) {
	req, err := router.codec.Decode(payload)
	if err != nil {
		nfour.NFourLogger.InfoLn(err)
		response, _ := router.codec.Encode(router.errorToRes(err, nil))
		return response, nil
	}
	return router.run(req), nil
}

// SrvRouter 服务端的方法路由器，它包含了编解码工具，方法注册表，方法名称提取工具等。
// 服务端需要把方法名称及其对应的方法处理函数注册到 SrvRouter , 当请求进入时，能够使用方法提取工具从请求中提取到方法名称，并正确的找到方法处理函数，然后执行
type SrvRouter[REQ any, RES any] struct {
	codec        SrvCodec[REQ, RES]
	regTable     sync.Map
	keyExtractor func(req *REQ) any
	errorToRes   HandleErrorFunc[RES]
}

// Register 注册方法名称及方法处理函数
// 如果相同的方法名称注册多个函数，最后一个会覆盖前面的，并输出日志
func (r *SrvRouter[REQ, RES]) Register(key any, fn HandleBiz[REQ, RES]) {
	if _, loaded := r.regTable.LoadOrStore(key, fn); loaded {
		nfour.NFourLogger.Info("%v exists\n", key)
	}
}

func (r *SrvRouter[REQ, RES]) run(req *REQ) []byte {
	key := r.keyExtractor(req)
	if key == nil {
		return r.handleErr(badReqErr, nil)
	}
	v, ok := r.regTable.Load(key)
	if !ok {
		return r.handleErr(badReqErr, key)
	}
	fn := v.(HandleBiz[REQ, RES])
	res, err := fn(req)
	if err != nil {
		return r.handleErr(err, key)
	}
	buff, _ := r.codec.Encode(res)
	return buff
}

func (r *SrvRouter[REQ, RES]) handleErr(err error, interfaceName any) []byte {
	buf, _ := r.codec.Encode(r.errorToRes(err, interfaceName))
	return buf
}
