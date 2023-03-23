// Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package gocc

import (
	"time"
)

// NewDefaultExecutor 构建指定并发任务量的任务执行器
//
//	concurLevel  可以并发执行的任务总量
func NewDefaultExecutor(concurLevel uint) Executor {
	return &chanExecutor{
		concurLevel:     concurLevel,
		concurrentLimit: NewDefaultSemaphore(concurLevel),
	}
}

// Task 需要执行的任务主体，它是包含任务业务逻辑的函数
type Task func() (any, error)

// Executor 多任务执行器
type Executor interface {

	// Execute 执行任务,如果执行器内资源已耗尽,则直接返回nil,false,否则返回future和true
	Execute(task Task) (*Future, bool)

	// ExecuteTimeout 执行任务,支持超时,如果执行器内资源已耗尽且在超时时间内依然不能获取到资源,则返回nil,false,如果执行器内有资源或者超时时间内能获取资源,返回future,true
	ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool)

	// ExecuteInGroup 与Execute类似,只是返回的future会被加到FutureGroup中, 可以使用FutureGroup来管理批量的任务:
	//
	//1. 主线程在FutureGroup上等待多个任务完成,没有必要自己循环扫描多个Future
	//
	//2. 从FutureGroup中拿出多个Future,不需要自己维护
	ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool)
	
	// ExecuteInGroupTimeout 与ExecuteTimeout类似,只是增加了FutureGroup
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
