// Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.

package gocc

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// code from https://gist.github.com/zviadm/c234426882bfc8acba88f3503edaaa36#file-cond2-go

// Condition 支持wait timeout的同步条件
type Condition interface {
	// Wait 同Cond Wait, 等待直到被唤醒
	Wait()
	// WaitWithTimeout , 带有超时的等待
	WaitWithTimeout(t time.Duration)
	// Broadcast 同Cond Broadcast
	Broadcast()
	// Signal 同Cond Signal
	Signal()
}

// NewCondTimeout 构建支持timeout的cond
func NewCondTimeout(l sync.Locker) Condition {
	c := &condTimeout{locker: l}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

// NewCondTimeoutWithName 构建Condition，可以指定名字, 日志输出是带有名字，方便排查问题
func NewCondTimeoutWithName(l sync.Locker, name string) Condition {
	c := &condTimeout{locker: l, name: name}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

type condTimeout struct {
	locker sync.Locker
	n      unsafe.Pointer
	name   string
}

// Wait Waits for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *condTimeout) Wait() {
	n := c.NotifyChan()
	c.locker.Unlock()
	<-n
	c.locker.Lock()
}

// WaitWithTimeout Same as Wait() call, but will only wait up to a given timeout.
func (c *condTimeout) WaitWithTimeout(t time.Duration) {
	n := c.NotifyChan()
	c.locker.Unlock()

	select {
	case <-n:
	case <-time.After(t):
		CcLogger.Info("name:%s,wait with timeout\n", c.name)
	}
	c.locker.Lock()
}

// NotifyChan Returns a channel that can be used to wait for next Broadcast() call.
func (c *condTimeout) NotifyChan() chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

// Broadcast call notifies everyone that something has changed.
func (c *condTimeout) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}

func (c *condTimeout) Signal() {
	n := c.NotifyChan()
	select {
	case n <- struct{}{}:
	default:
	}
}
