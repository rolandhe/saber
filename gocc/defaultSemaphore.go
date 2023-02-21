package gocc

import "time"

func NewDefaultSemaphore(limit uint) Semaphore {
	return &semaphoreChan{
		make(chan int8, limit),
		limit,
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
