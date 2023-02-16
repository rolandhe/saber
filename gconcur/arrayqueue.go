package gconcur

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

func (a *arrayBlockingQueue[T]) Offer(t T) bool {
	a.Lock()
	defer a.Unlock()

	if !a.q.hasCap() {
		return false
	}
	a.q.w(&Elem[T]{&t})
	a.rc.Signal()
	return true
}

func (a *arrayBlockingQueue[T]) OfferTimeout(t T, timeout time.Duration) bool {
	a.Lock()
	if a.q.hasCap() {
		a.q.w(&Elem[T]{&t})
		a.Unlock()
		return true
	}
	wt := int64(timeout)
	hasCap := false
	for !hasCap && wt > 0 {
		start := time.Now().UnixNano()
		a.wc.WaitWithTimeout(time.Duration(wt))
		hasCap = a.q.hasCap()
		if !hasCap {
			cost := time.Now().UnixNano() - start
			wt -= cost
		}
	}
	if hasCap {
		a.q.w(&Elem[T]{&t})
	}
	a.Unlock()
	a.rc.Signal()
	return hasCap
}

func (a *arrayBlockingQueue[T]) Pull() (*Elem[T], bool) {
	a.Lock()
	defer a.Unlock()

	if a.q.count() == 0 {
		return nil, false
	}
	e := a.q.r()
	a.wc.Signal()
	return e, true
}

func (a *arrayBlockingQueue[T]) PullTimeout(timeout time.Duration) (*Elem[T], bool) {
	a.Lock()
	if a.q.count() > 0 {
		e := a.q.r()
		a.Unlock()
		a.wc.Signal()
		return e, true
	}
	wt := int64(timeout)
	hasElem := false
	for !hasElem && wt > 0 {
		start := time.Now().UnixNano()
		a.wc.WaitWithTimeout(time.Duration(wt))
		hasElem = a.q.count() > 0
		if !hasElem {
			cost := time.Now().UnixNano() - start
			wt -= cost
		}
	}
	var e *Elem[T]
	if hasElem {
		e = a.q.r()
		a.wc.Signal()
	}
	a.Unlock()

	return e, hasElem
}
