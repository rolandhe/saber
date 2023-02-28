// Package sortutil, sort  tool
//
// Copyright 2023 The saber Authors. All rights reserved.

package sortutil

import "sort"

// 来自官方文档的例子，很巧妙的使用函数的方法实现类似泛型的排序，妙

type Cmp[T any] func(p1, p2 *T) bool

// Sort is a method on the function type, Cmp, that sorts the argument slice according to the function.
func (cmp Cmp[T]) Sort(data []T) {
	ps := &sortHolder[T]{
		data: data,
		cmp:  cmp, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

type sortHolder[T any] struct {
	data []T
	cmp  func(p1, p2 *T) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *sortHolder[T]) Len() int {
	return len(s.data)
}

// Swap is part of sort.Interface.
func (s *sortHolder[T]) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *sortHolder[T]) Less(i, j int) bool {
	return s.cmp(&s.data[i], &s.data[j])
}
