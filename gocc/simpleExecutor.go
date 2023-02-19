package gocc

func NewSimpleExecutor(concurLevel uint) Executor {
	return &simpleExecutor{
		concurLevel: concurLevel,
		semaphore:   NewChanSemaphore(concurLevel),
	}
}

func NewSimpleExecutorWithSemaphore(semaphore Semaphore) Executor {
	return &simpleExecutor{
		concurLevel: semaphore.GetTotal(),
		semaphore:   semaphore,
	}
}

type simpleExecutor struct {
	concurLevel uint
	semaphore   Semaphore
}

func (et *simpleExecutor) Execute(task Task) (*Future, bool) {
	if et.semaphore.Acquire() {
		return nil, false
	}
	future := newFuture()
	go runTask(task, future, et.semaphore)

	return future, true
}

func (et *simpleExecutor) ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool) {
	if et.semaphore.Acquire() {
		return nil, false
	}
	future := newFutureWithGroup(g)
	g.add(future)
	go runTask(task, future, et.semaphore)
	return future, true
}

func runTask(task Task, future *Future, concurrentLimit Semaphore) {
	r, err := task()
	future.ch <- &taskResult{r, err}
	close(future.ch)
	concurrentLimit.Release()
	future.TryGet()
}
