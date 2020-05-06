package server

import (
	"fmt"
	"io"
	"net"
)

type Session interface {
	io.Reader
	io.Writer
	GetSessionID() uint64
	SetMeta(key string, value interface{})
	GetMeta(key string) (interface{}, error)
	Close() error
}

// --------------------------------------------------
type BaseSession struct {
	sessionID uint64
	metas     map[string]interface{}
}

func NewBaseSession(sessionID uint64) *BaseSession {
	return &BaseSession{
		sessionID: sessionID,
		metas:     make(map[string]interface{}),
	}
}

func (bs *BaseSession) GetSessionID() uint64 {
	return bs.sessionID
}

func (bs *BaseSession) SetMeta(key string, value interface{}) {
	bs.metas[key] = value
}

func (bs *BaseSession) GetMeta(key string) (interface{}, error) {
	value, flag := bs.metas[key]
	if !flag {
		return nil, fmt.Errorf("meta(key = %s) is absent for session(id = %d)", key, bs.sessionID)
	}

	return value, nil
}

// --------------------------------------------------
//type UdpSession TcpSession
type TcpSession struct {
	BaseSession
	conn net.Conn
}

func NewTcpSession(sessionID uint64, conn net.Conn) *TcpSession {
	return &TcpSession{
		BaseSession: *NewBaseSession(sessionID),
		conn:        conn,
	}
}

func (ts *TcpSession) Write(data []byte) (int, error) {
	return ts.conn.Write(data)
}

func (ts *TcpSession) Read(p []byte) (n int, err error) {
	return ts.conn.Read(p)
}

func (ts *TcpSession) Close() error {
	return ts.conn.Close()
}
