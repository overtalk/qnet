package gnet

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet"
)

type (
	GNetConn   = gnet.Conn
	GNetServer = gnet.Server
	GNetAction = gnet.Action

	OnInitCompleteFunc func(server GNetServer) GNetAction
	OnShutdownFunc     func(server GNetServer)
	OnOpenedFunc       func(c GNetConn) ([]byte, GNetAction)
	OnClosedFunc       func(c GNetConn, err error) GNetAction
	PreWriteFunc       func()
	ReactFunc          func(frame []byte, c GNetConn) (out []byte, action GNetAction)
	TickFunc           func() (time.Duration, GNetAction)
)

var (
	NoneAction     = gnet.None
	CloseAction    = gnet.Close
	ShutdownAction = gnet.Shutdown
)

type QNetServer struct {
	sync.RWMutex
	baseID      uint64
	netMsgCodec INetMsgCodec
	// session manager
	sessions map[uint64]GNetConn
	// some func
	onInitComplete OnInitCompleteFunc
	onShutdown     OnShutdownFunc
	onOpened       OnOpenedFunc
	onClosed       OnClosedFunc
	preWrite       PreWriteFunc
	react          ReactFunc
	tick           TickFunc
}

func NewQNetServer(options ...Option) (*QNetServer, error) {
	svr := &QNetServer{
		sessions:       make(map[uint64]GNetConn),
		onInitComplete: func(server GNetServer) GNetAction { return NoneAction },
		onShutdown:     func(server GNetServer) {},
		onOpened:       func(c GNetConn) ([]byte, GNetAction) { return nil, NoneAction },
		onClosed:       func(c GNetConn, err error) GNetAction { return NoneAction },
		preWrite:       func() {},
	}

	for _, option := range options {
		if err := option(svr); err != nil {
			return nil, err
		}
	}

	return svr, nil
}

func (svr *QNetServer) Start(opt ...gnet.Option) {
	log.Fatal(gnet.Serve(svr, "tcp://:9999", opt...))
}

func (svr *QNetServer) RegisterMsgHandler(id uint16, handler Logic) {
	svr.netMsgCodec.RegisterMsgHandler(id, handler)
}

// OnInitComplete fires when the server is ready for accepting connections.
// The server parameter has information and various utilities.
func (svr *QNetServer) OnInitComplete(server GNetServer) GNetAction {
	return svr.onInitComplete(server)
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (svr *QNetServer) OnShutdown(server GNetServer) { svr.onShutdown(server) }

// OnOpened fires when a new connection has been opened.
// The info parameter has information about the connection such as
// it's local and remote address.
// Use the out return value to write data to the connection.
func (svr *QNetServer) OnOpened(c GNetConn) ([]byte, GNetAction) {
	c.SetContext(atomic.AddUint64(&svr.baseID, 1))
	svr.Lock()
	svr.sessions[svr.baseID] = c
	svr.Unlock()

	return svr.onOpened(c)
}

// OnClosed fires when a connection has been closed.
// The err parameter is the last known connection error.
func (svr *QNetServer) OnClosed(c GNetConn, err error) GNetAction {
	id := c.Context().(uint64)
	svr.Lock()
	delete(svr.sessions, id)
	svr.Unlock()
	return svr.onClosed(c, err)
}

// PreWrite fires just before any data is written to any client socket.
func (svr *QNetServer) PreWrite() { svr.preWrite() }

// React fires when a connection sends the server data.
// Invoke c.Read() or c.ReadN(n) within the parameter c to read incoming data from client/connection.
// Use the out return value to write data to the client/connection.
func (svr *QNetServer) React(frame []byte, c GNetConn) (out []byte, action GNetAction) {
	if frame == nil {
		// if frame is nil, this is called by yourself
		return nil, CloseAction
	}

	return svr.react(frame, c)
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (svr *QNetServer) Tick() (time.Duration, GNetAction) { return svr.tick() }
