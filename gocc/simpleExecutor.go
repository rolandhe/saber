package gocc

import "sync/atomic"

func NewSimpleExecutor(concurLevel uint32) Executor {
	return &simpleExecutor{
		concurLevel: int32(concurLevel),
	}
}

type simpleExecutor struct {
	concurLevel int32
	counter     int32
}

func (et *simpleExecutor) Execute(task Task) (*Future, bool) {
	c := atomic.AddInt32(&et.counter, 1)
	if c >= et.concurLevel {
		atomic.AddInt32(&et.counter, -1)
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, &et.counter)

	return future, true
}

func (et *simpleExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	c := atomic.AddInt32(&et.counter, 1)
	if c >= et.concurLevel {
		atomic.AddInt32(&et.counter, -1)
		return nil, false
	}
	future := newFutureWithGroup(g)
	g.add(future)
	go runTask(task, future, &et.counter)
	return future, true
}

func runTask(task Task, future *Future, counterAddr *int32) {
	r, err := task()
	future.ch <- &taskResult{r, err}
	close(future.ch)
	atomic.AddInt32(counterAddr, -1)
	future.TryGet()
}
