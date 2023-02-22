// Package gocc, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.
package gocc

import "time"

func NewDefaultSemaphore(limit uint) Semaphore {
	return &semaphoreChan{
		make(chan struct{}, limit),
		limit,
	}
}

// Semaphore 类似java的信号量, 但只支持单个信号量,支持超时
type Semaphore interface {
	// TryAcquire 尝试获取单个信号量,如果当下没有信号量,则直接返回false
	TryAcquire() bool
	// Acquire 获取单个信号量,如果当前没有信号量,一直阻塞等待,直到获取到位置
	Acquire()
	// AcquireTimeout 超时获取信号量,如果超时时间内没有信号量则返回false,否则返回true
	// d == 0 退化成 TryAcquire
	// d < 0 退出成 Acquire
	AcquireTimeout(d time.Duration) bool
	// Release 释放信号量
	Release()
	// TotalTokens 信号量总数
	TotalTokens() uint
}

// 基于channel实现信号量,这也是golang官方文档中的推荐实现
type semaphoreChan struct {
	tokens chan struct{}
	total  uint
}

func (s *semaphoreChan) TryAcquire() bool {
	select {
	case s.tokens <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *semaphoreChan) Acquire() {
	s.tokens <- struct{}{}
}
func (s *semaphoreChan) AcquireTimeout(d time.Duration) bool {
	if d == 0 {
		return s.TryAcquire()
	}
	if d < 0 {
		s.Acquire()
		return true
	}

	select {
	case s.tokens <- struct{}{}:
		return true
	case <-time.After(d):
		return false
	}
}

func (s *semaphoreChan) Release() {
	<-s.tokens
}
func (s *semaphoreChan) TotalTokens() uint {
	return s.total
}
