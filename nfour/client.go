package nfour

import (
	"errors"
	"github.com/rolandhe/saber/gocc"
	"github.com/rolandhe/saber/utils/bytutil"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TaskTimeout    = errors.New("timeout")
	ClientShutdown = errors.New("client shut down")
)

type ClientConf struct {
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

func NewClientConf(rwTimeout time.Duration, concurrent uint) *ClientConf {
	return &ClientConf{
		ReadTimeout:  rwTimeout,
		WriteTimeout: rwTimeout,
		IdleTimeout:  time.Minute * 30,
		concurrent:   gocc.NewDefaultSemaphore(concurrent),
	}
}

func NewClient(addr string, conf *ClientConf) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// handle error
		log.Println(err)
		return nil, err
	}

	c := &Client{
		conn:     conn,
		conf:     conf,
		sendCh:   make(chan *sendingTask, conf.concurrent.TotalTokens()),
		shutDown: make(chan struct{}),
	}

	go asyncSender(c)
	go asyncReader(c)

	return c, nil
}

type Client struct {
	conn     net.Conn
	conf     *ClientConf
	sendCh   chan *sendingTask
	shutDown chan struct{}
	status   int32
	cache    sync.Map
	idGen    atomic.Uint64
}

func (c *Client) Shutdown() {
	if atomic.CompareAndSwapInt32(&c.status, 0, 1) {
		close(c.shutDown)
	}
}

func (c *Client) IsShutdown() bool {
	return atomic.LoadInt32(&c.status) == 1
}

func (c *Client) SendPayload(req []byte, reqTimeout *ReqTimeout) ([]byte, error) {
	if c.IsShutdown() {
		return nil, ClientShutdown
	}
	if reqTimeout == nil {
		reqTimeout = &ReqTimeout{}
	}
	if !c.conf.concurrent.AcquireTimeout(reqTimeout.WaitConcurrent) {
		return nil, ExceedConcurrentError
	}
	if reqTimeout.WriteTimeout <= 0 {
		reqTimeout.WriteTimeout = c.conf.WriteTimeout
	}
	if reqTimeout.ReadTimeout <= 0 {
		reqTimeout.ReadTimeout = c.conf.ReadTimeout
	}
	if c.IsShutdown() {
		return nil, ClientShutdown
	}
	seqId := c.idGen.Add(1)
	fu := &future{
		seqId:    seqId,
		notifier: make(chan struct{}),
	}
	c.cache.Store(seqId, fu)
	c.sendCh <- &sendingTask{
		seqId:   seqId,
		payload: req,
		timeout: reqTimeout.WriteTimeout,
	}
	return fu.Get(reqTimeout.ReadTimeout)
}

// asyncSender/asyncReader以及外部都可以调用Shutdown发送关闭指令
// 但由sender 最终来关闭连接
// asyncSender识别到连接关闭指令后消除等待结果的任务
func asyncSender(c *Client) {
	releaseWait := false
	for {
		if c.IsShutdown() {
			c.conn.Close()
			releaseWait = true
			break
		}
		select {
		case task := <-c.sendCh:
			if !writeCore(task.payload, task.seqId, c.conn, task.timeout) {
				c.Shutdown()
				releaseWait = true
				break
			}
			log.Printf("send success\n")
		case <-c.shutDown:
			c.conn.Close()
			releaseWait = true
			log.Println("shut down")
			break
		case <-time.After(c.conf.IdleTimeout):
			log.Println("wait send task timeout")
		}
	}
	if releaseWait {
		c.cache.Range(func(key, value any) bool {
			fu := value.(*future)
			fu.accept(nil, TaskTimeout)
			return true
		})
	}
}

func asyncReader(c *Client) {
	header := make([]byte, 12)

	for {
		if c.IsShutdown() {
			break
		}
		c.conn.SetReadDeadline(time.Now().Add(c.conf.IdleTimeout))
		if readPayload(c.conn, header, 12, true) != nil {
			c.Shutdown()
			break
		}
		l, _ := bytutil.ToInt32(header[:4])
		bodyBuff := make([]byte, l, l)
		seqId, err := bytutil.ToUInt64(header[4:])
		c.conn.SetReadDeadline(time.Now().Add(c.conf.ReadTimeout))
		err = readPayload(c.conn, bodyBuff, int(l), false)

		f, ok := c.cache.Load(seqId)
		if !ok {
			log.Println("warning: lost seqId with read result", seqId)
			continue
		}
		if c.IsShutdown() {
			break
		}
		c.cache.Delete(seqId)
		fu := f.(*future)
		fu.accept(bodyBuff, err)
		c.conf.concurrent.Release()
		if err != nil {
			c.Shutdown()
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
