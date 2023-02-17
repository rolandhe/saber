package gocc

import (
	"sync/atomic"
	"time"
)

type bufferedExecutor struct {
	queue       BlockingQueue[*ExecTask]
	concurLevel int32
	counter     int32
}

func NewBufferedExecutor(queue BlockingQueue[*ExecTask], concurLevel uint32) Executor {
	executor := &bufferedExecutor{
		queue:       queue,
		concurLevel: int32(concurLevel),
	}
	go dispatch(queue, &executor.counter, executor.concurLevel)
	return executor
}

type ExecTask struct {
	task   Task
	future *Future
}

func dispatch(q BlockingQueue[*ExecTask], counter *int32, limit int32) {
	for {
		elem, ok := q.PullTimeout(time.Second * 5)
		if !ok {
			// todo log
			continue
		}
		c := atomic.AddInt32(counter, 1)
		if c > limit {
			for atomic.LoadInt32(counter) < limit {
				time.Sleep(time.Millisecond * 2)
			}
		}
		execTask := *(elem.v)
		go runTask(execTask.task, execTask.future, counter)
	}
}

func (be *bufferedExecutor) Execute(task Task) (*Future, bool) {
	future := newFuture()
	return be.tryOfferTask(task, future)
}

func (be *bufferedExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	future := newFutureWithGroup(g)
	return be.tryOfferTask(task, future)
}

func (be *bufferedExecutor) tryOfferTask(task Task, future *Future) (*Future, bool) {
	ok := be.queue.Offer(&ExecTask{
		task:   task,
		future: future,
	})
	if !ok {
		return nil, false
	}
	return future, true
}
