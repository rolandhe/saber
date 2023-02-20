package main

import (
	"github.com/rolandhe/saber/gocc"
	"log"
	"time"
)

func main() {
	executor := gocc.NewBufferedExecutor(gocc.NewChanBlockingQueue[*gocc.ExecTask](10), 16)
	f, ok := executor.ExecuteTimeout(func() (any, error) {
		time.Sleep(time.Millisecond * 100)
		return 325, nil
	}, time.Millisecond*3)
	if !ok {
		log.Println("error")
		return
	}
	ret, err := f.GetUntil()
	if err != nil {
		log.Println(err)
		return
	}
	v := ret.(int)
	log.Println(v)
}
