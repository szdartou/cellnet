package rpc

import (
	"errors"
	"github.com/szdartou/cellnet"
	"github.com/szdartou/cellnet/codec"
	"sync"
	"sync/atomic"
)

var (
	rpcIDSeq        int64
	requestByCallID sync.Map
)

type Request struct {
	id     int64
	onRecv func(interface{})
}

var ErrTimeout = errors.New("RPC time out")

func (self *Request) RecvFeedback(msg interface{}) {

	// 异步和同步执行复杂，队列处理在具体的逻辑中手动处理
	self.onRecv(msg)
}

func (self *Request) GetId() int64 {
	return self.id
}

func (self *Request) Send(ses cellnet.Session, msg interface{}) {

	//ctx, _ := ses.(cellnet.ContextSet)

	data, meta, err := codec.EncodeMessage(msg, nil)

	if err != nil {
		log.Errorf("rpc request message encode error: %s", err)
		return
	}

	ses.Send(&RemoteCallREQ{
		MsgID:  uint32(meta.ID),
		Data:   data,
		CallID: self.id,
	})

	//codec.FreeCodecResource(meta.Codec, data, ctx)
}

func (self *Request) SendByte(ses cellnet.Session, msg []byte) {

	//ctx, _ := ses.(cellnet.ContextSet)

	ses.Send(&RemoteCallREQ{
		Data:   msg,
		CallID: self.id,
	})

	//codec.FreeCodecResource(meta.Codec, data, ctx)
}

func CreateRequest(onRecv func(interface{})) *Request {

	self := &Request{
		onRecv: onRecv,
	}

	self.id = atomic.AddInt64(&rpcIDSeq, 1)

	requestByCallID.Store(self.id, self)

	return self
}

func GetRequest(callid int64) *Request {

	if v, ok := requestByCallID.Load(callid); ok {

		requestByCallID.Delete(callid)
		return v.(*Request)
	}

	return nil
}
