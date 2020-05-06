package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

type Session interface {
	io.Reader
	io.Writer
	ReadPacket() (p []byte, err error)
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

func (bs *BaseSession) Write(data []byte) (int, error) {
	return 0, errors.New("write func is unrealized")
}

func (bs *BaseSession) Read(p []byte) (n int, err error) {
	return 0, errors.New("read func is unrealized")
}

func (bs *BaseSession) ReadPacket() (p []byte, err error) {
	return nil, errors.New("readMessage func is unrealized")
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

// ------------------------------------------
type WsSession struct {
	BaseSession
	conn *websocket.Conn
}

func NewWsSession(sessionID uint64, conn *websocket.Conn) *WsSession {
	return &WsSession{
		BaseSession: *NewBaseSession(sessionID),
		conn:        conn,
	}
}

func (ws *WsSession) Write(data []byte) (int, error) {
	return len(data), ws.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (ws *WsSession) ReadPacket() (p []byte, err error) {
	mt, message, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if mt != websocket.BinaryMessage {
		// todo: handle error
		return nil, fmt.Errorf("invalid websocket messageType : %d", mt)
	}

	return message, nil
}

func (ws *WsSession) Close() error {
	return ws.conn.Close()
}
