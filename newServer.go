package qnet

import (
	"log"
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
	msgRouter       INetMsgRouter

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
	return &NServer{
		msgRouterSwitch: false,
		connections:     make(map[uint64]Conn),
		onInitComplete:  func(server interface{}) gnet.Action { return gnet.None },
		onShutdown:      func(server interface{}) {},
		onOpened:        func(c gnet.Conn) ([]byte, gnet.Action) { return nil, gnet.None },
		onClosed:        func(c gnet.Conn, err error) gnet.Action { return gnet.None },
		preWrite:        func() {},
	}
}

func (svr *NServer) Start() { gnet.Serve(svr, svr.ep.ToString()) }

// chained-mode set func
func (svr *NServer) SetOnInitCompleteFunc(f OnInitCompleteFunc) *NServer { return svr.setLogicFunc(f) }
func (svr *NServer) SetOnShutdownFunc(f OnShutdownFunc) *NServer         { return svr.setLogicFunc(f) }
func (svr *NServer) SetOnOpenedFunc(f OnOpenedFunc) *NServer             { return svr.setLogicFunc(f) }
func (svr *NServer) SetOnClosedFunc(f OnClosedFunc) *NServer             { return svr.setLogicFunc(f) }
func (svr *NServer) SetPreWriteFunc(f PreWriteFunc) *NServer             { return svr.setLogicFunc(f) }
func (svr *NServer) SetReactFunc(f ReactFunc) *NServer                   { return svr.setLogicFunc(f) }
func (svr *NServer) SetTickFunc(f TickFunc) *NServer                     { return svr.setLogicFunc(f) }
func (svr *NServer) setLogicFunc(handler interface{}) *NServer {
	switch handler.(type) {
	case OnInitCompleteFunc:
		svr.onInitComplete = handler.(OnInitCompleteFunc)
	case OnShutdownFunc:
		svr.onShutdown = handler.(OnShutdownFunc)
	case OnOpenedFunc:
		svr.onOpened = handler.(OnOpenedFunc)
	case OnClosedFunc:
		svr.onClosed = handler.(OnClosedFunc)
	case PreWriteFunc:
		svr.preWrite = handler.(PreWriteFunc)
	case ReactFunc:
		svr.react = handler.(ReactFunc)
	case TickFunc:
		svr.tick = handler.(TickFunc)
	default:
		log.Fatalf("invalid handler : %v\n", handler)
	}
	return svr
}

func (svr *NServer) SetMsgRouter(router INetMsgRouter) *NServer {
	svr.msgRouterSwitch = true
	svr.msgRouter = router
	return svr
}

func (svr *NServer) SetURL(url string) *NServer {
	ep, err := NewFromString(url)
	if err != nil {
		log.Fatal(err)
	}
	svr.ep = ep
	return svr
}

func (svr *NServer) RegisterMsgHandler(id uint16, handler Logic) *NServer {
	if svr.msgRouter == nil {
		log.Fatal("the codec for server is absent")
	}

	svr.msgRouter.RegisterMsgHandler(id, handler)
	return svr
}

// for gnet
func (svr *NServer) OnInitComplete(server gnet.Server) Action { return svr.onInitComplete(server) }
func (svr *NServer) OnShutdown(server interface{})            { svr.onShutdown(server) }
func (svr *NServer) OnOpened(c Conn) ([]byte, Action) {
	// add connection
	id := atomic.AddUint64(&svr.baseID, 1)
	c.SetContext(id)
	svr.Lock()
	svr.connections[id] = c
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
		return svr.msgRouter.React(frame, c)
	}
	return svr.react(frame, c)
}
