package qnet

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet"
)

type NServer struct {
	eventHandler
	rwMutex     sync.RWMutex
	baseID      uint64
	ep          *Endpoint
	connections map[uint64]Conn
	// msg router
	msgRouterSwitch bool
	msgRouter       INetMsgRouter
	// ws server
	wsServer *webSocketServer

	options *gnet.Options
}

func NewNServer() *NServer {
	return &NServer{
		rwMutex:         sync.RWMutex{},
		msgRouterSwitch: false,
		connections:     make(map[uint64]Conn),
		eventHandler:    newEventHandler(),
	}
}

func (svr *NServer) Start() {
	switch svr.ep.Proto() {
	case ProtoTypeWs:
		svr.wsServer = newWebSocketServer(svr)
		svr.wsServer.Start()
	case ProtoTypeTcp, ProtoTypeUdp:
		if svr.options != nil {
			gnet.Serve(svr, svr.ep.ToString(), gnet.WithOptions(*svr.options))
		} else {
			gnet.Serve(svr, svr.ep.ToString())
		}
	default:
		log.Fatal("invalid net protocol : ", svr.ep.Proto())
	}
}

// for gnet
func (svr *NServer) OnInitComplete(server gnet.Server) Action { return svr.onInitComplete(server) }
func (svr *NServer) OnShutdown(server interface{})            { svr.onShutdown(server) }
func (svr *NServer) OnOpened(c Conn) ([]byte, Action) {
	// add connection
	id := atomic.AddUint64(&svr.baseID, 1)
	c.SetContext(id)
	svr.rwMutex.Lock()
	svr.connections[id] = c
	svr.rwMutex.Unlock()

	return svr.onOpened(c)
}
func (svr *NServer) OnClosed(c Conn, err error) Action {
	// delete connections
	id := c.Context().(uint64)
	svr.rwMutex.Lock()
	delete(svr.connections, id)
	svr.rwMutex.Unlock()

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

func (svr *NServer) SetGNetOptions(option *gnet.Options) *NServer {
	svr.options = option
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
