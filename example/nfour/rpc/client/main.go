package main

import (
	"fmt"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/rpc/proto"
	"github.com/rolandhe/saber/utils/sortutil"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	conf := nfour.NewTransConf(time.Second*2, 500)
	t, err := nfour.NewTrans("localhost:11011", conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := proto.NewJsonRpcClient(t)

	concurrentSend(100000, client)

	client.Shutdown()
}

type req struct {
	val    string
	sortId int
}

func concurrentSend(taskCount int, c proto.JsonClient) {
	reqs := buildRequests(taskCount)
	reqTimeout := &nfour.ReqTimeout{
		WaitConcurrent: time.Millisecond * 1000,
	}

	wg := sync.WaitGroup{}
	wg.Add(taskCount)

	lock := sync.Mutex{}

	var rarray []string

	for _, req := range reqs {
		go func(r *proto.JsonProtoReq) {
			s := ""
			tryCount := 0
			for {
				resp, err := c.SendRequest(r, reqTimeout)
				if err != nil {
					s = "err:" + err.Error() + "###" + err.Error()
					if tryCount < 2 {
						time.Sleep(time.Millisecond * 10)
						tryCount++
					} else {
						break
					}
				} else {
					s = string(resp.Body)
					break
				}
			}

			lock.Lock()
			rarray = append(rarray, s)
			lock.Unlock()
			wg.Done()
		}(req)

	}
	wg.Wait()
	fmt.Println()
	fmt.Println("--------------------------------")
	nArrays := convertBatchResult(rarray)
	errCount := 0
	for _, v := range nArrays {
		if strings.HasPrefix(v.val, "err:") {
			errCount++
		} else {
			fmt.Println(v.val)
		}
	}
	fmt.Printf("......end(%d)..error(%d)..\n", len(rarray), errCount)
}

func convertBatchResult(res []string) []*req {
	var array []*req
	for _, s := range res {
		pos := strings.LastIndex(s, "-")
		ids := s[pos+1:]
		id, _ := strconv.Atoi(ids)
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
		v := "hello world-" + strconv.Itoa(i)
		ret = append(ret, &proto.JsonProtoReq{"rpc.test", []byte(v)})
	}
	return ret
}
