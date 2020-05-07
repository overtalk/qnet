package base

import (
	"errors"
	"fmt"
	"io"
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

func (bs *BasicSession) GetSessionID() uint64 {
	return bs.sessionID
}

func (bs *BasicSession) SetMeta(key string, value interface{}) {
	bs.metas[key] = value
}

func (bs *BasicSession) GetMeta(key string) (interface{}, error) {
	value, flag := bs.metas[key]
	if !flag {
		return nil, fmt.Errorf("meta(key = %s) is absent for session(id = %d)", key, bs.sessionID)
	}

	return value, nil
}

func (bs *BasicSession) Write(data []byte) (int, error) {
	return 0, errors.New("write func is unrealized")
}

func (bs *BasicSession) Read(p []byte) (n int, err error) {
	return 0, errors.New("read func is unrealized")
}

func (bs *BasicSession) ReadPacket() (p []byte, err error) {
	return nil, errors.New("readMessage func is unrealized")
}
