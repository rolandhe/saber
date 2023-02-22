// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
package gocc

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	TimeoutError       = errors.New("timeout")
	TaskCancelledError = errors.New("future task is cancelled")
)

func NewFutureGroup(count uint64) *FutureGroup {
	return &FutureGroup{
		notifier: NewCountdownLatch(int64(count)),
		total:    int64(count),
	}
}

func newFuture() *Future {
	return &Future{
		ch: make(chan struct{}),
	}
}

func newFutureWithGroup(g *FutureGroup) *Future {
	future := &Future{
		ch:            make(chan struct{}),
		groupNotifier: g.notifier,
	}
	g.add(future)
	return future
}

type Future struct {
	ch            chan struct{}
	result        *taskResult
	groupNotifier *CountdownLatch
	canceledFlag  atomic.Bool
}

func (f *Future) Get() (any, error) {
	if f.IsCancelled() {
		return nil, TaskCancelledError
	}
	select {
	case <-f.ch:
		return f.result.r, f.result.e
	default:
		return nil, TimeoutError
	}
}

func (f *Future) GetUntil() (any, error) {
	if f.IsCancelled() {
		return nil, TaskCancelledError
	}
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
	if f.IsCancelled() {
		return true
	}
	select {
	case <-f.ch:
		return true
	default:
		return false
	}
}

func (f *Future) GetTimeout(d time.Duration) (any, error) {
	if f.IsCancelled() {
		return nil, TaskCancelledError
	}
	if d < 0 {
		return f.GetUntil()
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
	if f.groupNotifier != nil {
		f.groupNotifier.Down()
	}
}

type FutureGroup struct {
	futureGroup []*Future
	notifier    *CountdownLatch
	total       int64
}

func (fg *FutureGroup) GetFutures() ([]*Future, bool) {
	ok := fg.TryWait()
	if !ok {
		return nil, false
	}
	return fg.futureGroup, true
}

func (fg *FutureGroup) Wait() {
	fg.check()
	fg.notifier.Wait()
}

func (fg *FutureGroup) TryWait() bool {
	fg.check()
	return fg.notifier.TryWait()
}

func (fg *FutureGroup) WaitTimeout(timeout time.Duration) error {
	fg.check()

	if fg.notifier.WaitTimeout(timeout) {
		return nil
	}
	return TimeoutError
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
