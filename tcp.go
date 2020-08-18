package qnet

import (
	"net"
	"os"
	"time"
)

// HandlerFunc a Handler wrapper
type HandlerFunc func(net.Conn)

type TcpServer struct {
	name    string      // listener's name
	network string      // eg: unix/tcp, see net.Dial
	address string      // eg: socket/ip:port, see net.Dial
	chmod   os.FileMode // file mode for unix socket, default 0666
	maxConn int         // listener's maximum connection number
	// if ReadSynced is true, when Listener is closed, all its connections
	// will not read any data. But it may be removed as a default option
	// that it's always stop reading data.
	readSynced bool

	connHandler HandlerFunc
}

func NewService(name, addr string, handler HandlerFunc) *TcpServer {
	return &TcpServer{
		name:        name,
		network:     "tcp",
		address:     addr,
		maxConn:     0,
		readSynced:  false,
		connHandler: handler,
	}
}

// GetName get the service name
func (ts *TcpServer) GetName() string { return ts.name }

// NewListener create a service listener
func (ts *TcpServer) NewListener() (net.Listener, error) {
	if ts.network == "unix" {
		os.Remove(ts.address)
	}
	l, err := net.Listen(ts.network, ts.address)
	if err != nil {
		// handle error
		return nil, err
	}
	if ts.network == "unix" {
		chmod := ts.chmod
		if chmod == 0 {
			chmod = 0666
		}
		os.Chmod(ts.address, chmod)
	}
	return l, nil
}

// Serve service logic
func (ts *TcpServer) Serve(l net.Listener) {
	//TODO：add log
	var tempDelay time.Duration
	for {
		// wait for a network connection
		conn, err := l.Accept()
		if err != nil {
			// referenced from $GOROOT/src/net/http/server.go:Serve()
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			//TODO：add log
			break
		}
		tempDelay = 0
		// handle every client in its own goroutine
		go ts.connHandler(conn)
	}
}
