package gocc

import (
	"sync/atomic"
	"time"
)

const sleepFixTime = time.Millisecond * 1

type Semaphore interface {
	Acquire() bool
	AcquireUntil()
	AcquireTimeout(d time.Duration) bool
	Release()
	TotalTokens() uint
}

func NewDefaultSemaphore(limit uint) Semaphore {
	return NewChanSemaphore(limit)
}

func NewChanSemaphore(limit uint) Semaphore {
	return &semaphoreChan{
		make(chan int8, limit),
		limit,
	}
}

func NewAtomicSemaphore(limit uint) Semaphore {
	return &semaphoreAtomic{
		limit:   int32(limit),
		counter: 0,
	}
}

type semaphoreChan struct {
	ch    chan int8
	total uint
}

func (s *semaphoreChan) Acquire() bool {
	select {
	case s.ch <- 1:
		return true
	default:
		return false
	}
}

func (s *semaphoreChan) AcquireUntil() {
	s.ch <- 1
}
func (s *semaphoreChan) AcquireTimeout(d time.Duration) bool {
	if d == 0 {
		return s.Acquire()
	}
	if d < 0 {
		panic("invalid timeout")
	}

	select {
	case s.ch <- 1:
		return true
	case <-time.After(d):
		return false
	}
}

func (s *semaphoreChan) Release() {
	<-s.ch
}
func (s *semaphoreChan) TotalTokens() uint {
	return s.total
}

type semaphoreAtomic struct {
	limit   int32
	counter int32
}

func (s *semaphoreAtomic) Acquire() bool {
	c := atomic.AddInt32(&s.counter, 1)
	if c <= s.limit {
		return true
	}
	atomic.AddInt32(&s.counter, -1)
	return false
}

func (s *semaphoreAtomic) AcquireTimeout(d time.Duration) bool {
	if d < 0 {
		panic("invalid timeout")
	}
	rest := d
	for {
		c := atomic.AddInt32(&s.counter, 1)
		if c <= s.limit {
			return true
		}
		atomic.AddInt32(&s.counter, -1)
		if rest == 0 {
			break
		}
		nextSleep := sleepFixTime
		if rest <= sleepFixTime {
			nextSleep = rest
		}
		rest -= nextSleep
		time.Sleep(nextSleep)
	}
	return false
}

func (s *semaphoreAtomic) AcquireUntil() {
	for {
		c := atomic.AddInt32(&s.counter, 1)
		if c <= s.limit {
			return
		}
		atomic.AddInt32(&s.counter, -1)
		time.Sleep(sleepFixTime)
	}
}

func (s *semaphoreAtomic) Release() {
	atomic.AddInt32(&s.counter, -1)
}

func (s *semaphoreAtomic) TotalTokens() uint {
	return uint(s.limit)
}
