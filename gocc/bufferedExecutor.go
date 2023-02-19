package gocc

import (
	"log"
	"time"
)

const waitTimeout = time.Second * 5

type bufferedExecutor struct {
	queue BlockingQueue[*ExecTask]
}

func NewBufferedExecutor(queue BlockingQueue[*ExecTask], concurLevel uint) Executor {
	executor := &bufferedExecutor{
		queue: queue,
	}
	go dispatch(queue, NewChanSemaphore(concurLevel))
	return executor
}

func NewBufferedExecutorWithSemaphore(queue BlockingQueue[*ExecTask], concurrentLimit Semaphore) Executor {
	executor := &bufferedExecutor{
		queue: queue,
	}
	go dispatch(queue, concurrentLimit)
	return executor
}

type ExecTask struct {
	task   Task
	future *Future
}

func dispatch(q BlockingQueue[*ExecTask], concurrentLimit Semaphore) {
	for {
		elem, ok := q.PullTimeout(waitTimeout)
		if !ok {
			log.Println("empty q, try to poll")
			continue
		}
		execTask := *(elem.v)
		if checkCancelled(execTask.future) {
			continue
		}
		skip := false
		for {
			if !concurrentLimit.AcquireTimeout(waitTimeout) {
				if checkCancelled(execTask.future) {
					skip = true
					break
				}
				log.Println("met concurrent limit, wait timeout , and next...")
				continue
			}
			break
		}
		if skip {
			continue
		}

		go runTask(execTask.task, execTask.future, concurrentLimit)
	}
}

func checkCancelled(future *Future) bool {
	if future.IsCancelled() {
		future.ch <- &taskResult{nil, cancelledError}
		close(future.ch)
		future.TryGet()
		return true
	}
	return false
}

func (be *bufferedExecutor) Execute(task Task) (*Future, bool) {
	future := newFuture()
	return be.tryToOfferTask(task, future, 0)
}
func (be *bufferedExecutor) ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool) {
	future := newFuture()
	return be.tryToOfferTask(task, future, timeout)
}
func (be *bufferedExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	future := newFutureWithGroup(g)
	return be.tryToOfferTask(task, future, 0)
}

func (be *bufferedExecutor) ExecuteInGroupTimeout(task Task, g *FutureGroup, timeout time.Duration) (*Future, bool) {
	future := newFutureWithGroup(g)
	return be.tryToOfferTask(task, future, timeout)
}

func (be *bufferedExecutor) tryToOfferTask(task Task, future *Future, timeout time.Duration) (*Future, bool) {
	var ok = false
	if timeout == 0 {
		ok = be.queue.Offer(&ExecTask{
			task:   task,
			future: future,
		})
	} else {
		ok = be.queue.OfferTimeout(&ExecTask{
			task:   task,
			future: future,
		}, timeout)
	}
	ok = be.queue.Offer(&ExecTask{
		task:   task,
		future: future,
	})
	if !ok {
		return nil, false
	}
	return future, true
}
