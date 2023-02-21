package gocc

import "time"

type Task func() (any, error)
type CompleteHandler func(ret any, err error)

type Executor interface {
	Execute(task Task) (*Future, bool)
	ExecuteTimeout(task Task, timeout time.Duration) (*Future, bool)
	ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool)
	ExecuteInGroupTimeout(task Task, g *FutureGroup, timeout time.Duration) (*Future, bool)
}

type taskResult struct {
	r any
	e error
}
