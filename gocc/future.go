package gocc

import (
	"errors"
	"sync/atomic"
	"time"
)

var TimeoutError = errors.New("timeout")

type Future struct {
	ch           chan struct{}
	result       *taskResult
	grpNotify    *notify
	canceledFlag atomic.Bool
}

func NewFutureGroup(count int) *FutureGroup {
	return &FutureGroup{
		notifier: &notify{
			counter:    int64(count),
			notifyChan: make(chan struct{}),
		},
		total: uint(count),
	}
}

func newFuture() *Future {
	return &Future{
		ch: make(chan struct{}),
	}
}

func newFutureWithGroup(g *FutureGroup) *Future {
	return &Future{
		ch:        make(chan struct{}),
		grpNotify: g.notifier,
	}
}

func (f *Future) Get() (any, error) {
	select {
	case <-f.ch:
		return f.result.r, f.result.e
	default:
		return nil, TimeoutError
	}
}

func (f *Future) GetUntill() (any, error) {
	<-f.ch
	return f.result.r, f.result.e
}

func (f *Future) Cancel() {
	f.canceledFlag.Store(true)
}

func (f *Future) IsCancelled() bool {
	return f.canceledFlag.Load()
}

func (f *Future) TryGet() bool {
	select {
	case <-f.ch:
		return true
	default:
		return false
	}
}

func (f *Future) GetTimeout(d time.Duration) (any, error) {
	if d < 0 {
		return f.GetUntill()
	}
	if d == 0 {
		return f.Get()
	}

	select {
	case <-f.ch:
		return f.result.r, f.result.e
	case <-time.After(d):
		return nil, TimeoutError
	}
}

func (f *Future) accept(v *taskResult) {
	f.result = v
	close(f.ch)
	if f.grpNotify != nil {
		f.grpNotify.notifyOne()
	}
}

type notify struct {
	counter    int64
	notifyChan chan struct{}
}

func (n *notify) notifyOne() {
	c := atomic.AddInt64(&n.counter, -1)
	if c == 0 {
		close(n.notifyChan)
	}
}

type FutureGroup struct {
	futureGroup []*Future
	//finish      atomic.Bool
	notifier *notify
	total    uint
}

func (fg *FutureGroup) Wait() {
	fg.check()

	<-fg.notifier.notifyChan
}

func (fg *FutureGroup) TryWait() bool {
	fg.check()

	select {
	case <-fg.notifier.notifyChan:
		return true
	default:
		return false
	}
}

func (fg *FutureGroup) WaitTimeout(timeout time.Duration) error {
	fg.check()

	select {
	case <-fg.notifier.notifyChan:
		return nil
	case <-time.After(timeout):
		return TimeoutError
	}
}
func (fg *FutureGroup) Cancel() {
	for _, future := range fg.futureGroup {
		future.Cancel()
	}
}

func (fg *FutureGroup) check() {
	if int(fg.total) != len(fg.futureGroup) {
		panic("future not enough")
	}
}

func (fg *FutureGroup) add(f *Future) {
	fg.futureGroup = append(fg.futureGroup, f)
	if len(fg.futureGroup) > int(fg.total) {
		panic("exceed group size")
	}
}
