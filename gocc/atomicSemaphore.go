// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package gocc

import (
	"sync/atomic"
	"time"
)

// unstable, 推荐使用NewDefaultSemaphore

const DefaultSleepFixTime = time.Millisecond * 1

func NewAtomicSemaphore(limit uint) Semaphore {
	return &semaphoreAtomic{
		limit:           int32(limit),
		counter:         0,
		fixWaitInterval: DefaultSleepFixTime,
	}
}

func NewAtomicSemaWithWaitInterval(limit uint, waitInterval time.Duration) Semaphore {
	return &semaphoreAtomic{
		limit:           int32(limit),
		counter:         0,
		fixWaitInterval: waitInterval,
	}
}

type semaphoreAtomic struct {
	limit           int32
	counter         int32
	fixWaitInterval time.Duration
}

func (s *semaphoreAtomic) TryAcquire() bool {
	c := atomic.AddInt32(&s.counter, 1)
	if c <= s.limit {
		return true
	}
	atomic.AddInt32(&s.counter, -1)
	return false
}

func (s *semaphoreAtomic) AcquireTimeout(d time.Duration) bool {
	if d < 0 {
		s.Acquire()
		return true
	}
	if d == 0 {
		return s.TryAcquire()
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
		nextSleep := s.fixWaitInterval
		if rest <= s.fixWaitInterval {
			nextSleep = rest
		}
		rest -= nextSleep
		time.Sleep(nextSleep)
	}
	return false
}

func (s *semaphoreAtomic) Acquire() {
	for {
		c := atomic.AddInt32(&s.counter, 1)
		if c <= s.limit {
			return
		}
		atomic.AddInt32(&s.counter, -1)
		time.Sleep(s.fixWaitInterval)
	}
}

func (s *semaphoreAtomic) Release() {
	atomic.AddInt32(&s.counter, -1)
}

func (s *semaphoreAtomic) TotalTokens() uint {
	return uint(s.limit)
}
