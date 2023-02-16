package gconcur

import "time"

const defaultLimit = 128

type Elem[T any] struct {
	v *T
}

func (el *Elem[T]) GetValue() T {
	return *(el.v)
}

type BlockingQueue[T any] interface {
	Offer(t T) bool
	OfferTimeout(t T, timeout time.Duration) bool
	Pull() (*Elem[T], bool)
	PullTimeout(timeout time.Duration) (*Elem[T], bool)
}
