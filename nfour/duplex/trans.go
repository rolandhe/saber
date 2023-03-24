// net framework basing tcp, tcp is 4th layer of osi net model
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
	// ErrTaskTimeout 请求执行超时异常
	ErrTaskTimeout = errors.New("task execute timeout")
	// ErrTransShutdown Trans 客户端已经被关闭
	ErrTransShutdown = errors.New("transport shut down")
)

// TransConf Trans 客户端配置
type TransConf struct {
	// ReadTimeout 网络读取超时时间
	ReadTimeout time.Duration
	// WriteTimeout 网络写出超时
	WriteTimeout time.Duration

	// IdleTimeout 连接长时间没有读取到数据的超时时间，该超过该时间，系统会输出日志，没有其他的处理，不会中断连接
	IdleTimeout time.Duration
	concurrent  gocc.Semaphore
}

// ReqTimeout 请求超时信息
type ReqTimeout struct {
	// ReadTimeout 网络读取超时时间
	ReadTimeout time.Duration
	// WriteTimeout 网络写出超时
	WriteTimeout time.Duration
	// WaitConcurrent 当到达最大并发时，等待执行的超时时间
	WaitConcurrent time.Duration
}

// NewTransConf 构建客户端的配置
// rwTimeout 读写超时，这种情况下，读写超时是相同的
func NewTransConf(rwTimeout time.Duration, concurrent uint) *TransConf {
	return &TransConf{
		ReadTimeout:  rwTimeout,
		WriteTimeout: rwTimeout,
		IdleTimeout:  time.Minute * 30,
		concurrent:   gocc.NewDefaultSemaphore(concurrent),
	}
}

// NewTrans 构建客户端 Trans
// name 表示该 Trans的名称，该名称会被输出到日志中，方便发现问题
func NewTrans(addr string, conf *TransConf, name string) (*Trans, error) {
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
		name:     name,
	}

	go asyncSender(t)
	go asyncReader(t)

	return t, nil
}

// Trans 多路复用模式下的客户端，每个Trans内持有一个连接，并且与服务端类似，由两个goroutine分别负责请求的发出和响应的接收。
// 使用者通过Trans发送请求到服务端，并返回响应
type Trans struct {
	conn     net.Conn
	conf     *TransConf
	sendCh   chan *sendingTask
	shutDown chan struct{}
	status   int32
	cache    sync.Map
	idGen    atomic.Uint64
	name     string
}

// Shutdown 关闭Trans
// source 发起Shutdown的场景，用于日志记录
func (t *Trans) Shutdown(source string) {
	if atomic.CompareAndSwapInt32(&t.status, 0, 1) {
		nfour.NFourLogger.Info("%s trigger %s shutdown\n", source, t.name)
		close(t.shutDown)
	}
}

// IsShutdown Trans是否已经被关闭，如果已经被关闭，将不能接收新的发送请求
func (t *Trans) IsShutdown() bool {
	return atomic.LoadInt32(&t.status) == 1
}

// SendPayload 发送二进制请求
// reqTimeout 本次请求的超时时间
func (t *Trans) SendPayload(req []byte, reqTimeout *ReqTimeout) ([]byte, error) {
	if t.IsShutdown() {
		return nil, ErrTransShutdown
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
		return nil, ErrTransShutdown
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
		f:       fu,
	}
	return fu.get(reqTimeout.ReadTimeout)
}

// asyncSender/asyncReader以及外部都可以调用Shutdown发送关闭指令
// 但由sender 最终来关闭连接
// asyncSender识别到连接关闭指令后消除等待结果的任务
func asyncSender(trans *Trans) {
	releaseWait := false
	for {
		select {
		case task := <-trans.sendCh:
			if !writeCore(task.payload, task.seqId, trans.conn, task.timeout) {
				nfour.NFourLogger.Info("%s write err,will shutdown\n", trans.name)
				trans.Shutdown("sender")
				releaseWait = true
				break
			}
			nfour.NFourLogger.Debug("%s send success\n", trans.name)
		case <-trans.shutDown:
			trans.conn.Close()
			releaseWait = true
			nfour.NFourLogger.Info("%s get shut down event,shut down\n", trans.name)
			break
		case <-time.After(trans.conf.IdleTimeout):
			nfour.NFourLogger.Info("%s wait send task timeout\n", trans.name)
		}
		if releaseWait {
			break
		}
	}
	if releaseWait {
		nfour.NFourLogger.Info("%s send release not sent task\n", trans.name)
		releaseCount := 0
		for {
			select {
			case task := <-trans.sendCh:
				task.f.accept(nil, ErrTransShutdown)
				releaseCount++
			default:
				nfour.NFourLogger.Info("%s send release not sent task:%d\n", trans.name, releaseCount)
				return
			}
		}

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
		if err := nfour.InternalReadPayload(trans.conn, header, fullHeaderLength, true); err != nil {
			nfour.NFourLogger.Info("%s read header error:%v\n", trans.name, err)
			trans.Shutdown("reader")
			break
		}
		l, _ := bytutil.ToInt32(header[:nfour.PayLoadLenBufLength])
		bodyBuff := make([]byte, l, l)
		seqId, err := bytutil.ToUint64(header[nfour.PayLoadLenBufLength:])
		trans.conn.SetReadDeadline(time.Now().Add(trans.conf.ReadTimeout))
		err = nfour.InternalReadPayload(trans.conn, bodyBuff, int(l), false)
		if err != nil {
			nfour.NFourLogger.Info("%s read payload error:%v,need %d bytes\n", trans.name, err, l)
			trans.Shutdown("reader")
			break
		}
		f, ok := trans.cache.Load(seqId)
		if !ok {
			nfour.NFourLogger.Info("warning: %s lost seqId:%d with read result\n", trans.name, seqId)
			continue
		}
		if trans.IsShutdown() {
			break
		}
		trans.cache.Delete(seqId)
		fu := f.(*future)
		fu.accept(bodyBuff, err)
		trans.conf.concurrent.Release()
	}
	nfour.NFourLogger.Info("%s async reader release futures\n", trans.name)
	releasedCount := 0

	trans.cache.Range(func(key, value any) bool {
		fu := value.(*future)
		fu.accept(nil, ErrTransShutdown)
		releasedCount++
		return true
	})
	nfour.NFourLogger.Info("%s async reader release futures:%d\n", trans.name, releasedCount)
}

type sendingTask struct {
	seqId   uint64
	payload []byte
	timeout time.Duration
	f       *future
}

type future struct {
	seqId    uint64
	notifier chan struct{}
	value    []byte
	err      error
	flag     atomic.Bool
}

func (f *future) get(timeout time.Duration) ([]byte, error) {
	select {
	case <-f.notifier:
		return f.value, f.err
	case <-time.After(timeout):
		return nil, ErrTaskTimeout
	}
}

func (f *future) accept(v []byte, err error) {
	if f.flag.CompareAndSwap(false, true) {
		f.value = v
		f.err = err
		close(f.notifier)
	}
}
