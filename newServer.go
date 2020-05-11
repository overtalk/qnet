package qnet

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet"
)

type (
	OnInitCompleteFunc func(server interface{}) Action
	OnShutdownFunc     func(server interface{})
	OnOpenedFunc       func(c Conn) ([]byte, Action)
	OnClosedFunc       func(c Conn, err error) Action
	PreWriteFunc       func()
	ReactFunc          func(frame []byte, c Conn) (out []byte, action Action)
	TickFunc           func() (time.Duration, Action)
)

type NServer struct {
	sync.RWMutex
	baseID      uint64
	ep          *Endpoint
	connections map[uint64]Conn
	// msg router
	msgRouterSwitch bool
	codec           INetMsgCodec
	// logic handler
	onInitComplete OnInitCompleteFunc
	onShutdown     OnShutdownFunc
	onOpened       OnOpenedFunc
	onClosed       OnClosedFunc
	preWrite       PreWriteFunc
	react          ReactFunc
	tick           TickFunc
}

func NewNServer() *NServer {
	ep, _ := NewFromString("udp://localhost:9999")
	return &NServer{
		ep:             ep,
		onInitComplete: func(server interface{}) gnet.Action { return gnet.None },
		onShutdown:     func(server interface{}) {},
		onOpened:       func(c gnet.Conn) ([]byte, gnet.Action) { return nil, gnet.None },
		onClosed:       func(c gnet.Conn, err error) gnet.Action { return gnet.None },
		preWrite:       func() {},
	}
}

func (svr *NServer) Start() { gnet.Serve(svr, svr.ep.ToString()) }

// for gnet
func (svr *NServer) OnInitComplete(server gnet.Server) Action { return svr.onInitComplete(server) }
func (svr *NServer) OnShutdown(server interface{})            { svr.onShutdown(server) }
func (svr *NServer) OnOpened(c Conn) ([]byte, Action) {
	// add connection
	c.SetContext(atomic.AddUint64(&svr.baseID, 1))
	svr.Lock()
	svr.connections[svr.baseID] = c
	svr.Unlock()

	return svr.onOpened(c)
}
func (svr *NServer) OnClosed(c Conn, err error) Action {
	// delete connections
	id := c.Context().(uint64)
	svr.Lock()
	delete(svr.connections, id)
	svr.Unlock()

	return svr.onClosed(c, err)
}
func (svr *NServer) PreWrite()                                  { svr.preWrite() }
func (svr *NServer) Tick() (delay time.Duration, action Action) { return svr.tick() }
func (svr *NServer) React(frame []byte, c Conn) ([]byte, Action) {
	if svr.msgRouterSwitch {
		// todo: use msg router
		return svr.codec.React(frame, c)
	}
	return svr.react(frame, c)
}

// chained-mode set func
func (svr *NServer) SetOnInitCompleteFunc(f OnInitCompleteFunc) *NServer {
	svr.onInitComplete = f
	return svr
}

func (svr *NServer) SetOnShutdownFunc(f OnShutdownFunc) *NServer {
	svr.onShutdown = f
	return svr
}

func (svr *NServer) SetOnOpenedFunc(f OnOpenedFunc) *NServer {
	svr.onOpened = f
	return svr
}

func (svr *NServer) SetOnClosedFunc(f OnClosedFunc) *NServer {
	svr.onClosed = f
	return svr
}

func (svr *NServer) SetPreWriteFunc(f PreWriteFunc) *NServer {
	svr.preWrite = f
	return svr
}

func (svr *NServer) SetReactFunc(f ReactFunc) *NServer {
	svr.react = f
	return svr
}

func (svr *NServer) SetTickFunc(f TickFunc) *NServer {
	svr.tick = f
	return svr
}
