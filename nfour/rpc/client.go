// rpc abstraction basing nfour
// Copyright 2023 The saber Authors. All rights reserved.

package rpc

import (
	"github.com/rolandhe/saber/nfour/duplex"
)

// Client 描述rpc的客户端
type Client[REQ any, RES any] struct {
	codec ClientCodec[REQ, RES]
	trans *duplex.Trans
}

// NewClient 构建rpc 客户端
//
// codec 请求编解码，可以把一个struct对象 编码成二进制，也可以把二进制解码成对象
func NewClient[REQ any, RES any](codec ClientCodec[REQ, RES], cli *duplex.Trans) *Client[REQ, RES] {
	return &Client[REQ, RES]{
		codec: codec,
		trans: cli,
	}
}

// ClientCodec 抽象数据的编解码，完成struct对象与二进制的转换
type ClientCodec[REQ any, RES any] interface {
	// Decode 二进制数据解码成对象
	Decode(payload []byte) (*RES, error)
	// Encode 对象编码成二进制
	Encode(req *REQ) ([]byte, error)
}

// SendRequest 发送业务对象请求并返回业务对象类型的响应值，底层通过编解码转成二进制后通过tcp发送
//
// req 业务对象类型的请求
//
// reqTimeout 超时配置
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

// Shutdown 关闭客户端，底层的Trans及连接资源会被释放
// source 关闭客户端的场景，会输出到日志，方便排除问题
func (c *Client[REQ, RES]) Shutdown(source string) {
	c.trans.Shutdown(source)
}
