// Package rpc, rpc abstraction basing nfour
//
// Copyright 2023 The saber Authors. All rights reserved.

package rpc

import (
	"github.com/rolandhe/saber/nfour/duplex"
)

type Client[REQ any, RES any] struct {
	codec ClientCodec[REQ, RES]
	trans *duplex.Trans
}

func NewClient[REQ any, RES any](codec ClientCodec[REQ, RES], cli *duplex.Trans) *Client[REQ, RES] {
	return &Client[REQ, RES]{
		codec: codec,
		trans: cli,
	}
}

type ClientCodec[REQ any, RES any] interface {
	Decode(payload []byte) (*RES, error)
	Encode(req *REQ) ([]byte, error)
}

func (c *Client[REQ, RES]) SendRequest(req *REQ, reqTimeout *duplex.ReqTimeout) (*RES, error) {
	payload, err := c.codec.Encode(req)
	if err != nil {
		return nil, err
	}
	resBuff, err := c.trans.SendPayload(payload, reqTimeout)
	if err != nil {
		return nil, err
	}
	res, err := c.codec.Decode(resBuff)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client[REQ, RES]) Shutdown(source string) {
	c.trans.Shutdown(source)
}
