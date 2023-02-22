package main

import (
	"fmt"
	"github.com/rolandhe/saber/gocc"
	"log"
	"time"
)

func main() {
	productAndConsumerUsingArrayQ()
	consumerWaitArrayQ()
}

func consumerWaitArrayQ() {
	q := gocc.NewArrayBlockingQueueDefault[int64](10)
	waiter := gocc.NewCountdownLatch(1)
	go func() {
		for {
			elem, ok := q.PullTimeout(time.Second * 20)
			if ok {
				v := elem.GetValue()
				end := time.Now().UnixNano()
				log.Printf("cost %d\n", end-v)
				waiter.Down()
				break
			}
		}
	}()

	time.Sleep(time.Millisecond * 800)

	ok := q.TryOffer(time.Now().UnixNano())
	log.Printf("offer:%v\n", ok)

	waiter.Wait()
}

func productAndConsumerUsingArrayQ() {
	q := gocc.NewArrayBlockingQueueDefault[int](10)
	productAndConsumerUsingQueue(q)
}

func productAndConsumerUsingQueue(q gocc.BlockingQueue[int]) {
	waiter := gocc.NewCountdownLatch(1000)
	for i := 1; i <= 10; i++ {
		go func(id int) {
			for {
				elem, ok := q.PullTimeout(time.Millisecond * 100)
				if !ok {
					log.Printf("wait timeout g:%d\n", id)
					continue
				}
				v := elem.GetValue()
				time.Sleep(time.Millisecond * 100)
				log.Printf("end g:%d,using %d nano\n", id, v)
				if 0 == waiter.Down() {
					break
				}
			}
		}(i)
	}

	start := time.Now().UnixNano()
	for i := 1; i <= 1000; {
		if q.OfferTimeout(i, time.Millisecond*1000) {
			i++
			log.Printf("offer timeout g:%d\n", i)
			continue
		}
	}

	waiter.Wait()
	cost := time.Now().UnixNano() - start
	fmt.Printf("end, cost %d nano\n", cost)
}
