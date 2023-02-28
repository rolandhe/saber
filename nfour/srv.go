package nfour

import (
	"errors"
	"github.com/rolandhe/saber/gocc"
	"github.com/rolandhe/saber/utils/bytutil"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	peerCloseError        = errors.New("peer closed")
	ExceedConcurrentError = errors.New("exceed concurrent")
)

type Working func(task *Task) ([]byte, error)
type HandleError func(err error) []byte

func NewConf(working Working, errHandle HandleError, concurrent uint) *Conf {
	return &Conf{
		working,
		errHandle,
		time.Millisecond * 2000,
		time.Millisecond * 2000,
		time.Minute * 10,
		gocc.NewDefaultSemaphore(concurrent),
	}
}

type Conf struct {
	Working      Working
	ErrHandle    HandleError
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	concurrent   gocc.Semaphore
}

func Startup(port int, conf *Conf) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		// handle error
		log.Println(err)
		return
	}
	log.Printf("listen tcp port %d,and next to accept\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Println(err)
		}
		handleConnection(conn, conf.concurrent.TotalTokens(), conf)
	}
}

func handleConnection(conn net.Conn, limitPerConn uint, conf *Conf) {
	writeCh := make(chan *result, limitPerConn)
	closeCh := make(chan struct{})
	go readConn(conn, writeCh, closeCh, conf)
	go writeConn(conn, writeCh, closeCh, conf)
}

func readConn(conn net.Conn, writeCh chan *result, closeCh chan struct{}, conf *Conf) {
	log.Println("start to read header info...")
	header := make([]byte, 12)
	for {
		conn.SetReadDeadline(time.Now().Add(conf.IdleTimeout))
		err := readPayload(conn, header, 12, true)
		if err != nil {
			close(closeCh)
			close(writeCh)
			break
		}
		l, _ := bytutil.ToInt32(header[:4])
		bodyBuff := make([]byte, l, l)
		conn.SetReadDeadline(time.Now().Add(conf.ReadTimeout))
		err = readPayload(conn, bodyBuff, int(l), false)
		if err != nil {
			close(closeCh)
			close(writeCh)
			break
		}
		seqId, _ := bytutil.ToUInt64(header[4:])
		if !conf.concurrent.TryAcquire() {
			writeCh <- &result{true, seqId, conf.ErrHandle(ExceedConcurrentError)}
			continue
		}
		go doBiz(bodyBuff, writeCh, conf, seqId)
	}
}

func doBiz(bodyBuff []byte, writeCh chan *result, conf *Conf, seqId uint64) {
	task := &Task{bodyBuff}
	resBody, err := conf.Working(task)

	if err != nil {
		resBody = conf.ErrHandle(err)
	}
	writeCh <- &result{false, seqId, resBody}
}

// readConn 感知是否需要关闭连接后，通过closeCh来通知,writeConn得到消息后最终关闭连接
// writeConn识别到网络失败，关闭连接后，等待readConn感知到后再通知writeConn是否资源
func writeConn(conn net.Conn, writeCh chan *result, closeCh chan struct{}, conf *Conf) {
	releaseTask := false
	for {
		if isClose(closeCh) {
			conn.Close()
			releaseTask = true
			break
		}
		select {
		case res := <-writeCh:
			if !writeCore(res.ret, res.seqId, conn, conf.WriteTimeout) {
				continue
			}
			if !res.quickFailed {
				conf.concurrent.Release()
			}
		case <-closeCh:
			conn.Close()
			releaseTask = true
			break
		}
	}
	if releaseTask {
		for {
			_, ok := <-writeCh
			if !ok {
				break
			}
		}
	}
}

func writeCore(res []byte, seqId uint64, conn net.Conn, timeout time.Duration) bool {
	conn.SetWriteDeadline(time.Now().Add(timeout))

	plen := len(res)
	payload := make([]byte, plen+12)
	copy(payload, bytutil.ToBytes(int32(plen)))
	copy(payload[4:], bytutil.Uint64ToBytes(seqId))
	copy(payload[12:], res)

	n, err := conn.Write(payload)
	if err != nil {
		conn.Close()
		log.Println(err)
		return false
	}
	log.Printf("write data:%d, expect:%d\n", n, plen+12)
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

func readPayload(conn net.Conn, buff []byte, expectLen int, notHalt bool) error {
	l := 0
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if !notHalt && errors.Is(err, os.ErrDeadlineExceeded) {
				log.Println(err)
				return err
			}
			if errors.Is(err, io.EOF) {
				log.Println("peer closed")
				return peerCloseError
			}
			return err
		}
		l += n
		if l == expectLen {
			break
		}
	}
	return nil
}

type Task struct {
	PayLoad []byte
}

type result struct {
	quickFailed bool
	seqId       uint64
	ret         []byte
}
