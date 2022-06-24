package http

import (
	"github.com/szdartou/cellnet"
	"github.com/szdartou/cellnet/proc"
)

func init() {

	proc.RegisterProcessor("http", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {
		// 如果http的peer有队列，依然会在队列中排队执行
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

}
