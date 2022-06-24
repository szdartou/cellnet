package main

import (
	"fmt"
	"github.com/szdartou/cellnet"
	"github.com/szdartou/cellnet/codec"
	"github.com/szdartou/cellnet/util"
	"reflect"
)

type TestEchoACK struct {
	Msg   string
	Value int32
}

func (self *TestEchoACK) String() string { return fmt.Sprintf("%+v", *self) }

// 将消息注册到系统
func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*TestEchoACK)(nil)).Elem(),
		ID:    int(util.StringHash("main.TestEchoACK")),
	})
}
