package gocc

import (
	"time"
)

func NewDefaultBlockingQueue[T any](limit int) BlockingQueue[T] {
	return NewChanBlockingQueue[T](limit)
}

func NewChanBlockingQueue[T any](limit int) BlockingQueue[T] {
	if limit < 0 {
		limit = defaultLimit
	}
	return &chanImpl[T]{
		make(chan *Elem[T], limit),
	}
}

type chanImpl[T any] struct {
	q chan *Elem[T]
}

func (ci *chanImpl[T]) Offer(t T) bool {
	select {
	case ci.q <- &Elem[T]{&t}:
		return true
	default:
		return false
	}
}

func (ci *chanImpl[T]) OfferTimeout(t T, timeout time.Duration) bool {
	select {
	case ci.q <- &Elem[T]{&t}:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (ci *chanImpl[T]) Pull() (*Elem[T], bool) {
	select {
	case v, ok := <-ci.q:
		if !ok {
			return nil, false
		}
		return v, true
	default:
		return nil, false
	}
}
func (ci *chanImpl[T]) PullTimeout(timeout time.Duration) (*Elem[T], bool) {
	select {
	case v, ok := <-ci.q:
		if !ok {
			return nil, false
		}
		return v, true
	case <-time.After(timeout):
		return nil, false
	}
}
