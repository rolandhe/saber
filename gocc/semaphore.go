package gocc

import (
	"time"
)

type Semaphore interface {
	Acquire() bool
	AcquireUntil()
	AcquireTimeout(d time.Duration) bool
	Release()
	TotalTokens() uint
}
