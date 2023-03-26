# nfour 是一个基于tcp的网络开发框架
因为它是运行在4层上的，所以命名问题nfour。它包含三层：
* 通信层，负责二进制数据的通信，支持多路复用模式和传统的单路模式，
  * 多路复用利用单个连接并发执行多个任务请求，请求发送和结果接收并发执行，可以利用较少的资源支撑大量并发，推荐使用这种模式。同时支持服务端和客户端。
  * 单路模式，即传统的request/response模式，在一个连接上同步的发送request，等接收到response后再发送下一个request。只实现了服务端。
* rpc层， 基于通信层的抽象层，它负责发送业务对象到服务端，响应数据也是业务对象，所谓业务对象，即struct，或者string、int等，是业务调用所看到的那一层。
* 协议层，它位于rpc和通信层，负责讲业务对象转换成二进制。nfour利用了依赖倒置的设计原则，在rpc层定义一组协议接口，可以由业务使用这自由扩展。
  * nfour实现了基于json转换协议的缺省实现。

一般业务使用只需要实现协议层即可，也可以使用开箱即可的json协议。

# 基于json协议的示例
## 服务端

```
    type Result[T any] struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Data    T      `json:"data"`
    }
    
    func JsonRpcErrHandler(err error, interfaceName any) *proto.JsonProtoRes {
        ret := &Result[string]{
            Code:    500,
            Message: err.Error(),
        }
    
        method := "can't parse request, so don't know"
        if interfaceName != nil {
            method = interfaceName.(string)
        }
    
        body, _ := json.Marshal(ret)
        return &proto.JsonProtoRes{
            Key:  method,
            Body: body,
        }
    }
    
    
    func RegisterAll(router *rpc.SrvRouter[proto.JsonProtoReq, proto.JsonProtoRes]) {
        router.Register("rpc.test", func(req *proto.JsonProtoReq) (*proto.JsonProtoRes, error) {
    
            rpcResult, _ := json.Marshal(&Result[string]{
                Code: 200,
                Data: string(req.Body),
            })
            res := &proto.JsonProtoRes{
                Key:  req.Key,
                Body: rpcResult,
            }
            return res, nil
        })
    }
    
    func main() {
        working, handlerErrFunc, router := proto.NewJsonRpcSrvWorking(handler.JsonRpcErrHandler)
        handler.RegisterAll(router)
        conf := nfour.NewSrvConf(working, handlerErrFunc, 10000)
    
        duplex.Startup(11011, conf)
    }
```


## 客户端示例

```
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
    
    func main() {
        core("single")
    }
```