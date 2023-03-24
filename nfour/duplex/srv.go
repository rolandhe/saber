// net framework basing tcp, tcp is 4th layer of osi net model
// Copyright 2023 The saber Authors. All rights reserved.

// Package duplex 多路复用模式的4层tcp通信，同时支持服务端和客户端。多路复用可以使用一条tcp连接并发响应多个请求，极少的使用资源。类似于http2.
// 利用多路复用模式可以利用极少资源实现高性能的网络通信功能。
package duplex

import (
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/utils/bytutil"
	"net"
	"strconv"
	"time"
)

const seqIdHeaderLength = 8

// Startup 启动一个多路复用的服务端，在多路复用模式下，每个连接由两个goroutine服务，一个负责读取请求，另一个负责写出响应，但一个读取goroutine可以持续的从连接中读取请求，
// 而没有必要等待上一个请求完成，多个请求可以并发的被执行，最终这些结果被负责写的goroutine写出。
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
		handleConnection(conn, conf.GetConcurrent().TotalTokens(), conf)
	}
}

func handleConnection(conn net.Conn, limitPerConn uint, conf *nfour.SrvConf) {
	writeCh := make(chan *result, limitPerConn)
	closeCh := make(chan struct{})
	go readConn(conn, writeCh, closeCh, conf)
	go writeConn(conn, writeCh, closeCh, conf)
}

func readConn(conn net.Conn, writeCh chan *result, closeCh chan struct{}, conf *nfour.SrvConf) {
	nfour.NFourLogger.DebugLn("start to read header info...")
	fullHeaderLength := seqIdHeaderLength + nfour.PayLoadLenBufLength
	header := make([]byte, fullHeaderLength)
	for {
		conn.SetReadDeadline(time.Now().Add(conf.IdleTimeout))
		err := nfour.InternalReadPayload(conn, header, fullHeaderLength, true)
		if err != nil {
			nfour.NFourLogger.InfoLn("read header error")
			close(closeCh)
			close(writeCh)
			break
		}
		l, _ := bytutil.ToInt32(header[:nfour.PayLoadLenBufLength])
		bodyBuff := make([]byte, l, l)
		conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout))
		err = nfour.InternalReadPayload(conn, bodyBuff, int(l), false)
		if err != nil {
			close(closeCh)
			close(writeCh)
			nfour.NFourLogger.Info("read payload error,need %d bytes\n", l)
			break
		}
		seqId, _ := bytutil.ToUint64(header[nfour.PayLoadLenBufLength:])
		if !conf.GetConcurrent().AcquireTimeout(conf.SemaWaitTime) {
			writeCh <- &result{true, seqId, conf.ErrHandle(nfour.ExceedConcurrentError)}
			continue
		}
		go doBiz(bodyBuff, writeCh, conf, seqId)
	}
}

func doBiz(bodyBuff []byte, writeCh chan *result, conf *nfour.SrvConf, seqId uint64) {
	task := &nfour.Task{PayLoad: bodyBuff}
	resBody, err := conf.Working(task)

	if err != nil {
		resBody = conf.ErrHandle(err)
	}
	writeCh <- &result{false, seqId, resBody}
}

// readConn 感知是否需要关闭连接后，通过closeCh来通知,writeConn得到消息后最终关闭连接
// writeConn识别到网络失败，关闭连接后，等待readConn感知到后再通知writeConn是否资源
func writeConn(conn net.Conn, writeCh chan *result, closeCh chan struct{}, conf *nfour.SrvConf) {
	writeCloseConn := false
	for {
		if isClose(closeCh) {
			if !writeCloseConn {
				conn.Close()
			}
			return
		}
		select {
		case res := <-writeCh:
			if !writeCore(res.ret, res.seqId, conn, conf.WriteTimeout) {
				writeCloseConn = true
				continue
			}
			// quickFailed=true代表没有执行也操作,直接返回超出并发错误,因此不需要释放信号量
			if !res.quickFailed {
				conf.GetConcurrent().Release()
			}
		case <-closeCh:
			if !writeCloseConn {
				conn.Close()
			}
			return
		}
	}
}

func writeCore(res []byte, seqId uint64, conn net.Conn, timeout time.Duration) bool {
	conn.SetWriteDeadline(time.Now().Add(timeout))

	fullHeaderLength := nfour.PayLoadLenBufLength + seqIdHeaderLength

	plen := len(res)
	allSize := plen + fullHeaderLength
	payload := make([]byte, allSize)
	copy(payload, bytutil.Int32ToBytes(int32(plen)))
	copy(payload[nfour.PayLoadLenBufLength:], bytutil.Uint64ToBytes(seqId))
	copy(payload[fullHeaderLength:], res)

	l := 0
	for {
		n, err := conn.Write(payload)
		if err != nil {
			conn.Close()
			nfour.NFourLogger.InfoLn(err, "write core failed")
			return false
		}
		l += n
		if l == allSize {
			break
		}
		payload = payload[l:]
	}

	nfour.NFourLogger.Debug("write data:%d, expect:%d\n", l, allSize)
	return true
}

func isClose(closeCh chan struct{}) bool {
	select {
	case <-closeCh:
		return true
	default:
		return false
	}
}

type result struct {
	quickFailed bool
	seqId       uint64
	ret         []byte
}
