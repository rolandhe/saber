package gocc

import (
	"errors"
	"time"
)

var cancelledError = errors.New("cancelled task")

func NewSimpleExecutor(concurLevel uint) Executor {
	return &simpleExecutor{
		concurLevel:     concurLevel,
		concurrentLimit: NewChanSemaphore(concurLevel),
	}
}

func NewSimpleExecutorWithSemaphore(concurrentLimit Semaphore) Executor {
	return &simpleExecutor{
		concurLevel:     concurrentLimit.GetTotal(),
		concurrentLimit: concurrentLimit,
	}
}

type simpleExecutor struct {
	concurLevel     uint
	concurrentLimit Semaphore
}

func (et *simpleExecutor) Execute(task Task) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, 0) {
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, et.concurrentLimit)

	return future, true
}

func (et *simpleExecutor) ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, timeout) {
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, et.concurrentLimit)

	return future, true
}

func (et *simpleExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, 0) {
		return nil, false
	}
	future := newFutureWithGroup(g)
	g.add(future)
	go runTask(task, future, et.concurrentLimit)
	return future, true
}

func (et *simpleExecutor) ExecuteInGroupTimeout(task Task, g *FutureGroup, timeout time.Duration) (*Future, bool) {
	if !acquireToken(et.concurrentLimit, timeout) {
		return nil, false
	}

	future := newFutureWithGroup(g)
	g.add(future)
	go runTask(task, future, et.concurrentLimit)
	return future, true
}

func runTask(task Task, future *Future, concurrentLimit Semaphore) {
	var r any
	var err = cancelledError
	if !future.Cancelled() {
		r, err = task()
	}

	future.ch <- &taskResult{r, err}
	close(future.ch)
	concurrentLimit.Release()
	future.TryGet()
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
