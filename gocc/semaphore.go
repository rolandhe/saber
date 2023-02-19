package gocc

import (
	"sync/atomic"
	"time"
)

const sleepFixTime = time.Millisecond * 2

type Semaphore interface {
	Acquire() bool
	AcquireUntil()
	AcquireTimeout(d time.Duration) bool
	Release()
	TotalTokens() uint
}

func NewChanSemaphore(limit uint) Semaphore {
	return &semaphoreImpl{
		make(chan struct{}, limit),
		limit,
	}
}

func NewAtomicSemaphore(limit uint) Semaphore {
	return &semaphoreAtomic{
		limit:   int32(limit),
		counter: 0,
	}
}

type semaphoreImpl struct {
	ch    chan struct{}
	total uint
}

func (s *semaphoreImpl) Acquire() bool {
	select {
	case <-s.ch:
		return true
	default:
		return false
	}
}

func (s *semaphoreImpl) AcquireUntil() {
	<-s.ch
}
func (s *semaphoreImpl) AcquireTimeout(d time.Duration) bool {
	select {
	case <-s.ch:
		return true
	case <-time.After(d):
		return false
	}
}

func (s *semaphoreImpl) Release() {
	s.ch <- struct{}{}
}
func (s *semaphoreImpl) TotalTokens() uint {
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
	nextSleep := sleepFixTime
	if d <= sleepFixTime {
		nextSleep = d
	}
	c := atomic.AddInt32(&s.counter, 1)
	if c <= s.limit {
		return true
	}

	rest := d - nextSleep
	for {
		time.Sleep(nextSleep)
		c = atomic.LoadInt32(&s.counter)
		if c <= s.limit {
			return true
		}
		if rest == 0 {
			break
		}
		if rest <= sleepFixTime {
			nextSleep = rest
		} else {
			nextSleep = sleepFixTime
		}
		rest -= sleepFixTime
	}
	atomic.AddInt32(&s.counter, -1)
	return false
}

func (s *semaphoreAtomic) AcquireUntil() {
	c := atomic.AddInt32(&s.counter, 1)
	if c <= s.limit {
		return
	}
	for {
		time.Sleep(sleepFixTime)
		c = atomic.LoadInt32(&s.counter)
		if c <= s.limit {
			return
		}
	}
}

func (s *semaphoreAtomic) Release() {
	atomic.AddInt32(&s.counter, -1)
}

func (s *semaphoreAtomic) TotalTokens() uint {
	return uint(s.limit)
}
