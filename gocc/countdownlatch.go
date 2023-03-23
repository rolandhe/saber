// Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.

// Package gocc, 提供类似java juc的库能力。包括:
//
// Cond with timeout，相比于标准库Cond，它最典型的特点是支持timeout
//
// Future，类似java的Future，提交到任务执行器的任务可以被异步执行，调用者持有Future来获取异步执行的结果或者取消任务
//
// FutureGroup, 包含多个Future，可以在FutureGroup上等待所有的Future任务都执行完成，也可以取消任务，相比于在多个Future上一个个轮询，调用更加简单
//
// BlockingQueue, 支持并发调用的、并行安全的队列，强制有界
//
// Executor, 用于异步执行任务的执行器，强制指定并发数。要执行的任务提交给Executor后马上返回Future，调用者持有Future来获取最终结果，Executor内执行完成任务或者发现任务取消后会修改Future的内部状态
//
// Semaphore，信号量
//
// CountdownLatch， 倒计数
package gocc

import (
	"sync/atomic"
	"time"
)

// NewCountdownLatch 构建总量为count的倒计数，相比于WaitGroup, CountdownLatch提供的能力更丰富
//
//	count  数据总量
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
