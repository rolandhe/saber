package main

import (
	"fmt"
	"github.com/rolandhe/saber/example/nfour/rpc/server/handler"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/duplex"
	"github.com/rolandhe/saber/nfour/rpc/proto"
	"github.com/rolandhe/saber/utils/sortutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	//wait := &sync.WaitGroup{}
	//wait.Add(2)
	//start := time.Now().UnixNano()
	//go func() {
	//	core("a")
	//	wait.Done()
	//}()
	//go func() {
	//	core("b")
	//	wait.Done()
	//}()
	//
	//wait.Wait()
	//nfour.NFourLogger.InfoLn("xxxx---", time.Now().UnixNano()-start)

	core("single")
}

func core(name string) {
	conf := duplex.NewTransConf(time.Second*2, 5000)
	t, err := duplex.NewTrans("localhost:11011", conf, name)
	if err != nil {
		log.Println(err)
		return
	}

	client := proto.NewJsonRpcClient(t)

	concurrentSend(50000, client)

	client.Shutdown(name + "-main")

}

type req struct {
	val    string
	sortId int
}

func concurrentSend(taskCount int, c proto.JsonClient) {
	reqs := buildRequests(taskCount)
	reqTimeout := &duplex.ReqTimeout{
		WaitConcurrent: time.Millisecond * 1000,
	}

	wg := sync.WaitGroup{}
	wg.Add(taskCount)

	lock := sync.Mutex{}

	var rarray []string

	start := time.Now().UnixNano()
	trigger := sync.WaitGroup{}
	trigger.Add(1)
	for _, req := range reqs {
		go func(r *proto.JsonProtoReq) {
			trigger.Wait()
			s := ""
			tryCount := 0
			for {
				resp, err := c.SendRequest(r, reqTimeout)
				if err != nil {
					s = "err:" + "###" + err.Error()
					if tryCount < 2 {
						time.Sleep(time.Millisecond * 10)
						tryCount++
					} else {
						break
					}
				} else {
					jsonResult, _ := proto.ParseJsonProtoRes[handler.Result[string]](resp, func() *handler.Result[string] {
						return &handler.Result[string]{}
					})
					if jsonResult.Code == 500 {
						s = "err:" + "###" + jsonResult.Message
					} else {
						s = jsonResult.Data
					}
					break
				}
			}

			lock.Lock()
			rarray = append(rarray, s)
			lock.Unlock()
			wg.Done()
		}(req)

	}
	trigger.Done()

	wg.Wait()

	cost := time.Now().UnixNano() - start

	nfour.NFourLogger.InfoLn("--------------------------------")
	nArrays := convertBatchResult(rarray)
	errCount := 0
	last := -1
	lostCount := 0
	for _, v := range nArrays {
		if strings.HasPrefix(v.val, "err:") {
			nfour.NFourLogger.InfoLn(v.val)
			errCount++
		} else {
			//fmt.Println(v.val)
			if (v.sortId - last) != 1 {
				lostCount++
			}
			last = v.sortId
		}
	}
	nfour.NFourLogger.Info("......end(%d)..error(%d)..lost(%d)..\n", len(rarray), errCount, lostCount)
	nfour.NFourLogger.InfoLn(cost, "nano")
}

func convertBatchResult(res []string) []*req {
	var array []*req
	for _, s := range res {
		pos := strings.LastIndex(s, "-")
		ids := s[pos+1:]
		id, _ := strconv.Atoi(strings.TrimLeft(ids, "0"))
		array = append(array, &req{s, id})
	}

	sortutil.Cmp[*req](func(p1, p2 **req) bool {
		return (*p1).sortId < (*p2).sortId
	}).Sort(array)

	return array
}

func buildRequests(num int) []*proto.JsonProtoReq {
	var ret []*proto.JsonProtoReq
	for i := 0; i < num; i++ {
		num := fmt.Sprintf("%08d", i)
		v := "hello worldjjjjjjjjjjjjjjkadsjfkdjlasfjkldklsafjkdsafjkldsajlfjkdsajkfdjksafjkldjkslafjkldsajkfjkldsajkfjkdlsajkfdjkasfjkdsajkfjkldsajklfdjksafjkdjkasfjkdasjkfjkldsajkfjkdlasjfkfdasjkfjdklasfjkdsaf ok-" + num
		ret = append(ret, &proto.JsonProtoReq{"rpc.test", []byte(v)})
	}
	return ret
}
