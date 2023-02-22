// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
package gocc

import (
	"sync/atomic"
	"time"
)

func NewCountdownLatch(count int64) *CountdownLatch {
	if count < 0 {
		panic("invalid tokenCount value")
	}
	p := &CountdownLatch{
		tokenCount: &atomic.Int64{},
		notifier:   make(chan struct{}),
	}
	p.tokenCount.Add(count)
	return p
}

// CountdownLatch ,类似java的CountdownLatch,实现倒计数功能，支持wait timeout
type CountdownLatch struct {
	tokenCount *atomic.Int64
	notifier   chan struct{}
}

// Down 倒计数减一
func (dw *CountdownLatch) Down() int64 {
	if dw.tokenCount.Load() <= 0 {
		return 0
	}
	v := dw.tokenCount.Add(-1)
	if v == 0 {
		close(dw.notifier)
	}
	if v < 0 {
		v = 0
	}
	return v
}

// TryWait 尝试等待, 如果已经倒计数到0,则返回true,否则立即返回false
func (dw *CountdownLatch) TryWait() bool {
	select {
	case <-dw.notifier:
		return true
	default:
		return false
	}
}

// Wait 等待，直到倒计数到0
func (dw *CountdownLatch) Wait() {
	<-dw.notifier
}

// WaitTimeout 带超时的等待倒计数到0,如果超时内倒计数到0，返回true,否则返回false
func (dw *CountdownLatch) WaitTimeout(timeout time.Duration) bool {
	select {
	case <-dw.notifier:
		return true
	case <-time.After(timeout):
		return false
	}
}
