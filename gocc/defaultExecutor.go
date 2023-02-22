// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package gocc

import (
	"time"
)

func NewDefaultExecutor(concurLevel uint) Executor {
	return &chanExecutor{
		concurLevel:     concurLevel,
		concurrentLimit: NewDefaultSemaphore(concurLevel),
	}
}

type Task func() (any, error)

// Executor 多任务执行器
type Executor interface {
	// Execute 执行任务,如果执行器内资源已耗尽,则直接返回nil,false,否则返回future和true
	Execute(task Task) (*Future, bool)
	// ExecuteTimeout 执行任务,支持超时,如果执行器内资源已耗尽且在超时时间内依然不能获取到资源,则返回nil,false,如果执行器内有资源或者超时时间内能获取资源,返回future,true
	ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool)
	ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool)
	ExecuteInGroupTimeout(task Task, g *FutureGroup, timeout time.Duration) (*Future, bool)
}

type taskResult struct {
	r any
	e error
}

type chanExecutor struct {
	concurLevel     uint
	concurrentLimit Semaphore
}

func (et *chanExecutor) Execute(task Task) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, 0) {
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, et.concurrentLimit)

	return future, true
}

func (et *chanExecutor) ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, timeout) {
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, et.concurrentLimit)

	return future, true
}

func (et *chanExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, 0) {
		return nil, false
	}
	future := newFutureWithGroup(g)
	g.add(future)
	go runTask(task, future, et.concurrentLimit)
	return future, true
}

func (et *chanExecutor) ExecuteInGroupTimeout(task Task, g *FutureGroup, timeout time.Duration) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, timeout) {
		return nil, false
	}

	future := newFutureWithGroup(g)
	go runTask(task, future, et.concurrentLimit)
	return future, true
}

func runTask(task Task, future *Future, concurrentLimit Semaphore) {
	var r any
	var err = TaskCancelledError
	if !future.IsCancelled() {
		r, err = task()
	}

	future.accept(&taskResult{r, err})
	concurrentLimit.Release()
}

func acquireToken(concurrentLimit Semaphore, timeout time.Duration) bool {
	if timeout == 0 {
		if !concurrentLimit.TryAcquire() {
			return false
		}
	} else {
		if !concurrentLimit.AcquireTimeout(timeout) {
			return false
		}
	}
	return true
}
