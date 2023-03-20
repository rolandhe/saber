// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package gocc

import (
	"time"
)

const defaultLimit = 128

// NewDefaultBlockingQueue 构建一个有界队列，缺省的队列底层使用channel
//
//	limit 队列容量
func NewDefaultBlockingQueue[T any](limit int64) BlockingQueue[T] {
	if limit < 0 {
		limit = defaultLimit
	}
	return &chanBlockingQueue[T]{
		make(chan *Elem[T], limit),
	}
}

// Elem 进入队列的元素使用Elem来封装
type Elem[T any] struct {
	v *T
}

// GetValue 获取Elem内封装的数据
func (el *Elem[T]) GetValue() T {
	return *(el.v)
}

// BlockingQueue 线程安全地有界队列, 支持阻塞offer/pull, 超时offer/pull,立即offer/pull
type BlockingQueue[T any] interface {
	// Offer 写入队列,如果队列已满,等待,直到写入位置
	Offer(t T)
	// TryOffer 尝试写入队列,如果当前队列已满,立即返回false,如果能够写入,返回true
	TryOffer(t T) bool
	// OfferTimeout 超时写入,等待时间内写入返回true,否则返回false
	OfferTimeout(t T, timeout time.Duration) bool
	// Pull 与Offer相反, 从队列读取数据,如果队列为空,阻塞,直至有数据位置
	Pull() *Elem[T]
	// TryPull 尝试读取数据,如果队列有数据直接返回,如果队列为空,返回false
	TryPull() (*Elem[T], bool)
	// PullTimeout 超时读取,等待时间内读取到数据返回true,否则返回false
	PullTimeout(timeout time.Duration) (*Elem[T], bool)
}

// 基于golang channel实现,构建buffered channel,利用channel的并发安全能力来实现并发队列,
// channel是不能被close的,所以应用是推荐使用TryOffer/TryPull和超时方法
type chanBlockingQueue[T any] struct {
	q chan *Elem[T]
}

func (ci *chanBlockingQueue[T]) Offer(t T) {
	ci.q <- &Elem[T]{&t}
}

func (ci *chanBlockingQueue[T]) TryOffer(t T) bool {
	select {
	case ci.q <- &Elem[T]{&t}:
		return true
	default:
		return false
	}
}

func (ci *chanBlockingQueue[T]) OfferTimeout(t T, timeout time.Duration) bool {
	select {
	case ci.q <- &Elem[T]{&t}:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (ci *chanBlockingQueue[T]) Pull() *Elem[T] {
	v := <-ci.q
	return v
}

func (ci *chanBlockingQueue[T]) TryPull() (*Elem[T], bool) {
	select {
	case v := <-ci.q:
		return v, true
	default:
		return nil, false
	}
}
func (ci *chanBlockingQueue[T]) PullTimeout(timeout time.Duration) (*Elem[T], bool) {
	select {
	case v := <-ci.q:
		return v, true
	case <-time.After(timeout):
		return nil, false
	}
}
