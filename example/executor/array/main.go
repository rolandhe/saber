package main

import (
	"github.com/rolandhe/saber/gocc"
	"log"
	"time"
)

func main() {
	//execSingleTaskUsingChanQ()
	//execSingleTaskUsingArrayQ()
	//execGroupTaskUsingChanQ()
	execGroupTaskUsingArrayQ()
}

func execSingleTaskUsingChanQ() {
	q := gocc.NewDefaultBlockingQueue[*gocc.ExecTask](10)
	execSingleTask(q)
}

func execGroupTaskUsingChanQ() {
	q := gocc.NewDefaultBlockingQueue[*gocc.ExecTask](10)
	execGroupTask(q)
}

func execGroupTaskUsingArrayQ() {
	q := gocc.NewArrayBlockingQueueDefault[*gocc.ExecTask](10)
	execGroupTask(q)
}

func execSingleTaskUsingArrayQ() {
	q := gocc.NewArrayBlockingQueueDefault[*gocc.ExecTask](10)
	execSingleTask(q)
}

func execSingleTask(q gocc.BlockingQueue[*gocc.ExecTask]) {
	executor := gocc.NewBufferedExecutor(q, 16)
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

func execGroupTask(q gocc.BlockingQueue[*gocc.ExecTask]) {
	executor := gocc.NewBufferedExecutor(q, 16)
	fg := gocc.NewFutureGroup(100)
	for i := 0; i < 100; i++ {
		_, ok := executor.ExecuteInGroupTimeout(func() (any, error) {
			time.Sleep(time.Millisecond * 100)
			return 325, nil
		}, fg, time.Second*3)
		if !ok {
			log.Println("error")
			return
		}
	}

	fg.WaitUntil()
	futures, _ := fg.GetFutures()
	for _, f := range futures {
		v, _ := f.Get()
		log.Println(v)
	}

}
