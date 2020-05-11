package ggnet

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet"

	"github.com/overtalk/qnet"
)

type (
	OnInitCompleteFunc func(server gnet.Server) gnet.Action
	OnShutdownFunc     func(server gnet.Server)
	OnOpenedFunc       func(c gnet.Conn) ([]byte, gnet.Action)
	OnClosedFunc       func(c gnet.Conn, err error) gnet.Action
	PreWriteFunc       func()
	ReactFunc          func(frame []byte, c gnet.Conn) (out []byte, action gnet.Action)
	TickFunc           func() (time.Duration, gnet.Action)
)

type QNetServer struct {
	sync.RWMutex
	baseID       uint64
	protocolType qnet.ProtoType
	ep           *qnet.Endpoint
	netMsgCodec  INetMsgCodec
	// session manager
	sessions map[uint64]gnet.Conn
	options  []gnet.Option
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
		sessions:       make(map[uint64]gnet.Conn),
		onInitComplete: func(server gnet.Server) gnet.Action { return gnet.None },
		onShutdown:     func(server gnet.Server) {},
		onOpened:       func(c gnet.Conn) ([]byte, gnet.Action) { return nil, gnet.None },
		onClosed:       func(c gnet.Conn, err error) gnet.Action { return gnet.None },
		preWrite:       func() {},
	}

	for _, option := range options {
		if err := option(svr); err != nil {
			return nil, err
		}
	}

	return svr, nil
}

func (svr *QNetServer) Start() {
	if svr.ep == nil {
		log.Fatal("absent endpoint for server")
	}

	fmt.Println(svr.ep.ToString())

	log.Fatal(gnet.Serve(svr, svr.ep.ToString(), svr.options...))

}

func (svr *QNetServer) RegisterMsgHandler(id uint16, handler Logic) {
	svr.netMsgCodec.RegisterMsgHandler(id, handler)
}

// OnInitComplete fires when the server is ready for accepting connections.
// The server parameter has information and various utilities.
func (svr *QNetServer) OnInitComplete(server gnet.Server) gnet.Action {
	return svr.onInitComplete(server)
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (svr *QNetServer) OnShutdown(server gnet.Server) { svr.onShutdown(server) }

// OnOpened fires when a new connection has been opened.
// The info parameter has information about the connection such as
// it's local and remote address.
// Use the out return value to write data to the connection.
func (svr *QNetServer) OnOpened(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(atomic.AddUint64(&svr.baseID, 1))
	svr.Lock()
	svr.sessions[svr.baseID] = c
	svr.Unlock()

	return svr.onOpened(c)
}

// OnClosed fires when a connection has been closed.
// The err parameter is the last known connection error.
func (svr *QNetServer) OnClosed(c gnet.Conn, err error) gnet.Action {
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
func (svr *QNetServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	if frame == nil {
		// if frame is nil, this is called by yourself
		return nil, gnet.Close
	}

	return svr.react(frame, c)
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (svr *QNetServer) Tick() (time.Duration, gnet.Action) { return svr.tick() }
