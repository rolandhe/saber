package gocc

import (
	"errors"
	"sync/atomic"
	"time"
)

var TimeoutError = errors.New("timeout")

type Future struct {
	ch        chan *taskResult
	tr        atomic.Pointer[taskResult]
	grpNotify *notify
}

func NewFutureGroup(count int) *FutureGroup {
	return &FutureGroup{
		notifier: &notify{
			counter:    int64(count),
			notifyChan: make(chan struct{}, 1),
		},
		total: uint(count),
	}
}

func newFuture() *Future {
	return &Future{
		ch: make(chan *taskResult, 1),
	}
}

func newFutureWithGroup(g *FutureGroup) *Future {
	return &Future{
		ch:        make(chan *taskResult, 1),
		grpNotify: g.notifier,
	}
}

func (f *Future) Get() (any, error) {
	p := f.tr.Load()
	if p != nil {
		return p.r, p.e
	}
	v, ok := <-f.ch
	if ok {
		f.accept(v)
		return v.r, v.e
	}
	return f.getInternalResult(p)
}

func (f *Future) TryGet() bool {
	p := f.tr.Load()
	if p != nil {
		return true
	}
	select {
	case v, ok := <-f.ch:
		if ok {
			f.accept(v)
		}
		return true
	default:
		return false
	}
}

func (f *Future) GetTimeout(d time.Duration) (any, error) {
	if d < 0 {
		return f.Get()
	}
	if d == 0 {
		if f.TryGet() {
			return f.Get()
		}
		return nil, TimeoutError
	}
	p := f.tr.Load()
	if p != nil {
		return p.r, p.e
	}
	select {
	case v, ok := <-f.ch:
		if ok {
			f.accept(v)
			return v.r, v.e
		}
		return f.getInternalResult(p)
	case <-time.After(d):
		return nil, TimeoutError
	}
}

func (f *Future) getInternalResult(p *taskResult) (any, error) {
	for {
		p = f.tr.Load()
		if p != nil {
			return p.r, p.e
		}
	}
}
func (f *Future) accept(v *taskResult) {
	f.tr.Store(v)
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
		n.notifyChan <- struct{}{}
		close(n.notifyChan)
	}
}

type FutureGroup struct {
	futureGroup []*Future
	finish      atomic.Bool
	notifier    *notify
	total       uint
}

func (fg *FutureGroup) Wait() {
	fg.check()
	if fg.finish.Load() {
		return
	}
	_, ok := <-fg.notifier.notifyChan
	if ok {
		fg.finish.Store(true)
	}
}

func (fg *FutureGroup) TryWait() bool {
	fg.check()
	if fg.finish.Load() {
		return true
	}
	select {
	case _, ok := <-fg.notifier.notifyChan:
		if ok {
			fg.finish.Store(true)
		}
		return true
	default:
		return false
	}
}

func (fg *FutureGroup) WaitTimeout(timeout time.Duration) error {
	fg.check()
	if fg.finish.Load() {
		return nil
	}
	select {
	case _, ok := <-fg.notifier.notifyChan:
		if ok {
			fg.finish.Store(true)
		}
		return nil
	case <-time.After(timeout):
		return TimeoutError
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
