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
	// PayLoadLenBufLength header中的首4个字节，用于记录header后面数据负载的长度
	// header格式：4个字节 + 8个字节 + n个字节的payload， 其中首4个字节里存储 n，8个字节表示request id， 最后的n个字节表示request 数据
	PayLoadLenBufLength = 4
)

var (
	// PeerCloseError 连接的另一头已经关闭连接
	PeerCloseError = errors.New("peer closed")
	// ExceedConcurrentError 当前的请求已经超出设定的最大并发数
	ExceedConcurrentError = errors.New("exceed concurrent")
	defaultSemaWaitTime   = time.Millisecond
)

// Task 描述一个请求的数据, 这个请求会被封装成Task 交于任务执行器执行
type Task struct {
	// Payload 请求数据，二进制格式，可以被上层业务解析
	PayLoad []byte
}

// WorkingFunc 请求的处理函数，请求数据会被解析，执行业务逻辑，生成业务结果，业务结果被转换成二进制格式返回
type WorkingFunc func(task *Task) ([]byte, error)

// HandleError 请求在处理过程中可能发生err， HandleError 描述一个err应该被转换成哪种返回结果
type HandleError func(err error) []byte

// NewSrvConf 构建server段的配置
// concurrent 服务的最大并发数, 当到达最大并发后，当前请求等待执行的缺省超时时间是 1 毫秒
func NewSrvConf(working WorkingFunc, errHandle HandleError, concurrent uint) *SrvConf {
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

// NewSrvConfSemaWait 构建server段的配置
//
// concurrent 服务的最大并发数, 当到达最大并发后，当前请求等待执行的超时时间由 semaWaitTime 指定
//
// semaWaitTime 当到达最大并发后，当前请求等待执行的超时时间
func NewSrvConfSemaWait(working WorkingFunc, errHandle HandleError, concurrent uint, semaWaitTime time.Duration) *SrvConf {
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

// SrvConf 描述服务端的配置， 包括:
//
//	Working 请求处理函数
//
// # ErrHandle 出错信息出来
//
// SemaWaitTime 如果当前已经到达最大并发，当前请求等待被执行的超时时间
type SrvConf struct {
	Working      WorkingFunc
	ErrHandle    HandleError
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	SemaWaitTime time.Duration
	concurrent   gocc.Semaphore
}

// GetConcurrent 获取当前服务配置的最大并发数的信号量
func (conf *SrvConf) GetConcurrent() gocc.Semaphore {
	return conf.concurrent
}

// InternalReadPayload 从连接中读取指定长度的数据， 主要是内部使用
// notHalt 当长时间读取不到数据且收到超时异常时，是不是不中断连接，true，不中断连接，继续读取
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
