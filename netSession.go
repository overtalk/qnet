package qnet

import (
	"errors"
	"fmt"
	"net"
)

var ErrUnrealized = errors.New("unrealized function")

type Session interface {
	// common
	GetSessionID() uint64
	SetMeta(key string, value interface{})
	GetMeta(key string) (interface{}, error)
	Close() error

	// tcp
	TcpRead(p []byte) (n int, err error)
	TcpWrite(p []byte) (n int, err error)

	// udp
	UdpRead(b []byte) (int, *net.UDPAddr, error)
	UdpWrite(b []byte, addr *net.UDPAddr) (int, error)

	// ws
	WsRead() (messageType int, data []byte, err error)
	WsWrite(messageType int, data []byte) error

	// for msg router, all kind of protocol include
	GetNetMsg(length HeadLength, decoderFunc HeadDeserializeFunc) (*NetMsg, *net.UDPAddr, error)
	SendNetMsg(headSerializeFunc HeadSerializeFunc, msg *NetMsg, addr *net.UDPAddr) error
}

// --------------------------------------------------
type BasicSession struct {
	sessionID uint64
	metas     map[string]interface{}
}

func NewBasicSession(sessionID uint64) *BasicSession {
	return &BasicSession{
		sessionID: sessionID,
		metas:     make(map[string]interface{}),
	}
}

// common
func (bs *BasicSession) GetSessionID() uint64                  { return bs.sessionID }
func (bs *BasicSession) SetMeta(key string, value interface{}) { bs.metas[key] = value }
func (bs *BasicSession) GetMeta(key string) (interface{}, error) {
	value, flag := bs.metas[key]
	if !flag {
		return nil, fmt.Errorf("meta(key = %s) is absent for session(id = %d)", key, bs.sessionID)
	}

	return value, nil
}

// tcp
func (bs *BasicSession) TcpWrite(data []byte) (int, error)   { return 0, ErrUnrealized }
func (bs *BasicSession) TcpRead(p []byte) (n int, err error) { return 0, ErrUnrealized }

// ws
func (bs *BasicSession) WsRead() (t int, data []byte, err error)    { return 0, nil, ErrUnrealized }
func (bs *BasicSession) WsWrite(messageType int, data []byte) error { return ErrUnrealized }

// udp
func (bs *BasicSession) UdpRead(b []byte) (int, *net.UDPAddr, error)       { return 0, nil, ErrUnrealized }
func (bs *BasicSession) UdpWrite(b []byte, addr *net.UDPAddr) (int, error) { return 0, ErrUnrealized }

// for net message
func (bs *BasicSession) GetNetMsg(length HeadLength, decoderFunc HeadDeserializeFunc) (*NetMsg, *net.UDPAddr, error) {
	return nil, nil, ErrUnrealized
}
func (bs *BasicSession) SendNetMsg(headSerializeFunc HeadSerializeFunc, msg *NetMsg, addr *net.UDPAddr) error {
	return ErrUnrealized
}
