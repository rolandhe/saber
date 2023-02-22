package main

import (
	"github.com/rolandhe/saber/gocc"
	"log"
	"time"
)

func main() {
	//execGroupTask()
	execSingleTask()
}

func execSingleTask() {
	executor := gocc.NewDefaultExecutor(10)
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

func execGroupTask() {
	executor := gocc.NewDefaultExecutor(10)
	fg := gocc.NewFutureGroup(100)
	for i := 0; i < 100; i++ {
		_, ok := executor.ExecuteInGroupTimeout(func() (any, error) {
			time.Sleep(time.Millisecond * 100)
			return 325, nil
		}, fg, time.Second*1)
		if !ok {
			log.Println("error")
			return
		}
	}

	fg.Wait()
	futures, _ := fg.GetFutures()
	for _, f := range futures {
		v, _ := f.Get()
		log.Println(v)
	}
}
