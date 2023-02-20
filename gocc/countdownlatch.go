package gocc

import (
	"sync/atomic"
	"time"
)

type CountdownLatch struct {
	count *atomic.Int64
	ch    chan struct{}
}

func NewCountdownLatch(count int64) *CountdownLatch {
	if count < 0 {
		panic("invalid count value")
	}
	p := &CountdownLatch{
		count: &atomic.Int64{},
		ch:    make(chan struct{}),
	}
	p.count.Add(count)
	return p
}

func (dw *CountdownLatch) Down() int64 {
	v := dw.count.Add(-1)
	if v == 0 {
		close(dw.ch)
	}
	if v < 0 {
		dw.count.Add(1)
		v = 0
	}
	return v
}
func (dw *CountdownLatch) Wait() bool {
	select {
	case <-dw.ch:
		return true
	default:
		return false
	}
}

func (dw *CountdownLatch) WaitUtil() {
	<-dw.ch
}

func (dw *CountdownLatch) WaitTimeout(timeout time.Duration) bool {
	select {
	case <-dw.ch:
		return true
	case <-time.After(timeout):
		return false

	}
}
