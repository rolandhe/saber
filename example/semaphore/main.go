package main

import (
	"fmt"
	"github.com/rolandhe/saber/gocc"
	"log"
	"sync"
	"time"
)

func main() {
	waitUntilChan()
	start()
}

func start() {
	ch := make(chan struct{})
	v := 1
	go func() {
		fmt.Println(v)
		ch <- struct{}{}
	}()
	v = 12
	<-ch
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
