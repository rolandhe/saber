// Package nfour, net framework basing tcp, tcp is 4th layer of osi net model
//
// Copyright 2023 The saber Authors. All rights reserved.

package duplex

import (
	"errors"
	"github.com/rolandhe/saber/gocc"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/utils/bytutil"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TaskTimeout   = errors.New("timeout")
	TransShutdown = errors.New("transport shut down")
)

type TransConf struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	concurrent   gocc.Semaphore
}

type ReqTimeout struct {
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	WaitConcurrent time.Duration
}

func NewTransConf(rwTimeout time.Duration, concurrent uint) *TransConf {
	return &TransConf{
		ReadTimeout:  rwTimeout,
		WriteTimeout: rwTimeout,
		IdleTimeout:  time.Minute * 30,
		concurrent:   gocc.NewDefaultSemaphore(concurrent),
	}
}

func NewTrans(addr string, conf *TransConf) (*Trans, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// handle error
		nfour.NFourLogger.InfoLn(err)
		return nil, err
	}

	t := &Trans{
		conn:     conn,
		conf:     conf,
		sendCh:   make(chan *sendingTask, conf.concurrent.TotalTokens()),
		shutDown: make(chan struct{}),
	}

	go asyncSender(t)
	go asyncReader(t)

	return t, nil
}

type Trans struct {
	conn     net.Conn
	conf     *TransConf
	sendCh   chan *sendingTask
	shutDown chan struct{}
	status   int32
	cache    sync.Map
	idGen    atomic.Uint64
}

func (t *Trans) Shutdown() {
	if atomic.CompareAndSwapInt32(&t.status, 0, 1) {
		close(t.shutDown)
	}
}

func (t *Trans) IsShutdown() bool {
	return atomic.LoadInt32(&t.status) == 1
}

func (t *Trans) SendPayload(req []byte, reqTimeout *ReqTimeout) ([]byte, error) {
	if t.IsShutdown() {
		return nil, TransShutdown
	}
	if reqTimeout == nil {
		reqTimeout = &ReqTimeout{}
	}
	if !t.conf.concurrent.AcquireTimeout(reqTimeout.WaitConcurrent) {
		return nil, nfour.ExceedConcurrentError
	}
	if reqTimeout.WriteTimeout <= 0 {
		reqTimeout.WriteTimeout = t.conf.WriteTimeout
	}
	if reqTimeout.ReadTimeout <= 0 {
		reqTimeout.ReadTimeout = t.conf.ReadTimeout
	}
	if t.IsShutdown() {
		return nil, TransShutdown
	}
	seqId := t.idGen.Add(1)
	fu := &future{
		seqId:    seqId,
		notifier: make(chan struct{}),
	}
	t.cache.Store(seqId, fu)
	t.sendCh <- &sendingTask{
		seqId:   seqId,
		payload: req,
		timeout: reqTimeout.WriteTimeout,
	}
	return fu.Get(reqTimeout.ReadTimeout)
}

// asyncSender/asyncReader以及外部都可以调用Shutdown发送关闭指令
// 但由sender 最终来关闭连接
// asyncSender识别到连接关闭指令后消除等待结果的任务
func asyncSender(trans *Trans) {
	releaseWait := false
	for {
		if trans.IsShutdown() {
			trans.conn.Close()
			releaseWait = true
			break
		}
		select {
		case task := <-trans.sendCh:
			if !writeCore(task.payload, task.seqId, trans.conn, task.timeout) {
				nfour.NFourLogger.InfoLn("write err,will shutdown")
				trans.Shutdown()
				releaseWait = true
				break
			}
			nfour.NFourLogger.Debug("send success\n")
		case <-trans.shutDown:
			trans.conn.Close()
			releaseWait = true
			nfour.NFourLogger.InfoLn("shut down")
			break
		case <-time.After(trans.conf.IdleTimeout):
			nfour.NFourLogger.InfoLn("wait send task timeout")
		}
	}
	if releaseWait {
		trans.cache.Range(func(key, value any) bool {
			fu := value.(*future)
			fu.accept(nil, TaskTimeout)
			return true
		})
	}
}

func asyncReader(trans *Trans) {
	fullHeaderLength := nfour.PayLoadLenBufLength + seqIdHeaderLength
	header := make([]byte, fullHeaderLength)

	for {
		if trans.IsShutdown() {
			break
		}
		trans.conn.SetReadDeadline(time.Now().Add(trans.conf.IdleTimeout))
		if nfour.ReadPayload(trans.conn, header, fullHeaderLength, true) != nil {
			trans.Shutdown()
			break
		}
		l, _ := bytutil.ToInt32(header[:nfour.PayLoadLenBufLength])
		bodyBuff := make([]byte, l, l)
		seqId, err := bytutil.ToUint64(header[nfour.PayLoadLenBufLength:])
		trans.conn.SetReadDeadline(time.Now().Add(trans.conf.ReadTimeout))
		err = nfour.ReadPayload(trans.conn, bodyBuff, int(l), false)

		f, ok := trans.cache.Load(seqId)
		if !ok {
			nfour.NFourLogger.InfoLn("warning: lost seqId with read result", seqId)
			continue
		}
		if trans.IsShutdown() {
			break
		}
		trans.cache.Delete(seqId)
		fu := f.(*future)
		fu.accept(bodyBuff, err)
		trans.conf.concurrent.Release()
		if err != nil {
			trans.Shutdown()
			break
		}
	}
}

type sendingTask struct {
	seqId   uint64
	payload []byte
	timeout time.Duration
}

type future struct {
	seqId    uint64
	notifier chan struct{}
	value    []byte
	err      error
	flag     atomic.Bool
}

func (f *future) Get(timeout time.Duration) ([]byte, error) {
	select {
	case <-f.notifier:
		return f.value, f.err
	case <-time.After(timeout):
		return nil, TaskTimeout
	}
}

func (f *future) accept(v []byte, err error) {
	if f.flag.CompareAndSwap(false, true) {
		f.value = v
		f.err = err
		close(f.notifier)
	}
}
