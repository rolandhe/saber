package gocc

import (
	"sync"
	"time"
)

func NewArrayBlockingQueue[T any](limit int, condFactory func(locker sync.Locker) SyncCondition) BlockingQueue[T] {
	locker := &sync.Mutex{}

	return &arrayBlockingQueue[T]{
		locker,
		condFactory(locker),
		condFactory(locker),
		&ringBuffer[T]{make([]*Elem[T], limit, limit), int64(limit), 0, 0},
		limit,
	}
}

type ringBuffer[T any] struct {
	buf   []*Elem[T]
	limit int64
	wi    int64
	ri    int64
}

func (b *ringBuffer[T]) hasCap() bool {
	return b.limit > b.count()
}

func (b *ringBuffer[T]) count() int64 {
	return b.wi - b.ri
}

func (b *ringBuffer[T]) w(e *Elem[T]) {
	index := int(b.wi % b.limit)
	b.buf[index] = e
	b.wi++
}
func (b *ringBuffer[T]) r() *Elem[T] {
	index := int(b.ri % b.limit)
	b.ri++
	return b.buf[index]
}

type arrayBlockingQueue[T any] struct {
	sync.Locker
	rc    SyncCondition
	wc    SyncCondition
	q     *ringBuffer[T]
	limit int
}

func (aq *arrayBlockingQueue[T]) Offer(t T) bool {
	aq.Lock()
	defer aq.Unlock()

	if !aq.q.hasCap() {
		return false
	}
	aq.q.w(&Elem[T]{&t})
	aq.rc.Signal()
	return true
}

func (aq *arrayBlockingQueue[T]) OfferTimeout(t T, timeout time.Duration) bool {
	aq.Lock()
	if aq.q.hasCap() {
		aq.q.w(&Elem[T]{&t})
		aq.Unlock()
		return true
	}
	wt := int64(timeout)
	hasCap := false
	for !hasCap && wt > 0 {
		start := time.Now().UnixNano()
		aq.wc.WaitWithTimeout(time.Duration(wt))
		hasCap = aq.q.hasCap()
		if !hasCap {
			cost := time.Now().UnixNano() - start
			wt -= cost
		}
	}
	if hasCap {
		aq.q.w(&Elem[T]{&t})
	}
	aq.Unlock()
	aq.rc.Signal()
	return hasCap
}

func (aq *arrayBlockingQueue[T]) Pull() (*Elem[T], bool) {
	aq.Lock()
	defer aq.Unlock()

	if aq.q.count() == 0 {
		return nil, false
	}
	e := aq.q.r()
	aq.wc.Signal()
	return e, true
}

func (aq *arrayBlockingQueue[T]) PullTimeout(timeout time.Duration) (*Elem[T], bool) {
	aq.Lock()
	if aq.q.count() > 0 {
		elem := aq.q.r()
		aq.Unlock()
		aq.wc.Signal()
		return elem, true
	}
	wt := int64(timeout)
	hasElem := false
	for !hasElem && wt > 0 {
		start := time.Now().UnixNano()
		aq.wc.WaitWithTimeout(time.Duration(wt))
		hasElem = aq.q.count() > 0
		if !hasElem {
			cost := time.Now().UnixNano() - start
			wt -= cost
		}
	}
	var elem *Elem[T]
	if hasElem {
		elem = aq.q.r()
		aq.wc.Signal()
	}
	aq.Unlock()

	return elem, hasElem
}