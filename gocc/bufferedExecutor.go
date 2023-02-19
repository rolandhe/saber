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
		for {
			if !concurrentLimit.AcquireTimeout(waitTimeout) {
				log.Println("met concurrent limit, wait....")
				continue
			}
			break
		}

		execTask := *(elem.v)
		go runTask(execTask.task, execTask.future, concurrentLimit)
	}
}

func (be *bufferedExecutor) Execute(task Task) (*Future, bool) {
	future := newFuture()
	return be.tryToOfferTask(task, future)
}

func (be *bufferedExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	future := newFutureWithGroup(g)
	return be.tryToOfferTask(task, future)
}

func (be *bufferedExecutor) tryToOfferTask(task Task, future *Future) (*Future, bool) {
	ok := be.queue.Offer(&ExecTask{
		task:   task,
		future: future,
	})
	if !ok {
		return nil, false
	}
	return future, true
}
