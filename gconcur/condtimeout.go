package gconcur

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type condTimeout struct {
	locker sync.Locker
	n      unsafe.Pointer
}

type SyncCondition interface {
	Wait()
	WaitWithTimeout(t time.Duration)
	Broadcast()
	Signal()
}

func NewCondTimeout(l sync.Locker) SyncCondition {
	c := &condTimeout{locker: l}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

// Wait Waits for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *condTimeout) Wait() {
	n := c.NotifyChan()
	c.locker.Unlock()
	<-n
	c.locker.Lock()
}

// WaitWithTimeout Same as Wait() call, but will only wait up to a given timeout.
func (c *condTimeout) WaitWithTimeout(t time.Duration) {
	n := c.NotifyChan()
	c.locker.Unlock()
	select {
	case <-n:
	case <-time.After(t):
	}
	c.locker.Lock()
}

// NotifyChan Returns a channel that can be used to wait for next Broadcast() call.
func (c *condTimeout) NotifyChan() chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

// Broadcast call notifies everyone that something has changed.
func (c *condTimeout) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}

func (c *condTimeout) Signal() {
	n := c.NotifyChan()
	select {
	case n <- struct{}{}:
	default:
	}
}
