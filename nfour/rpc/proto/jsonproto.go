package proto

import (
	"encoding/json"
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/rpc"
)

func NewJsonRpcSrvWorking(errToRes rpc.HandleErrorFunc[JsonProtoRes]) (nfour.Working, *rpc.SrvRouter[JsonProtoReq, JsonProtoRes]) {
	codec := &jsonSerCodec[JsonProtoReq, JsonProtoRes]{}
	kExtractor := func(req *JsonProtoReq) any {
		return req.Key
	}
	return rpc.NewRpcWorking[JsonProtoReq, JsonProtoRes](codec, kExtractor, errToRes)
}

type JsonClient interface {
	SendRequest(req *JsonProtoReq, reqTimeout *nfour.ReqTimeout) (*JsonProtoRes, error)
	Shutdown()
}

func NewJsonRpcClient(trans *nfour.Trans) JsonClient {
	codec := &jsonClientCodec[JsonProtoReq, JsonProtoRes]{}
	c := rpc.NewClient[JsonProtoReq, JsonProtoRes](codec, trans)
	return c
}

type JsonProtoReq struct {
	Key  string `json:"key"`
	Body []byte `json:"body"`
}

type JsonProtoRes JsonProtoReq

type jsonClientCodec[REQ JsonProtoReq, RES JsonProtoRes] struct {
}

func (jc *jsonClientCodec[JsonProto, JsonProtoRes]) Decode(payload []byte) (*JsonProtoRes, error) {
	o := new(JsonProtoRes)
	if err := json.Unmarshal(payload, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (jc *jsonClientCodec[JsonProto, JsonProtoRes]) Encode(req *JsonProto) ([]byte, error) {
	return json.Marshal(req)
}

type jsonSerCodec[REQ JsonProtoReq, RES JsonProtoRes] struct {
}

func (jc *jsonSerCodec[JsonProto, JsonProtoRes]) Decode(payload []byte) (*JsonProto, error) {
	o := new(JsonProto)
	if err := json.Unmarshal(payload, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (jc *jsonSerCodec[JsonProto, JsonProtoRes]) Encode(req *JsonProtoRes) ([]byte, error) {
	return json.Marshal(req)
}
