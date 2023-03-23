// Golang concurrent tools like java juc.
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

// NewFutureGroup 构建一个Future组，每组内的Future数量需要事先知道。
//
//	count 组内Future的数量
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

// Future 与java类似,提供任务执行结果占位符的能力,持有Future的线程可以调用:
//
// Get 等待结果返回,如果已经调用Cancel方法,则err是TaskCancelledError
//
// TryGet, 尝试获取返回结果,无论有无结果都立即返回,有则err是nil,无则返回TimeoutError
//
// GetTimeout, 待超时的等待,如果超时时间内有结果则err是nil, 无怎返回TimeoutError
//
// Cancel, 取消当前任务,执行取消后,任务可以已经被执行,也可能不执行,当前线程可以直接去处理其他事情,无需关心任务的执行结果
type Future struct {
	ch            chan struct{}
	result        *taskResult
	groupNotifier *CountdownLatch
	canceledFlag  atomic.Bool
}

// Get 等待结果返回,如果已经调用Cancel方法,则err是TaskCancelledError
func (f *Future) Get() (any, error) {
	if f.IsCancelled() {
		return nil, TaskCancelledError
	}
	<-f.ch
	return f.result.r, f.result.e
}

// TryGet 尝试获取返回结果,无论有无结果都立即返回,有则err是nil,无则返回TimeoutError
func (f *Future) TryGet() (any, error) {
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

// Cancel 取消当前任务,执行取消后,任务可以已经被执行,也可能不执行,当前线程可以直接去处理其他事情,无需关心任务的执行结果
func (f *Future) Cancel() {
	f.canceledFlag.Store(true)
}

// IsCancelled 任务是否被设置为取消状态
func (f *Future) IsCancelled() bool {
	return f.canceledFlag.Load()
}

// GetTimeout 待超时的等待,如果超时时间内有结果则err是nil, 无怎返回TimeoutError
func (f *Future) GetTimeout(d time.Duration) (any, error) {
	if f.IsCancelled() {
		return nil, TaskCancelledError
	}
	if d < 0 {
		return f.Get()
	}
	if d == 0 {
		return f.TryGet()
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

// FutureGroup 用于批量管理任务的组件, 需要同时执行一批任务再等待他们的执行结果的场景可以是FutureGroup, 你可以先设置需要执行任务的个数,然后逐个的加入待执行任务,最后在FutureGroup上等待执行结果.
//
// 注意:
//
// 1. FutureGroup需要预先知道待执行任务的个数, 这么设计有两个考虑,一是这符合大多数的使用场景,二是相比动态计算任务数底层无需太多的并发安全考虑,性能更好
//
// 2. add/Cancel,方法并不是线程安全的,也就是说必须要在一个线程内把所有的任务放入Group, 且在同一个线程内Cancel;Get/GetTimeout/TryGet线程安全的
//
// 3. 切记, Group内的任务数必须和预先设定的个数相同,否则永远也无法等到结束
type FutureGroup struct {
	futureGroup []*Future
	notifier    *CountdownLatch
	total       int64
}

// Get 等待所有任务完成,返回结果
func (fg *FutureGroup) Get() []*Future {
	fg.notifier.Wait()
	return fg.futureGroup
}

// TryGet 测试所有任务完成,如果完成直接返回结果,否则返回TimeoutError
func (fg *FutureGroup) TryGet() ([]*Future, error) {
	if fg.notifier.TryWait() {
		return fg.futureGroup, nil
	}
	return nil, TimeoutError
}

// GetTimeout 超时时间内测试所有任务完成,如果完成直接返回结果,否则返回TimeoutError
func (fg *FutureGroup) GetTimeout(timeout time.Duration) ([]*Future, error) {
	if fg.notifier.WaitTimeout(timeout) {
		return fg.futureGroup, nil
	}
	return nil, TimeoutError
}

// Cancel 取消所有的任务, 无法保证线程安全, 或者在创建Group的线程内执行,这是大多数场景,或者需要调用者保持线程安全,比如创建一个新的goroutine来执行Cancels
func (fg *FutureGroup) Cancel() {
	for _, future := range fg.futureGroup {
		future.Cancel()
	}
}

func (fg *FutureGroup) add(f *Future) {
	fg.futureGroup = append(fg.futureGroup, f)
	if len(fg.futureGroup) > int(fg.total) {
		panic("exceed group size")
	}
}
