package tcp

import (
	"github.com/szdartou/cellnet"
	"github.com/szdartou/cellnet/msglog"
	"github.com/szdartou/cellnet/relay"
	"github.com/szdartou/cellnet/rpc"
)

// 带有RPC和relay功能
type MsgHooker struct {
}

func (self MsgHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	var handled bool
	var err error

	inputEvent, handled, err = rpc.ResolveInboundEvent(inputEvent)

	if err != nil {
		log.Errorln("rpc.ResolveInboundEvent:", err)
		return
	}

	if !handled {

		inputEvent, handled, err = relay.ResoleveInboundEvent(inputEvent)

		if err != nil {
			log.Errorln("relay.ResoleveInboundEvent:", err)
			return
		}

		if !handled {
			msglog.WriteRecvLogger(log, "tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}

func (self MsgHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	handled, err := rpc.ResolveOutboundEvent(inputEvent)

	if err != nil {
		log.Errorln("rpc.ResolveOutboundEvent:", err)
		return nil
	}

	if !handled {

		handled, err = relay.ResolveOutboundEvent(inputEvent)

		if err != nil {
			log.Errorln("relay.ResolveOutboundEvent:", err)
			return nil
		}

		if !handled {
			msglog.WriteSendLogger(log, "tcp", inputEvent.Session(), inputEvent.Message())
		}
	}

	return inputEvent
}
