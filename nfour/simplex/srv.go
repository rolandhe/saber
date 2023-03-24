// rpc abstraction basing nfour
// Copyright 2023 The saber Authors. All rights reserved.

// Package simplex 单路模式的服务端实现，类似于http1.1， 大量客户端但每个客户请求较少的场景。每个连接上request/response是同步模式，每个请求必须得到响应以后才能发送另一个请求。
// 用于兼容一些老的场景，因此该模式下没有提供客户端组件。
package simplex

import (
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/utils/bytutil"
	"net"
	"strconv"
	"time"
)

// Startup 启动一个单路服务端,在单路模式下，每个客户端连接对应服务端由一个goroutine服务，且服务是同步模式，每个请求必须被处理且收到响应后才能发送下一个请求，类似http1.1。
//
// port 服务监听的端口
// conf.concurrent 指定了最大并发数
func Startup(port int, conf *nfour.SrvConf) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		// handle error
		nfour.NFourLogger.InfoLn(err)
		return
	}
	nfour.NFourLogger.Info("listen tcp port %d,and next to accept\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			nfour.NFourLogger.InfoLn(err)
			return
		}
		go handleConnection(conn, conf)
	}
}

func handleConnection(conn net.Conn, conf *nfour.SrvConf) {
	nfour.NFourLogger.DebugLn("start to read header info...")
	header := make([]byte, nfour.PayLoadLenBufLength)
	for {
		conn.SetReadDeadline(time.Now().Add(conf.IdleTimeout))
		err := nfour.InternalReadPayload(conn, header, nfour.PayLoadLenBufLength, true)
		if err != nil {
			releaseConn(conn)
			break
		}
		l, _ := bytutil.ToInt32(header[:nfour.PayLoadLenBufLength])
		bodyBuff := make([]byte, l, l)
		conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout))
		err = nfour.InternalReadPayload(conn, bodyBuff, int(l), false)
		if err != nil {
			releaseConn(conn)
			break
		}

		if !conf.GetConcurrent().AcquireTimeout(conf.SemaWaitTime) {
			if !writeCore(conf.ErrHandle(nfour.ExceedConcurrentError), conn, conf.WriteTimeout) {
				releaseConn(conn)
				break
			}
			continue
		}
		ok := doBiz(bodyBuff, conn, conf)
		conf.GetConcurrent().Release()
		if !ok {
			releaseConn(conn)
			break
		}
	}
}

func doBiz(bodyBuff []byte, conn net.Conn, conf *nfour.SrvConf) bool {
	task := &nfour.Task{PayLoad: bodyBuff}
	resBody, err := conf.Working(task)

	if err != nil {
		resBody = conf.ErrHandle(err)
	}
	return writeCore(resBody, conn, conf.WriteTimeout)
}

func releaseConn(conn net.Conn) {
	if err := conn.Close(); err != nil {
		nfour.NFourLogger.InfoLn(err)
	}
}

func writeCore(res []byte, conn net.Conn, timeout time.Duration) bool {
	conn.SetWriteDeadline(time.Now().Add(timeout))

	plen := len(res)
	payload := make([]byte, plen+nfour.PayLoadLenBufLength)
	copy(payload, bytutil.Int32ToBytes(int32(plen)))
	copy(payload[nfour.PayLoadLenBufLength:], res)

	n, err := conn.Write(payload)
	if err != nil {
		conn.Close()
		nfour.NFourLogger.InfoLn(err)
		return false
	}
	nfour.NFourLogger.Debug("write data:%d, expect:%d\n", n, plen+nfour.PayLoadLenBufLength)
	return true
}
