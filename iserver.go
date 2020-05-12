package qnet

import (
	"time"

	"github.com/panjf2000/gnet"
)

type INetMsgRouter interface {
	RegisterMsgHandler(id uint16, handler Logic)
	DecodeNetMsg(data []byte) (*NetMsg, error)
	EncodeNetMsg(msg *NetMsg) []byte
	React(frame []byte, c Conn) (out []byte, action Action)
}

var (
	None     = gnet.None
	Close    = gnet.Close
	Shutdown = gnet.Shutdown
)

type (
	Action = gnet.Action
	Conn   = gnet.Conn

	OnInitCompleteFunc func(server interface{}) Action
	OnShutdownFunc     func(server interface{})
	OnOpenedFunc       func(c Conn) ([]byte, Action)
	OnClosedFunc       func(c Conn, err error) Action
	PreWriteFunc       func()
	ReactFunc          func(frame []byte, c Conn) (out []byte, action Action)
	TickFunc           func() (time.Duration, Action)

	Logic func(msg *NetMsg, c Conn) *NetMsg
)

type eventHandler struct {
	onInitComplete OnInitCompleteFunc
	onShutdown     OnShutdownFunc
	onOpened       OnOpenedFunc
	onClosed       OnClosedFunc
	preWrite       PreWriteFunc
	react          ReactFunc
	tick           TickFunc
}

func newEventHandler() eventHandler {
	return eventHandler{
		onInitComplete: func(server interface{}) gnet.Action { return gnet.None },
		onShutdown:     func(server interface{}) {},
		onOpened:       func(c gnet.Conn) ([]byte, gnet.Action) { return nil, gnet.None },
		onClosed:       func(c gnet.Conn, err error) gnet.Action { return gnet.None },
		preWrite:       func() {},
	}
}
