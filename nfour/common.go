// tcp network framework
//
// Copyright 2023 The saber Authors. All rights reserved.

// Package nfour, tcp 网路工具框架，因为 tcp工作在4层上，所以该包命名为 nfour。包含:
//
// 1. 多路复用网络模型实现
//
// 2. 单路复用网络模型实现，即 同步request/response模型
//
// 3. 基于多路复用的rpc框架
//
// 4. json encode的多路复用rpc框架
package nfour

import (
	"errors"
	"github.com/rolandhe/saber/gocc"
	"io"
	"net"
	"os"
	"time"
)

const (
	PayLoadLenBufLength = 4
)

var (
	PeerCloseError        = errors.New("peer closed")
	ExceedConcurrentError = errors.New("exceed concurrent")
	defaultSemaWaitTime   = time.Millisecond
)

type Task struct {
	PayLoad []byte
}

type Working func(task *Task) ([]byte, error)
type HandleError func(err error) []byte

type ConnReader func(conn net.Conn, buff []byte, expectLen int, notHalt bool) error

func NewSrvConf(working Working, errHandle HandleError, concurrent uint) *SrvConf {
	return &SrvConf{
		working,
		errHandle,
		time.Millisecond * 2000,
		time.Millisecond * 2000,
		time.Minute * 10,
		defaultSemaWaitTime,
		gocc.NewDefaultSemaphore(concurrent),
	}
}

func NewSrvConfSemaWait(working Working, errHandle HandleError, concurrent uint, semaWaitTime time.Duration) *SrvConf {
	if semaWaitTime < 0 {
		semaWaitTime = defaultSemaWaitTime
	}
	return &SrvConf{
		working,
		errHandle,
		time.Millisecond * 2000,
		time.Millisecond * 2000,
		time.Minute * 10,
		semaWaitTime,
		gocc.NewDefaultSemaphore(concurrent),
	}
}

type SrvConf struct {
	Working      Working
	ErrHandle    HandleError
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	SemaWaitTime time.Duration
	concurrent   gocc.Semaphore
}

func (conf *SrvConf) GetConcurrent() gocc.Semaphore {
	return conf.concurrent
}

func InternalReadPayload(conn net.Conn, buff []byte, expectLen int, notHalt bool) error {
	l := 0
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if !notHalt && errors.Is(err, os.ErrDeadlineExceeded) {
				NFourLogger.InfoLn(err, l)
				return err
			}
			if errors.Is(err, io.EOF) {
				NFourLogger.InfoLn("peer closed")
				return PeerCloseError
			}
			return err
		}
		l += n

		if l == expectLen {
			break
		}
		buff = buff[n:]
	}
	return nil
}
