package gocc

import "time"

type Semaphore interface {
	Acquire() bool
	AcquireTimeout(d time.Duration) bool
	Release()
}

func NewSemaphore(limit uint) Semaphore {
	return &semaphoreImpl{
		make(chan struct{}, limit),
	}
}

type semaphoreImpl struct {
	ch chan struct{}
}

func (s *semaphoreImpl) Acquire() bool {
	select {
	case <-s.ch:
		return true
	default:
		return false
	}
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
