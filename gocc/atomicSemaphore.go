package gocc

import (
	"sync/atomic"
	"time"
)

// unstable, 推荐使用NewDefaultSemaphore

const sleepFixTime = time.Millisecond * 1

// unstable, 推荐使用NewDefaultSemaphore

func NewAtomicSemaphore(limit uint) Semaphore {
	return &semaphoreAtomic{
		limit:   int32(limit),
		counter: 0,
	}
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
