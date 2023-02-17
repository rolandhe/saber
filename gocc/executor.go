package gocc

type Task func() (any, error)
type CompleteHandler func(ret any, err error)

type taskResult struct {
	r any
	e error
}

type Executor interface {
	Execute(task Task) (*Future, bool)
	ExecuteInGroup(task Task, g *FutureGroup) (*Future, bool)
}
