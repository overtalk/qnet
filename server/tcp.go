package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/overtalk/qnet/base"
)

type tcp struct {
	svr      *server
	ep       *base.Endpoint   // endpoint
	listener *net.TCPListener // for tcp
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func newTcp(ep *base.Endpoint, svr *server) (*tcp, error) {
	addr, err := ep.TCPAddr()
	if err != nil {
		return nil, err
	}

	ln, err := net.ListenTCP(string(ep.Proto()), addr)
	if err != nil {
		return nil, err
	}

	return &tcp{
		ep:       ep,
		svr:      svr,
		listener: ln,
		stopChan: make(chan interface{}),
	}, nil
}

func (t *tcp) Start() error {
	if t.listener == nil {
		return errors.New("tcp listener is nil")
	}

	go func(t *tcp) {
		<-t.stopChan
		// TODO: change log
		fmt.Println("Stop Tcp Server ...")
		if err := t.listener.Close(); err != nil {
			// TODO: change log
			fmt.Println(err)
		}
	}(t)

	t.stopFlag = false

	go func() {
		var baseSessionID uint64 = 0
		for {
			conn, err := t.listener.AcceptTCP()
			if err != nil {
				if t.stopFlag {
					// TODO: change log
					fmt.Println("stop listen :", err.Error())
					break
				}
				// TODO: change log
				fmt.Println("failed to accept connection :", err.Error())
				continue
			}

			baseSessionID++

			// default is true
			//conn.SetNoDelay(true)

			// handle connection
			go func(svr *server, sessionID uint64, conn *net.TCPConn) {
				session := NewTcpSession(sessionID, conn)

				// do some hook
				for _, connectHook := range svr.connectHookList {
					connectHook(session)
				}

				// do logic
				svr.handler(session)

				// do some hook
				for _, connectHook := range svr.disconnectHookList {
					connectHook(session)
				}
			}(t.svr, baseSessionID, conn)
		}
	}()

	return nil
}

func (t *tcp) Stop() {
	t.stopFlag = true
	t.stopChan <- struct{}{}
}