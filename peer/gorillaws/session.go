package gorillaws

import (
	"github.com/szdartou/cellnet"
	"github.com/szdartou/cellnet/peer"
	"github.com/szdartou/cellnet/util"
	"github.com/gorilla/websocket"
	"sync"
)

// Socket会话
type wsSession struct {
	peer.CoreContextSet
	peer.CoreSessionIdentify
	*peer.CoreProcBundle

	pInterface cellnet.Peer

	conn *websocket.Conn

	// 退出同步器
	exitSync sync.WaitGroup

	// 发送队列
	sendQueue *cellnet.Pipe

	cleanupGuard sync.Mutex

	endNotify func()
}

func (self *wsSession) Peer() cellnet.Peer {
	return self.pInterface
}

// 取原始连接
func (self *wsSession) Raw() interface{} {
	if self.conn == nil {
		return nil
	}

	return self.conn
}

func (self *wsSession) Close() {
	self.sendQueue.Add(nil)
}

// 发送封包
func (self *wsSession) Send(msg interface{}) {
	self.sendQueue.Add(msg)
}

func (self *wsSession) protectedReadMessage() (msg interface{}, err error) {

	defer func() {

		if err := recover(); err != nil {
			log.Errorf("IO read panic: %s", err)
			self.Close()
		}

	}()

	msg, err = self.ReadMessage(self)

	return
}

// 接收循环
func (self *wsSession) recvLoop() {

	var capturePanic bool

	if i, ok := self.Peer().(cellnet.PeerCaptureIOPanic); ok {
		capturePanic = i.CaptureIOPanic()
	}

	for self.conn != nil {

		var msg interface{}
		var err error

		if capturePanic {
			msg, err = self.protectedReadMessage()
		} else {
			msg, err = self.ReadMessage(self)
		}

		if err != nil {

			log.Debugln(err)

			if !util.IsEOFOrNetReadError(err) {
				log.Errorln("session closed:", err)
			}

			self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: &cellnet.SessionClosed{}})
			break
		}

		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: msg})
	}

	self.Close()

	// 通知完成
	self.exitSync.Done()
}

// 发送循环
func (self *wsSession) sendLoop() {

	var writeList []interface{}

	for {
		writeList = writeList[0:0]
		exit := self.sendQueue.Pick(&writeList)

		// 遍历要发送的数据
		for _, msg := range writeList {

			// TODO SendMsgEvent并不是很有意义
			self.SendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
		}

		if exit {
			break
		}
	}

	// 关闭连接
	if self.conn != nil {
		self.conn.Close()
		self.conn = nil
	}

	// 通知完成
	self.exitSync.Done()
}

// 启动会话的各种资源
func (self *wsSession) Start() {

	// 将会话添加到管理器
	self.Peer().(peer.SessionManager).Add(self)

	// 需要接收和发送线程同时完成时才算真正的完成
	self.exitSync.Add(2)

	go func() {
		// 等待2个任务结束
		self.exitSync.Wait()

		// 将会话从管理器移除
		self.Peer().(peer.SessionManager).Remove(self)

		if self.endNotify != nil {
			self.endNotify()
		}

	}()

	// 启动并发接收goroutine
	go self.recvLoop()

	// 启动并发发送goroutine
	go self.sendLoop()
}

func newSession(conn *websocket.Conn, p cellnet.Peer, endNotify func()) *wsSession {
	self := &wsSession{
		conn:       conn,
		endNotify:  endNotify,
		sendQueue:  cellnet.NewPipe(),
		pInterface: p,
		CoreProcBundle: p.(interface {
			GetBundle() *peer.CoreProcBundle
		}).GetBundle(),
	}

	return self
}
