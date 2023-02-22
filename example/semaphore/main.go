package main

import (
	"github.com/rolandhe/saber/gocc"
	"log"
	"sync"
	"time"
)

func main() {
	//waitUntil()
	waitUntilChan()
	//waitWithTimeout()
}

func waitUntil() {
	limit := gocc.NewAtomicSemaphore(500)
	waiter := sync.WaitGroup{}
	waiter.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(id int) {
			start := time.Now().UnixNano()
			limit.Acquire()
			waitTime := time.Now().UnixNano() - start
			time.Sleep(time.Millisecond * 200)
			log.Printf("g:%d,wait:%d\n", id, waitTime)
			limit.Release()
			waiter.Done()
		}(i)
	}
	waiter.Wait()
}

func waitUntilChan() {
	limit := gocc.NewDefaultSemaphore(500)
	waiter := sync.WaitGroup{}
	waiter.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(id int) {
			start := time.Now().UnixNano()
			limit.Acquire()
			waitTime := time.Now().UnixNano() - start
			time.Sleep(time.Millisecond * 200)
			log.Printf("g:%d,wait:%d\n", id, waitTime)
			limit.Release()
			waiter.Done()
		}(i)
	}
	waiter.Wait()
}

func waitWithTimeout() {
	limit := gocc.NewAtomicSemaphore(300)
	waiter := sync.WaitGroup{}
	waiter.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(id int) {
			start := time.Now().UnixNano()
			for {
				if limit.AcquireTimeout(time.Millisecond * 1000) {
					break
				}
				log.Printf("g:%d,timeout...\n", id)
			}

			waitTime := time.Now().UnixNano() - start
			time.Sleep(time.Millisecond * 200)
			log.Printf("g:%d,wait:%d\n", id, waitTime)
			limit.Release()
			waiter.Done()
		}(i)
	}
	waiter.Wait()
}
