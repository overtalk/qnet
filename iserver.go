package qnet

import (
	"github.com/panjf2000/gnet"
	"time"
)

// SessionHandler defines the func to handle a session/connection
type SessionHandler func(session Session)

type MsgHandler func(session Session, msg *NetMsg) *NetMsg

type IServer interface {
	Start() error
	Stop()
}

// ----------------------------------------
type (
	Action = gnet.Action
	Conn   = gnet.Conn
)

type IEventHandler interface {
	// OnInitComplete fires when the server is ready for accepting connections.
	// The server parameter has information and various utilities.
	OnInitComplete(server interface{}) (action Action)

	// OnShutdown fires when the server is being shut down, it is called right after
	// all event-loops and connections are closed.
	OnShutdown(server interface{})

	// OnOpened fires when a new connection has been opened.
	// The info parameter has information about the connection such as
	// it's local and remote address.
	// Use the out return value to write data to the connection.
	OnOpened(conn Conn) (out []byte, action Action)

	// OnClosed fires when a connection has been closed.
	// The err parameter is the last known connection error.
	OnClosed(conn Conn, err error) (action Action)

	// PreWrite fires just before any data is written to any client socket.
	PreWrite()

	// React fires when a connection sends the server data.
	// Invoke c.Read() or c.ReadN(n) within the parameter c to read incoming data from client/connection.
	// Use the out return value to write data to the client/connection.
	React(frame []byte, c Conn) (out []byte, action Action)

	// Tick fires immediately after the server starts and will fire again
	// following the duration specified by the delay return value.
	Tick() (delay time.Duration, action Action)
}
