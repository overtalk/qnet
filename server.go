package qnet

import (
	"github.com/panjf2000/gnet"
	"log"
	"sync"
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

	gNetOption *gnet.Options
	gNetServer *gNetServer
	wsServer   *webSocketServer
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
		svr.gNetServer = newGNetServer(svr)
		svr.gNetServer.Start()
	default:
		log.Fatal("invalid net protocol : ", svr.ep.Proto())
	}
}

func (svr *NServer) React(frame []byte, c Conn) ([]byte, Action) {
	if svr.msgRouterSwitch {
		// todo: use msg router
		return svr.msgRouter.React(frame, c)
	}
	return svr.react(frame, c)
}

func (svr *NServer) CloseConn(id uint64) {
	svr.rwMutex.Lock()
	delete(svr.connections, id)
	svr.rwMutex.Unlock()
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
		svr.SetOption(gnet.WithTicker(true))
		svr.tick = handler.(TickFunc)
	default:
		log.Fatalf("invalid handler : %v\n", handler)
	}
	return svr
}
func (svr *NServer) SetOption(opt gnet.Option) *NServer {
	if svr.gNetOption == nil {
		svr.gNetOption = &gnet.Options{}
	}
	opt(svr.gNetOption)
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
