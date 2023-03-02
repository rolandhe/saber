// Package nfour, net framework basing tcp, tcp is 4th layer of osi net model
//
// Copyright 2023 The saber Authors. All rights reserved.

package duplex

import (
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/utils/bytutil"
	"net"
	"strconv"
	"time"
)

const fullHeaderLength = 12

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
	header := make([]byte, fullHeaderLength)
	for {
		conn.SetReadDeadline(time.Now().Add(conf.IdleTimeout))
		err := nfour.ReadPayload(conn, header, fullHeaderLength, true)
		if err != nil {
			close(closeCh)
			close(writeCh)
			break
		}
		l, _ := bytutil.ToInt32(header[:nfour.PayLoadLenBufLength])
		bodyBuff := make([]byte, l, l)
		conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout))
		err = nfour.ReadPayload(conn, bodyBuff, int(l), false)
		if err != nil {
			close(closeCh)
			close(writeCh)
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
			break
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
			break
		}
	}
}

func writeCore(res []byte, seqId uint64, conn net.Conn, timeout time.Duration) bool {
	conn.SetWriteDeadline(time.Now().Add(timeout))

	plen := len(res)
	payload := make([]byte, plen+fullHeaderLength)
	copy(payload, bytutil.Int32ToBytes(int32(plen)))
	copy(payload[nfour.PayLoadLenBufLength:], bytutil.Uint64ToBytes(seqId))
	copy(payload[fullHeaderLength:], res)

	n, err := conn.Write(payload)
	if err != nil {
		conn.Close()
		nfour.NFourLogger.InfoLn(err)
		return false
	}
	nfour.NFourLogger.Debug("write data:%d, expect:%d\n", n, plen+fullHeaderLength)
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
