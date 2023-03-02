package main

import (
	"fmt"
	"github.com/rolandhe/saber/nfour/duplex"
	randutil "github.com/rolandhe/saber/utils/rand"
	"github.com/rolandhe/saber/utils/sortutil"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	reqs := buildRequests(5)
	fmt.Println(reqs)

	conf := duplex.NewTransConf(time.Minute*2, 200)
	c, err := duplex.NewTrans("localhost:11011", conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	concurrentSend(1000, c)

	c.Shutdown()
}

func concurrentSend(concur int, c *duplex.Trans) {
	reqs := buildRequests(concur)
	reqTimeout := &duplex.ReqTimeout{
		WaitConcurrent: time.Millisecond * 1000,
	}

	wg := sync.WaitGroup{}
	wg.Add(concur)

	lock := sync.Mutex{}

	var rarray []string

	for _, req := range reqs {
		go func(v string) {
			resp, err := c.SendPayload([]byte(v), reqTimeout)

			s := ""
			if err != nil {
				s = err.Error() + "###" + err.Error()
			} else {
				s = string(resp)
			}
			lock.Lock()
			rarray = append(rarray, s)
			lock.Unlock()
			wg.Done()
		}(req.val)

	}
	wg.Wait()
	fmt.Println()
	fmt.Println("--------------------------------")
	nArrays := convertBatchResult(rarray)
	for _, v := range nArrays {
		fmt.Println(v.val)
	}
	fmt.Printf("......end(%d)....\n", len(rarray))
}

type req struct {
	val    string
	sortId int
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

func buildRequests(num int) []*req {
	var ret []*req
	for i := 0; i < num; i++ {
		v := "hello world-" + strconv.Itoa(i)
		sortId := int(randutil.FastRandN(10000))
		ret = append(ret, &req{v, sortId})
	}

	sortutil.Cmp[*req](func(p1, p2 **req) bool {
		return (*p1).sortId < (*p2).sortId
	}).Sort(ret)

	return ret
}

func send(c *duplex.Trans, req string) {
	resp, err := c.SendPayload([]byte(req), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(resp))
}
