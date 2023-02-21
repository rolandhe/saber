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
		if !concurrentLimit.Acquire() {
			return false
		}
	} else {
		if !concurrentLimit.AcquireTimeout(timeout) {
			return false
		}
	}
	return true
}
