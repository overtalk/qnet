package iface

import "github.com/overtalk/qnet/base"

// Handler defines the func to handle net message
type Handler func(session base.Session)

type MsgHandler func(session base.Session, msg *base.NetMsg) *base.NetMsg

type IServer interface {
	Start() error
	Stop()

	//SendBySessionID(sessionID uint64, data []byte) (int, error)
	//SetSessionMeta(sessionID uint64, key string, value interface{}) error
	//GetSessionMeta(sessionID uint64, key string) (interface{}, error)
}
