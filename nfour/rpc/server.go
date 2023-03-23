// rpc abstraction basing nfour
// Copyright 2023 The saber Authors. All rights reserved.

package rpc

import (
	"errors"
	"github.com/rolandhe/saber/nfour"
	"sync"
)

var (
	badReqErr = errors.New("bad request")
)

type SrvCodec[REQ any, RES any] interface {
	Decode(payload []byte) (*REQ, error)
	Encode(res *RES) ([]byte, error)
}

type HandleBiz[REQ any, RES any] func(req *REQ) (*RES, error)
type HandleErrorFunc[RES any] func(err error, interfaceName any) *RES

func NewRpcWorking[REQ any, RES any](codec SrvCodec[REQ, RES], kExtractor func(req *REQ) any, errToRes HandleErrorFunc[RES]) (nfour.Working, *SrvRouter[REQ, RES]) {
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

type SrvRouter[REQ any, RES any] struct {
	codec        SrvCodec[REQ, RES]
	regTable     sync.Map
	keyExtractor func(req *REQ) any
	errorToRes   HandleErrorFunc[RES]
}

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
