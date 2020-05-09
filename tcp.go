package qnet

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type tcp struct {
	svr      *Server
	ep       *Endpoint        // endpoint
	listener *net.TCPListener // for tcp
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func newTcp(ep *Endpoint, svr *Server) (*tcp, error) {
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

			// gen session id
			baseSessionID++

			// handle connection
			go func(svr *Server, sessionID uint64, conn *net.TCPConn) {
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

// --------------------------------------------------
//type UdpSession TcpSession
type TcpSession struct {
	BasicSession
	conn net.Conn
}

func NewTcpSession(sessionID uint64, conn net.Conn) *TcpSession {
	return &TcpSession{
		BasicSession: *NewBasicSession(sessionID),
		conn:         conn,
	}
}

func (ts *TcpSession) Close() error                        { return ts.conn.Close() }
func (ts *TcpSession) Write(p []byte) (n int, err error)   { return ts.TcpWrite(p) }
func (ts *TcpSession) Read(p []byte) (n int, err error)    { return ts.TcpRead(p) }
func (ts *TcpSession) TcpWrite(data []byte) (int, error)   { return ts.conn.Write(data) }
func (ts *TcpSession) TcpRead(p []byte) (n int, err error) { return ts.conn.Read(p) }

// for msg router
// second flag in return means exit
func (ts *TcpSession) GetNetMsg(length HeadLength, decoderFunc HeadDeserializeFunc) (*NetMsg, *net.UDPAddr, error) {
	// decode head
	headerBytes := make([]byte, length)
	if _, err := io.ReadFull(ts, headerBytes); err != nil {
		return nil, nil, err
	}

	head, err := decoderFunc(headerBytes)
	if err != nil {
		return nil, nil, err
	}

	bodyByte := make([]byte, head.GetMsgLength())
	if _, err := io.ReadFull(ts, bodyByte); err != nil {
		return nil, nil, err
	}

	return NewNetMsg(head, bodyByte), nil, nil
}

func (ts *TcpSession) SendNetMsg(headSerializeFunc HeadSerializeFunc, msg *NetMsg, _ *net.UDPAddr) error {
	bytes := headSerializeFunc(msg)
	bytes = append(bytes, msg.GetMsg()...)
	_, err := ts.Write(bytes)
	return err
}
